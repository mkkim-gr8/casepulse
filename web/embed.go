// Package web은 React 화면을 빌드한 결과(web/dist)를 Go 실행 파일 안에 넣어 두는 패키지입니다.
// 덕분에 Aegis 실행 파일 하나만 있으면 API와 화면을 함께 제공할 수 있습니다.
package web

import "embed"

// Dist는 aegis가 제공하는 Vite 빌드 파일을 실행 파일에 포함합니다.
// 프런트엔드를 아직 빌드하지 않아도 포함 작업이 되도록 dist/.gitkeep을 유지합니다.
//
//go:embed all:dist
var Dist embed.FS
