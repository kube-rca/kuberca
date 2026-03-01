import { useState } from 'react';
import { Hexagon, Sun, Moon } from 'lucide-react';
import { login, register } from '../utils/auth';
import { useTheme } from '../context/ThemeContext';

interface AuthPanelProps {
  allowSignup: boolean;
  oidcEnabled: boolean;
  oidcLoginUrl: string;
  oidcProvider: string;
  onAuthenticated: () => void;
}

type Mode = 'login' | 'register';

function getOidcError(): string | null {
  const params = new URLSearchParams(window.location.search);
  const oidcError = params.get('error');
  if (!oidcError) return null;

  const url = new URL(window.location.href);
  url.searchParams.delete('error');
  window.history.replaceState({}, '', url.pathname + url.search);

  if (oidcError === 'oidc_not_allowed') return '허용되지 않은 이메일입니다. 관리자에게 문의하세요.';
  if (oidcError === 'oidc_failed') return '로그인에 실패했습니다. 다시 시도해주세요.';
  if (oidcError === 'oidc_state_mismatch') return '인증 세션이 만료되었습니다. 다시 시도해주세요.';
  if (oidcError === 'oidc_invalid_request') return '잘못된 인증 요청입니다.';
  return '로그인 중 오류가 발생했습니다.';
}

const providerConfig: Record<string, { label: string; icon: React.ReactNode }> = {
  google: {
    label: 'Google로 로그인',
    icon: (
      <svg className="w-5 h-5" viewBox="0 0 24 24">
        <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"/>
        <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
        <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
        <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
      </svg>
    ),
  },
  azure: {
    label: 'Microsoft로 로그인',
    icon: (
      <svg className="w-5 h-5" viewBox="0 0 24 24">
        <path fill="#F25022" d="M1 1h10v10H1z"/>
        <path fill="#00A4EF" d="M1 13h10v10H1z"/>
        <path fill="#7FBA00" d="M13 1h10v10H13z"/>
        <path fill="#FFB900" d="M13 13h10v10H13z"/>
      </svg>
    ),
  },
  gitlab: {
    label: 'GitLab으로 로그인',
    icon: (
      <svg className="w-5 h-5" viewBox="0 0 24 24">
        <path fill="#E24329" d="m12 22.2 4.03-12.4H7.97z"/>
        <path fill="#FC6D26" d="m12 22.2-4.03-12.4H1.52z"/>
        <path fill="#FCA326" d="M1.52 9.8 0.34 13.4a.81.81 0 0 0 .29.9L12 22.2z"/>
        <path fill="#E24329" d="M1.52 9.8h6.45L5.24 1.65a.39.39 0 0 0-.74 0z"/>
        <path fill="#FC6D26" d="m12 22.2 4.03-12.4h6.45z"/>
        <path fill="#FCA326" d="m22.48 9.8 1.18 3.6a.81.81 0 0 1-.29.9L12 22.2z"/>
        <path fill="#E24329" d="M22.48 9.8h-6.45l2.73-8.15a.39.39 0 0 1 .74 0z"/>
      </svg>
    ),
  },
  keycloak: {
    label: 'Keycloak으로 로그인',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" />
      </svg>
    ),
  },
  okta: {
    label: 'Okta로 로그인',
    icon: (
      <svg className="w-5 h-5" viewBox="0 0 24 24">
        <path fill="#007DC1" d="M12 0C5.389 0 0 5.389 0 12s5.389 12 12 12 12-5.389 12-12S18.611 0 12 0zm0 18c-3.314 0-6-2.686-6-6s2.686-6 6-6 6 2.686 6 6-2.686 6-6 6z"/>
      </svg>
    ),
  },
  oidc: {
    label: 'SSO로 로그인',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1 1 21.75 8.25Z" />
      </svg>
    ),
  },
};

