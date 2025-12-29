import { useState } from 'react';
import { login, register } from '../utils/auth';

interface AuthPanelProps {
  allowSignup: boolean;
  onAuthenticated: () => void;
}

type Mode = 'login' | 'register';

const AuthPanel = ({ allowSignup, onAuthenticated }: AuthPanelProps) => {
  const [mode, setMode] = useState<Mode>('login');
  const [loginId, setLoginId] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);
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
