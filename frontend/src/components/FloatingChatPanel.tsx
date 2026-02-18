import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from 'react';
import { useLocation } from 'react-router-dom';
import { chatWithAgent, ChatRequest } from '../utils/api';

type ChatMessage = {
  id: string;
  role: 'user' | 'assistant';
  content: string;
};

type RouteContext = {
  page: string;
  routeIncidentId: string;
  routeAlertId: string;
  routeKey: string;
};

interface FloatingChatPanelProps {
  onDockedChange?: (docked: boolean) => void;
}

const parseRouteContext = (pathname: string): RouteContext => {
  const incidentMatch = pathname.match(/^\/incidents\/([^/]+)$/);
  if (incidentMatch) {
    const incidentId = decodeURIComponent(incidentMatch[1]);
    return {
      page: 'incident_detail',
      routeIncidentId: incidentId,
      routeAlertId: '',
      routeKey: `incident:${incidentId}`,
    };
  }

  const alertMatch = pathname.match(/^\/alerts\/([^/]+)$/);
  if (alertMatch) {
    const alertId = decodeURIComponent(alertMatch[1]);
    return {
      page: 'alert_detail',
      routeIncidentId: '',
      routeAlertId: alertId,
      routeKey: `alert:${alertId}`,
    };
  }

  if (pathname === '/alerts') {
    return { page: 'alert_dashboard', routeIncidentId: '', routeAlertId: '', routeKey: 'alerts' };
  }

  if (pathname.startsWith('/muted')) {
    return { page: 'archived_dashboard', routeIncidentId: '', routeAlertId: '', routeKey: pathname };
  }

  return { page: 'main_dashboard', routeIncidentId: '', routeAlertId: '', routeKey: pathname };
};

