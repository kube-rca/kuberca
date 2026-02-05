import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import './index.css';
import { ThemeProvider } from './context/ThemeProvider';
import { SearchProvider } from './context/SearchContext';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ThemeProvider>
      <SearchProvider>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </SearchProvider>
    </ThemeProvider>
  </React.StrictMode>
);