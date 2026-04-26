import '@testing-library/jest-dom';

// jsdom does not implement window.matchMedia — provide a minimal stub so
// components that query media features (e.g. ThemeProvider) render without error.
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  }),
});
