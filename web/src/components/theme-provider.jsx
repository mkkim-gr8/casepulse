import { createContext, useContext, useEffect, useState } from "react";

// shadcn/ui Vite 다크 모드 안내(https://ui.shadcn.com/docs/dark-mode/vite)를 JavaScript로 옮긴 것입니다.
// theme은 "light" | "dark" | "system"이며, 고른 값은 localStorage에 저장됩니다.
// resolvedTheme은 "system"일 때 운영체제 설정을 반영한 실제 값("light" | "dark")입니다.

const ThemeProviderContext = createContext(null);

function systemTheme() {
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

export function ThemeProvider({ children, defaultTheme = "system", storageKey = "casepulse.theme" }) {
  const [theme, setThemeState] = useState(() => localStorage.getItem(storageKey) || defaultTheme);
  const [system, setSystem] = useState(systemTheme);

  // 운영체제의 밝은/어두운 설정이 바뀌면 "system" 모드에 바로 반영합니다.
  useEffect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = () => setSystem(systemTheme());
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  const resolvedTheme = theme === "system" ? system : theme;

  useEffect(() => {
    const root = window.document.documentElement;
    root.classList.remove("light", "dark");
    root.classList.add(resolvedTheme);
  }, [resolvedTheme]);

  const setTheme = (value) => {
    localStorage.setItem(storageKey, value);
    setThemeState(value);
  };

  return (
    <ThemeProviderContext.Provider value={{ theme, resolvedTheme, setTheme }}>{children}</ThemeProviderContext.Provider>
  );
}

export function useTheme() {
  const context = useContext(ThemeProviderContext);
  if (!context) throw new Error("useTheme must be used within a ThemeProvider");
  return context;
}
