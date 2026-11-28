import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import { ThemeProvider } from "@/components/theme-provider";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Toaster } from "@/components/ui/sonner";
import App from "./App.jsx";

// 화면의 시작점입니다. index.html의 <div id="root"> 안에 React 앱 전체를 그립니다.
//
// 바깥에서 안쪽으로 감싸는 순서:
//   - StrictMode: 개발 중에 흔한 실수(부작용이 있는 코드 등)를 찾도록 React가 추가 검사를 합니다.
//   - ThemeProvider: 밝은/어두운/시스템 테마를 정하고 브라우저(localStorage)에 기억합니다.
//   - TooltipProvider: 마우스를 올리면 뜨는 설명(툴팁)이 0.3초 뒤 나타나게 합니다.
//   - Toaster: 화면 위쪽에 잠깐 뜨는 알림("사건을 등록했습니다" 등)을 표시할 자리입니다.
createRoot(document.getElementById("root")).render(
  <StrictMode>
    <ThemeProvider defaultTheme="system" storageKey="casepulse.theme">
      <TooltipProvider delayDuration={300}>
        <App />
        <Toaster position="top-center" richColors />
      </TooltipProvider>
    </ThemeProvider>
  </StrictMode>,
);
