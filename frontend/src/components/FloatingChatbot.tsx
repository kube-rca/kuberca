import { useState, useMemo } from 'react';
import { useLocation } from 'react-router-dom';
import { LLM_API_KEY } from '../utils/config';
import ChatPanel from './ChatPanel';

/** Incident, Alert, 메인 리스트(및 상세) 경로에서만 챗봇 표시 */
const CHATBOT_PATHS = ['/', '/alerts', '/muted'];
const CHATBOT_PATH_PREFIXES = ['/incidents/', '/alerts/', '/muted/'];

function useShowChatbot(): boolean {
  const location = useLocation();
  return useMemo(() => {
    const path = location.pathname;
    if (CHATBOT_PATHS.includes(path)) return true;
    return CHATBOT_PATH_PREFIXES.some((prefix) => path.startsWith(prefix));
  }, [location.pathname]);
}

export default function FloatingChatbot() {
  const [expanded, setExpanded] = useState(false);
  const showChatbot = useShowChatbot();

  if (!showChatbot) return null;

  const hasApiKey = Boolean(LLM_API_KEY?.trim());

  return (
    <>
      {/* 플로팅 버튼 - 패널이 열려 있을 땐 숨겨서 입력창이 가려지지 않도록 함 */}
      {!expanded && (
        <button
          type="button"
          onClick={() => setExpanded(true)}
          className="fixed bottom-6 right-6 z-50 flex h-14 w-14 items-center justify-center rounded-full bg-blue-600 text-white shadow-lg transition-all hover:bg-blue-700 hover:shadow-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:bg-blue-500 dark:hover:bg-blue-600"
          aria-label="챗봇 열기"
        >
          <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden>
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
            />
          </svg>
        </button>
      )}

      {/* Expanded Chat 패널 - z-50으로 챗봇이 최상단에 오도록 */}
      {expanded && (
        <div
          className="fixed bottom-6 right-6 z-50 flex h-[min(70vh,520px)] w-[min(400px,calc(100vw-3rem))] flex-col overflow-hidden rounded-xl border border-gray-200 bg-white shadow-2xl dark:border-gray-700 dark:bg-gray-800"
          role="dialog"
          aria-label="챗봇 대화"
        >
          <div className="flex shrink-0 items-center justify-between border-b border-gray-200 px-4 py-3 dark:border-gray-700">
            <span className="font-medium text-gray-800 dark:text-white">챗봇</span>
            <button
              type="button"
              onClick={() => setExpanded(false)}
              className="rounded p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-gray-600 dark:hover:text-gray-300"
              aria-label="닫기"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div className="relative min-h-0 flex-1 overflow-hidden">
            {hasApiKey ? (
              <ChatPanel />
            ) : (
              <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-center">
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  챗봇을 사용하려면 LLM API Key가 필요합니다.
                </p>
                <p className="rounded-lg bg-amber-50 px-3 py-2 text-left font-mono text-xs text-amber-800 dark:bg-amber-900/30 dark:text-amber-200">
                  .env 파일에 다음 변수를 추가한 뒤 값을 채워주세요:
                  <br />
                  <strong>VITE_LLM_API_KEY=your-openai-api-key</strong>
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  (OpenAI API Key 등 호환되는 키를 사용할 수 있습니다.)
                </p>
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
}
