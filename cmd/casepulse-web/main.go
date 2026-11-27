// casepulse-web은 사건 목록을 조회하는 웹 서비스입니다.
//
// 아직 데이터베이스가 없으므로 시작할 때 더미 사건 목록(dummy.go)을 만들어 메모리에 둡니다.
// 화면은 web/의 React(JSX) + Blueprint.js 앱이며, Vite로 빌드한 web/dist를 실행 파일에 포함합니다.
//
// 실행:
//
//	(cd cmd/casepulse-web/web && npm install && npm run build)
//	go run ./cmd/casepulse-web            # http://localhost:8080
//	go run ./cmd/casepulse-web -addr :9000
//
// 화면 개발 중에는 이 서버를 켜 둔 채 web/에서 npm run dev를 실행하면
// Vite 개발 서버(http://localhost:5173)가 /api 요청을 이 서버로 넘깁니다.
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
)

// web/dist에는 빌드 전에도 go:embed가 실패하지 않도록 .gitkeep을 둡니다(all: 접두사로 포함).
//
//go:embed all:web/dist
var webDist embed.FS

// notBuiltPage는 화면을 아직 빌드하지 않았을 때 / 에서 보여 줄 안내입니다.
const notBuiltPage = `<!doctype html><meta charset="utf-8"><title>CasePulse</title>
<p>화면이 아직 빌드되지 않았습니다. <code>cd cmd/casepulse-web/web &amp;&amp; npm install &amp;&amp; npm run build</code> 후 서버를 다시 시작하세요.</p>
<p>API는 <a href="/api/cases">/api/cases</a>에서 동작합니다.</p>`

func newHandler() http.Handler {
	mux := http.NewServeMux()
	newCaseStore(buildDummyCases()).routes(mux)

	dist, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		panic(err)
	}
	if _, err := fs.Stat(dist, "index.html"); err != nil {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(notBuiltPage))
		})
	} else {
		mux.Handle("GET /", http.FileServerFS(dist))
	}
	return mux
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	log.Printf("casepulse-web listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, newHandler()))
}
