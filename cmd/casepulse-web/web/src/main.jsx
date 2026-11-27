import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import { ThemeProvider } from "@/components/theme-provider";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Toaster } from "@/components/ui/sonner";
import App from "./App.jsx";

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
