// aegis는 사건 목록을 조회하는 웹 서비스입니다.
//
// 사건 데이터는 Clio가 관리하고(PostgreSQL 또는 메모리), 기관 데이터는 Leviathan이 제공합니다.
// Aegis 자신은 데이터를 저장하지 않는 "관문(gateway)" 역할입니다.
//
//	브라우저 ──HTTP/JSON──▶ Aegis(:8080) ──gRPC──▶ Clio(:9091)      사건 조회·등록
//	                                     └─gRPC──▶ Leviathan(:9090) 법원 목록(시작할 때 한 번)
//
// Gin은 Go의 웹 프레임워크로, URL 경로별로 처리 함수를 등록하는 일을 쉽게 해 줍니다.
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
	"github.com/mkkim-gr8/casepulse/internal/healthcheck"
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
//
// 테스트(main_test.go)에서도 이 함수를 불러 실제 서버를 켜지 않고 요청·응답을 확인합니다.
func newHandler(courts []*leviathanv1.Court, clioClient *cliogrpc.Client) *gin.Engine {
	router := gin.New()
	// Use로 등록한 함수(미들웨어)는 모든 요청에 먼저 실행됩니다.
	//   - Logger: 요청마다 한 줄씩 로그를 남깁니다(/healthz는 너무 자주 와서 제외).
	//   - Recovery: 처리 중 panic이 나도 서버가 죽지 않고 500 응답을 돌려주게 합니다.
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/healthz"}}), gin.Recovery())
	router.HandleMethodNotAllowed = true // 경로는 있는데 메서드가 틀리면 404 대신 405를 돌려줍니다.
	router.RedirectTrailingSlash = false // "/api/cases/"를 "/api/cases"로 자동 이동시키지 않습니다.
	newAPIHandlers(clioClient, courts).routes(router)
	// /healthz는 컨테이너 healthcheck용입니다. 서버는 gRPC 연결을 확인한 뒤에만 열리므로
	// 응답이 오면 요청을 받을 준비가 된 것입니다.
	healthz := func(c *gin.Context) { c.String(http.StatusOK, "ok") }
	router.GET("/healthz", healthz)
	router.HEAD("/healthz", healthz)

	// 위에서 등록하지 않은 모든 경로(NoRoute)는 빌드된 React 화면 파일(index.html, JS, CSS, 로고)로 응답합니다.
	// web.Dist는 실행 파일 안에 포함된 web/dist 폴더이고, fs.Sub로 그 안의 "dist" 폴더를 루트로 삼습니다.
	dist, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		panic(err)
	}
	// index.html이 없으면 아직 npm run build를 하지 않은 것이므로 안내 페이지를 보여 줍니다.
	if _, err := fs.Stat(dist, "index.html"); err != nil {
		router.NoRoute(func(c *gin.Context) {
			if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
				c.Data(http.StatusServiceUnavailable, "text/html; charset=utf-8", []byte(notBuiltPage))
				return
			}
			c.Status(http.StatusNotFound)
		})
	} else {
		// FileServerFS는 요청 경로에 맞는 파일을 찾아 보내 주는 표준 라이브러리 핸들러입니다.
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
	probe := flag.Bool("healthcheck", false, "check the running server at -addr and exit")
	flag.Parse()

	if *probe {
		if err := healthcheck.ProbeHTTP(*addr, "/healthz"); err != nil {
			log.Printf("healthcheck failed: %v", err)
			os.Exit(1)
		}
		return
	}
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
	// 시작 준비에 최대 30초를 허용합니다. WaitForReady(true)는 상대 서비스가 켜질 때까지 기다리게 합니다.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// 법원 목록은 시작할 때 한 번만 받아 보관합니다(등록 화면의 법원·재판부 선택지).
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
	// Clio가 요청을 받을 수 있는 상태인지 health RPC로 미리 확인한 뒤에 HTTP 서버를 엽니다.
	if err := healthcheck.CheckConn(ctx, clioConn, grpc.WaitForReady(true)); err != nil {
		return fmt.Errorf("connect to Clio at %q: %w", clioAddr, err)
	}

	// Run은 HTTP 서버를 열고 요청을 계속 처리하며, 서버가 멈출 때까지 돌아오지 않습니다.
	return newHandler(courts, clioClient).Run(addr)
}

// envOrDefault는 환경변수가 비어 있을 때 사용할 기본 주소를 돌려줍니다.
func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
