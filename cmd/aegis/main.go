// aegis는 사건 목록을 조회하는 웹 서비스입니다.
//
// 사건 데이터는 Clio가 메모리에서 관리하고, 기관 데이터는 Leviathan이 제공합니다.
// 화면은 저장소 최상위 web/의 React(JSX) 앱이며, Vite로 빌드한 web/dist를 실행 파일에 포함합니다.
//
// 실행:
//
//	(cd web && npm install && npm run build)
//	go run ./cmd/leviathan
//	go run ./cmd/clio
//	go run ./cmd/aegis                    # http://localhost:8080
//	go run ./cmd/aegis -addr :9000
//
// 화면 개발 중에는 이 서버를 켜 둔 채 저장소 최상위 web/에서 npm run dev를 실행하면
// Vite 개발 서버(http://localhost:5173)가 /api 요청을 이 서버로 넘깁니다.
package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	leviathanv1 "github.com/mkkim-gr8/casepulse/gen/go/leviathan/v1"
	"github.com/mkkim-gr8/casepulse/internal/cliogrpc"
	"github.com/mkkim-gr8/casepulse/internal/leviathangrpc"
	"github.com/mkkim-gr8/casepulse/web"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// notBuiltPage는 화면을 아직 빌드하지 않았을 때 / 에서 보여 줄 안내입니다.
const notBuiltPage = `<!doctype html><meta charset="utf-8"><title>CasePulse</title>
<p>화면이 아직 빌드되지 않았습니다. <code>cd web &amp;&amp; npm install &amp;&amp; npm run build</code> 후 서버를 다시 시작하세요.</p>
<p>API는 <a href="/api/cases">/api/cases</a>에서 동작합니다.</p>`

// newHandler는 HTTP API와 빌드된 React 화면을 한 Gin 서버에 등록합니다.
func newHandler(courts []*leviathanv1.Court, clioClient *cliogrpc.Client) *gin.Engine {
	router := gin.Default()
	router.HandleMethodNotAllowed = true
	router.RedirectTrailingSlash = false
	newAPIHandlers(clioClient, courts).routes(router)

	dist, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		panic(err)
	}
	if _, err := fs.Stat(dist, "index.html"); err != nil {
		router.NoRoute(func(c *gin.Context) {
			if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
				c.Data(http.StatusServiceUnavailable, "text/html; charset=utf-8", []byte(notBuiltPage))
				return
			}
			c.Status(http.StatusNotFound)
		})
	} else {
		fileServer := http.FileServerFS(dist)
		router.NoRoute(func(c *gin.Context) {
			if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
				c.Status(http.StatusNotFound)
				return
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
		})
	}
	return router
}

// main은 HTTP와 두 gRPC 서비스의 주소를 옵션 또는 환경변수에서 읽습니다.
func main() {
	addr := flag.String("addr", ":8080", "listen address")
	leviathanAddr := flag.String("leviathan-addr", envOrDefault("LEVIATHAN_GRPC_ADDR", "localhost:9090"), "Leviathan gRPC address")
	clioAddr := flag.String("clio-addr", envOrDefault("CLIO_GRPC_ADDR", "localhost:9091"), "Clio gRPC address")
	flag.Parse()

	if err := run(*addr, *leviathanAddr, *clioAddr); err != nil {
		log.Fatal(err)
	}
}

// run은 gRPC 연결을 확인한 다음 HTTP 서버를 시작합니다.
func run(addr, leviathanAddr, clioAddr string) error {
	leviathanConn, err := grpc.NewClient(leviathanAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create Leviathan gRPC client: %w", err)
	}
	defer leviathanConn.Close()

	leviathanClient := leviathangrpc.NewClient(leviathanConn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	courts, err := leviathanClient.ListCourts(ctx, grpc.WaitForReady(true))
	if err != nil {
		return fmt.Errorf("load courts from Leviathan at %q: %w", leviathanAddr, err)
	}
	if len(courts) == 0 {
		return fmt.Errorf("Leviathan gRPC service at %q returned no courts", leviathanAddr)
	}

	clioConn, err := grpc.NewClient(clioAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create Clio gRPC client: %w", err)
	}
	defer clioConn.Close()
	clioClient := cliogrpc.NewClient(clioConn)
	if _, _, err := clioClient.ListCases(ctx, grpc.WaitForReady(true)); err != nil {
		return fmt.Errorf("connect to Clio at %q: %w", clioAddr, err)
	}

	return newHandler(courts, clioClient).Run(addr)
}

// envOrDefault는 환경변수가 비어 있을 때 사용할 기본 주소를 돌려줍니다.
func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
