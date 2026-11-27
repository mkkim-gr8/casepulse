import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// npm run dev: Vite 개발 서버(http://localhost:5173). /api 요청은 Go 서버(:8080)로 넘깁니다.
// npm run build: dist/에 결과물을 만들고, Go 서버가 이 폴더를 실행 파일에 포함(go:embed)합니다.
//
// 화면은 shadcn/ui(Tailwind CSS v4) 컴포넌트로 만듭니다. "@/..."는 src/를 가리킵니다(jsconfig.json과 같게 유지).
// 저장소 맨 위의 casepulse-logo/ 폴더(로고·파비콘)를 그대로 사이트 루트에 내보냅니다(예: /favicon.ico).
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { "@": path.resolve(__dirname, "./src") },
  },
  publicDir: "../../../casepulse-logo",
  server: {
    proxy: { "/api": "http://localhost:8080" },
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
});