const makeMessage = (role: 'user' | 'assistant', content: string): ChatMessage => ({
  id: `${role}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
  role,
  content,
});

const FloatingChatPanel = ({ onDockedChange }: FloatingChatPanelProps) => {
  const location = useLocation();
  const routeContext = useMemo(() => parseRouteContext(location.pathname), [location.pathname]);

  const [isOpen, setIsOpen] = useState(false);
  const [isDocked, setIsDocked] = useState(false);
  const [showContextInputs, setShowContextInputs] = useState(false);
  const [sending, setSending] = useState(false);
  const [input, setInput] = useState('');
  const [conversationId, setConversationId] = useState('');
  const [manualIncidentId, setManualIncidentId] = useState('');
  const [manualAlertId, setManualAlertId] = useState('');
  const [messages, setMessages] = useState<ChatMessage[]>([
    makeMessage('assistant', '질문을 입력하면 Incident/Alert 컨텍스트를 함께 분석해 답변합니다.'),
  ]);

  const listRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    onDockedChange?.(isDocked);
  }, [isDocked, onDockedChange]);

  useEffect(() => {
    if (!isOpen || !listRef.current) {
      return;
    }
    listRef.current.scrollTop = listRef.current.scrollHeight;
  }, [isOpen, messages]);

  useEffect(() => {
    if (manualIncidentId) {
      return;
    }
    setManualIncidentId(routeContext.routeIncidentId);
  }, [manualIncidentId, routeContext.routeIncidentId]);

  useEffect(() => {
    if (manualAlertId) {
      return;
    }
    setManualAlertId(routeContext.routeAlertId);
  }, [manualAlertId, routeContext.routeAlertId]);

  const sendChat = async (request: ChatRequest, userPreview?: string) => {
    if (sending) {
      return;
    }

    const preview = userPreview?.trim();
    if (preview) {
      setMessages((prev) => [...prev, makeMessage('user', preview)]);
    }

    setSending(true);
    try {
      const response = await chatWithAgent(request);
      setConversationId(response.conversation_id || request.conversation_id || '');
      setMessages((prev) => [...prev, makeMessage('assistant', response.answer)]);
    } catch (error) {
      const message = error instanceof Error ? error.message : '채팅 요청에 실패했습니다.';
      setMessages((prev) => [...prev, makeMessage('assistant', `오류: ${message}`)]);
    } finally {
      setSending(false);
    }
  };

  const handleSubmit = async () => {
    const message = input.trim();
    if (!message || sending) {
      return;
    }

    const incidentId = manualIncidentId.trim() || routeContext.routeIncidentId;
    const alertId = manualAlertId.trim() || routeContext.routeAlertId;

    const request: ChatRequest = {
      message,
      conversation_id: conversationId,
      page: routeContext.page,
      auto: false,
      incident_id: incidentId,
      alert_id: alertId,
      incident_title: '',
      incident_content: '',
      alert_title: '',
      alert_content: '',
    };

    setInput('');
    await sendChat(request, message);
  };

  const handleInputKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      void handleSubmit();
    }
  };

  const panelClass = isDocked
    ? 'fixed top-16 right-0 w-full md:w-[26rem] h-[calc(100vh-4rem)] bg-white dark:bg-gray-900 border-l border-gray-200 dark:border-gray-800 shadow-2xl z-50 flex flex-col'
    : 'fixed bottom-24 right-4 sm:right-6 w-[calc(100vw-2rem)] sm:w-[24rem] h-[32rem] bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-2xl shadow-2xl z-50 flex flex-col';

  return (
    <>
      {isOpen && (
        <section className={panelClass}>
          <header className="px-4 py-3 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between gap-2">
            <div>
              <p className="text-sm font-semibold text-gray-900 dark:text-gray-100">Agent Chat</p>
              <p className="text-xs text-gray-500 dark:text-gray-400">Page: {routeContext.page}</p>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setShowContextInputs((prev) => !prev)}
                className="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded-md text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
              >
                Context
              </button>
              <button
                onClick={() => setIsDocked((prev) => !prev)}
                className="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded-md text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
              >
                {isDocked ? '축소' : '확장'}
              </button>
              <button
                onClick={() => setIsOpen(false)}
                className="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded-md text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
              >
                닫기
              </button>
            </div>
          </header>

          {showContextInputs && (
            <div className="px-4 py-3 border-b border-gray-200 dark:border-gray-800 grid grid-cols-1 gap-2 bg-gray-50 dark:bg-gray-800/60">
              <input
                value={manualIncidentId}
                onChange={(e) => setManualIncidentId(e.target.value)}
                placeholder="incident_id (optional)"
                className="w-full px-3 py-2 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-sm text-gray-900 dark:text-gray-100"
              />
              <input
                value={manualAlertId}
                onChange={(e) => setManualAlertId(e.target.value)}
                placeholder="alert_id (optional)"
                className="w-full px-3 py-2 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-sm text-gray-900 dark:text-gray-100"
              />
              <p className="text-[11px] text-gray-500 dark:text-gray-400">
                상세 페이지에서는 ID가 자동으로 채워지며, 대시보드에서는 직접 넣어 정확도를 높일 수 있습니다.
              </p>
            </div>
          )}

          <div ref={listRef} className="flex-1 overflow-y-auto px-4 py-3 space-y-3">
            {messages.map((message) => (
              <div
                key={message.id}
                className={`max-w-[92%] px-3 py-2 rounded-xl text-sm whitespace-pre-wrap leading-relaxed ${
                  message.role === 'user'
                    ? 'ml-auto bg-blue-600 text-white'
                    : 'mr-auto bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-100'
                }`}
              >
                {message.content}
              </div>
            ))}
            {sending && (
              <div className="mr-auto max-w-[92%] px-3 py-2 rounded-xl text-sm bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300">
                답변 생성 중...
              </div>
            )}
          </div>

          <footer className="p-3 border-t border-gray-200 dark:border-gray-800">
            <div className="flex gap-2">
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleInputKeyDown}
                rows={2}
                placeholder="질문을 입력하세요 (Enter 전송, Shift+Enter 줄바꿈)"
                className="flex-1 resize-none px-3 py-2 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-sm text-gray-900 dark:text-gray-100"
              />
              <button
                onClick={() => {
                  void handleSubmit();
                }}
                disabled={sending || input.trim() === ''}
                className="px-4 py-2 rounded-md bg-blue-600 text-white text-sm font-semibold hover:bg-blue-700 disabled:opacity-60 disabled:cursor-not-allowed"
              >
                전송
              </button>
            </div>
          </footer>
        </section>
      )}

      <button
        onClick={() => setIsOpen((prev) => !prev)}
        className="fixed bottom-5 right-4 sm:right-6 z-50 h-14 px-5 rounded-full shadow-xl bg-blue-600 hover:bg-blue-700 text-white font-semibold text-sm"
      >
        {isOpen ? 'Chat 닫기' : 'AI Chat'}
      </button>
    </>
  );
};

export default FloatingChatPanel;
