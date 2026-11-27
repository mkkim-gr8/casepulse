package web

import "embed"

// Dist는 aegis가 제공하는 Vite 빌드 파일을 실행 파일에 포함합니다.
// 프런트엔드를 아직 빌드하지 않아도 포함 작업이 되도록 dist/.gitkeep을 유지합니다.
//
//go:embed all:dist
var Dist embed.FS
