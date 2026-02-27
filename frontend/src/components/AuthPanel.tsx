import { useState } from 'react';
import { login, register } from '../utils/auth';

interface AuthPanelProps {
  allowSignup: boolean;
  oidcEnabled: boolean;
  oidcLoginUrl: string;
  onAuthenticated: () => void;
}

type Mode = 'login' | 'register';

function getOidcError(): string | null {
  const params = new URLSearchParams(window.location.search);
  const oidcError = params.get('error');
  if (!oidcError) return null;

  // URL에서 error 파라미터 제거 (새로고침 시 에러 반복 방지)
  const url = new URL(window.location.href);
  url.searchParams.delete('error');
  window.history.replaceState({}, '', url.pathname + url.search);

  if (oidcError === 'oidc_not_allowed') return '허용되지 않은 이메일입니다. 관리자에게 문의하세요.';
  if (oidcError === 'oidc_failed') return 'Google 로그인에 실패했습니다. 다시 시도해주세요.';
  if (oidcError === 'oidc_state_mismatch') return '인증 세션이 만료되었습니다. 다시 시도해주세요.';
  if (oidcError === 'oidc_invalid_request') return '잘못된 인증 요청입니다.';
  return 'Google 로그인 중 오류가 발생했습니다.';
}

const AuthPanel = ({ allowSignup, oidcEnabled, oidcLoginUrl, onAuthenticated }: AuthPanelProps) => {
  const [mode, setMode] = useState<Mode>('login');
  const [loginId, setLoginId] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [oidcError] = useState<string | null>(getOidcError);
  const [error, setError] = useState<string | null>(null);

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
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 flex items-center justify-center px-4 transition-colors duration-300">
      <div className="w-full max-w-md bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6 transition-colors duration-300">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-white mb-2">Kube-RCA</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
          {mode === 'login' ? '아이디와 비밀번호로 로그인하세요.' : '새 계정을 생성하세요.'}
        </p>

        {oidcError && (
          <div className="mb-4 rounded-md border border-red-300 dark:border-red-700 bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-red-700 dark:text-red-300 flex items-start gap-2">
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
              className="w-full flex items-center justify-center gap-2 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
            >
              <svg className="w-5 h-5" viewBox="0 0 24 24">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"/>
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
              </svg>
              Google로 로그인
            </button>

            <div className="relative my-4">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-300 dark:border-gray-600" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="bg-white dark:bg-gray-800 px-2 text-gray-500 dark:text-gray-400">또는</span>
              </div>
            </div>
          </>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1" htmlFor="login-id">
              ID
            </label>
            <input
              id="login-id"
              type="text"
              value={loginId}
              onChange={(event) => setLoginId(event.target.value)}
              className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 placeholder:text-gray-400 dark:placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-gray-700 dark:focus:ring-gray-300"
              placeholder="아이디"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1" htmlFor="password">
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 placeholder:text-gray-400 dark:placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-gray-700 dark:focus:ring-gray-300"
              placeholder="비밀번호"
              required
            />
          </div>

          {error && (
            <div className="rounded-md border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 px-3 py-2 text-sm text-red-700 dark:text-red-300">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-md bg-gray-900 py-2 text-sm font-semibold text-white hover:bg-gray-800 disabled:opacity-60 dark:bg-gray-100 dark:text-gray-900 dark:hover:bg-gray-200"
          >
            {submitting ? '처리 중...' : mode === 'login' ? '로그인' : '회원가입'}
          </button>
        </form>

        {allowSignup && (
          <div className="mt-4 text-center text-sm text-gray-500 dark:text-gray-400">
            {mode === 'login' ? '계정이 없나요?' : '이미 계정이 있나요?'}{' '}
            <button
              type="button"
              className="font-semibold text-gray-800 dark:text-gray-200 hover:text-gray-900 dark:hover:text-white"
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
