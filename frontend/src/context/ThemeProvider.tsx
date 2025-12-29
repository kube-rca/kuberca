import { useEffect, useState, ReactNode } from 'react';
import { Theme, ThemeContext } from './ThemeContext';

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
  const [theme, setTheme] = useState<Theme>(() => {
    // 1. 로컬 스토리지 확인
    const storedTheme = localStorage.getItem('theme') as Theme | null;
    if (storedTheme) return storedTheme;
    
    // 2. 시스템 설정 확인
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark';
    
    return 'light';
  });

  useEffect(() => {
    const root = window.document.documentElement;

    // Tailwind 방식: classList 조작
    if (theme === 'dark') {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
    
    localStorage.setItem('theme', theme);
  }, [theme]);

  const toggleTheme = () => {
    setTheme((prev) => (prev === 'light' ? 'dark' : 'light'));
  };

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  );
};