const AuthPanel = ({ allowSignup, oidcEnabled, oidcLoginUrl, oidcProvider, onAuthenticated }: AuthPanelProps) => {
  const { theme, toggleTheme } = useTheme();
  const [mode, setMode] = useState<Mode>('login');
  const [loginId, setLoginId] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [oidcError] = useState<string | null>(getOidcError);
  const [error, setError] = useState<string | null>(null);

  const provider = providerConfig[oidcProvider] || providerConfig.oidc;

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setSubmitting(true);
    setError(null);

    try {
      if (mode === 'login') {
        await login(loginId, password);
      } else {
        await register(loginId, password);
      }
      onAuthenticated();
    } catch (err) {
      console.error(err);
      setError(mode === 'login' ? '로그인에 실패했습니다.' : '회원가입에 실패했습니다.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-950 bg-grid-pattern flex items-center justify-center px-4 transition-colors duration-300">
      <div className="w-full max-w-md bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-lg p-8 transition-colors duration-300 relative">
        <button
          onClick={toggleTheme}
          className="absolute top-4 right-4 p-2 rounded-lg text-slate-400 dark:text-slate-500 hover:text-slate-700 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          aria-label="Toggle Dark Mode"
        >
          {theme === 'light' ? <Moon className="w-4 h-4" /> : <Sun className="w-4 h-4" />}
        </button>
        <div className="flex items-center gap-3 mb-2">
          <Hexagon className="w-8 h-8 text-cyan-600 dark:text-cyan-400" strokeWidth={2} />
          <h1 className="text-2xl font-semibold font-mono tracking-wider text-slate-900 dark:text-slate-100">Kube-RCA</h1>
        </div>
        <p className="text-sm text-slate-500 dark:text-slate-400 mb-6">
          {mode === 'login' ? '아이디와 비밀번호로 로그인하세요.' : '새 계정을 생성하세요.'}
        </p>

        {oidcError && (
          <div className="mb-4 rounded-md border border-rose-300 dark:border-rose-700 bg-rose-50 dark:bg-rose-900/30 px-4 py-3 text-sm text-rose-700 dark:text-rose-300 flex items-start gap-2">
            <svg className="w-5 h-5 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
            </svg>
            <span>{oidcError}</span>
          </div>
        )}

        {oidcEnabled && (
          <>
            <button
              type="button"
              onClick={() => { window.location.href = oidcLoginUrl; }}
              className="w-full flex items-center justify-center gap-2 rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 py-2 text-sm font-medium text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
            >
              {provider.icon}
              {provider.label}
            </button>

            <div className="relative my-4">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-slate-300 dark:border-slate-700" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="bg-white dark:bg-slate-900 px-2 text-slate-500 dark:text-slate-400">또는</span>
              </div>
            </div>
          </>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1" htmlFor="login-id">
              ID
            </label>
            <input
              id="login-id"
              type="text"
              value={loginId}
              onChange={(event) => setLoginId(event.target.value)}
              className="w-full rounded-md border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:focus:ring-cyan-400"
              placeholder="아이디"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1" htmlFor="password">
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              className="w-full rounded-md border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:focus:ring-cyan-400"
              placeholder="비밀번호"
              required
            />
          </div>

          {error && (
            <div className="rounded-md border border-rose-200 dark:border-rose-800 bg-rose-50 dark:bg-rose-900/20 px-3 py-2 text-sm text-rose-700 dark:text-rose-300">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-md bg-cyan-600 hover:bg-cyan-700 dark:bg-cyan-500 dark:hover:bg-cyan-400 py-2 text-sm font-semibold text-white disabled:opacity-60"
          >
            {submitting ? '처리 중...' : mode === 'login' ? '로그인' : '회원가입'}
          </button>
        </form>

        {allowSignup && (
          <div className="mt-4 text-center text-sm text-slate-500 dark:text-slate-400">
            {mode === 'login' ? '계정이 없나요?' : '이미 계정이 있나요?'}{' '}
            <button
              type="button"
              className="font-semibold text-slate-800 dark:text-slate-200 hover:text-slate-900 dark:hover:text-white"
              onClick={() => setMode(mode === 'login' ? 'register' : 'login')}
            >
              {mode === 'login' ? '회원가입' : '로그인'}
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default AuthPanel;
