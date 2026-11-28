import { createContext, useContext, useEffect, useState } from "react";

// shadcn/ui Vite 다크 모드 안내(https://ui.shadcn.com/docs/dark-mode/vite)를 JavaScript로 옮긴 것입니다.
// theme은 "light" | "dark" | "system"이며, 고른 값은 localStorage에 저장됩니다.
// resolvedTheme은 "system"일 때 운영체제 설정을 반영한 실제 값("light" | "dark")입니다.

// Context는 여러 단계 아래의 컴포넌트에도 값을 바로 전달하는 React 기능입니다.
// ThemeProvider가 값을 넣고, 어느 컴포넌트에서든 useTheme()로 꺼내 씁니다(props로 일일이 넘기지 않아도 됨).
const ThemeProviderContext = createContext(null);

// systemTheme은 운영체제가 어두운 화면 설정인지 CSS 미디어 쿼리로 확인합니다.

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

  // <html> 태그에 "light" 또는 "dark" 클래스를 붙입니다. Tailwind의 dark: 스타일은 이 클래스를 보고 적용됩니다.
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

// useTheme은 { theme, resolvedTheme, setTheme }을 돌려줍니다. ThemeProvider 바깥에서 부르면 실수이므로 오류를 냅니다.
export function useTheme() {
  const context = useContext(ThemeProviderContext);
  if (!context) throw new Error("useTheme must be used within a ThemeProvider");
  return context;
}
