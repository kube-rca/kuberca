import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from 'react';
import { useLocation } from 'react-router-dom';
import { useLanguage } from '../context/LanguageContext';
import { chatWithAgent, ChatRequest, fetchAlertDetail } from '../utils/api';

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
  const { language, t } = useLanguage();
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
    makeMessage('assistant', t('chatIntro')),
  ]);

  const listRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    // `t` is derived from `language` via context; depending on it would create a stale-closure loop.
    void (async () => {
      await Promise.resolve();
      setMessages([makeMessage('assistant', t('chatIntro'))]);
    })();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [language]);

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
    let cancelled = false;

    const syncContextFromRoute = async () => {
      if (routeContext.page === 'incident_detail') {
        setManualIncidentId(routeContext.routeIncidentId);
        setManualAlertId('');
        return;
      }

      if (routeContext.page === 'alert_detail') {
        setManualAlertId(routeContext.routeAlertId);
        setManualIncidentId('');
        try {
          const detail = await fetchAlertDetail(routeContext.routeAlertId);
          if (cancelled) {
            return;
          }
          setManualIncidentId(detail.incident_id || '');
        } catch {
          if (!cancelled) {
            setManualIncidentId('');
          }
        }
        return;
      }

      // Dashboard 이동 시 컨텍스트 입력값을 비웁니다.
      setManualIncidentId('');
      setManualAlertId('');
    };

    void syncContextFromRoute();
    return () => {
      cancelled = true;
    };
  }, [routeContext.page, routeContext.routeIncidentId, routeContext.routeAlertId, routeContext.routeKey]);

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
      const message = error instanceof Error ? error.message : 'Failed chat request.';
      setMessages((prev) => [...prev, makeMessage('assistant', `Error: ${message}`)]);
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
      language,
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
    ? 'fixed top-16 right-0 w-full md:w-[26rem] h-[calc(100vh-4rem)] bg-white dark:bg-slate-900 border-l border-slate-200 dark:border-slate-800 shadow-2xl z-50 flex flex-col'
    : 'fixed bottom-24 right-4 sm:right-6 w-[calc(100vw-2rem)] sm:w-[24rem] h-[32rem] bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-2xl shadow-2xl z-50 flex flex-col';

  return (
    <>
      {isOpen && (
        <section className={panelClass}>
          <header className="px-4 py-3 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between gap-2">
            <div>
              <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">{t('agentChat')}</p>
              <p className="text-xs text-slate-500 dark:text-slate-400">{t('page')}: {routeContext.page}</p>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setShowContextInputs((prev) => !prev)}
                className="px-2 py-1 text-xs border border-slate-300 dark:border-slate-600 rounded-md text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800"
              >
                {t('context')}
              </button>
              <button
                onClick={() => setIsDocked((prev) => !prev)}
                className="px-2 py-1 text-xs border border-slate-300 dark:border-slate-600 rounded-md text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800"
              >
                {isDocked ? t('collapse') : t('expand')}
              </button>
              <button
                onClick={() => setIsOpen(false)}
                className="px-2 py-1 text-xs border border-slate-300 dark:border-slate-600 rounded-md text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800"
              >
                {t('close')}
              </button>
            </div>
          </header>

          {showContextInputs && (
            <div className="px-4 py-3 border-b border-slate-200 dark:border-slate-800 grid grid-cols-1 gap-2 bg-slate-50 dark:bg-slate-800/60">
              <input
                value={manualIncidentId}
                onChange={(e) => setManualIncidentId(e.target.value)}
                placeholder={t('incidentOptional')}
                className="w-full px-3 py-2 rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-900 text-sm text-slate-900 dark:text-slate-100"
              />
              <input
                value={manualAlertId}
                onChange={(e) => setManualAlertId(e.target.value)}
                placeholder={t('alertOptional')}
                className="w-full px-3 py-2 rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-900 text-sm text-slate-900 dark:text-slate-100"
              />
              <p className="text-[11px] text-slate-500 dark:text-slate-400">
                {t('contextHint')}
              </p>
            </div>
          )}

          <div ref={listRef} className="flex-1 overflow-y-auto px-4 py-3 space-y-3">
            {messages.map((message) => (
              <div
                key={message.id}
                className={`max-w-[92%] px-3 py-2 rounded-xl text-sm whitespace-pre-wrap leading-relaxed ${
                  message.role === 'user'
                    ? 'ml-auto bg-cyan-600 text-white'
                    : 'mr-auto bg-slate-100 dark:bg-slate-800 text-slate-800 dark:text-slate-100'
                }`}
              >
                {message.content}
              </div>
            ))}
            {sending && (
              <div className="mr-auto max-w-[92%] px-3 py-2 rounded-xl text-sm bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
                {t('generating')}
              </div>
            )}
          </div>

          <footer className="p-3 border-t border-slate-200 dark:border-slate-800">
            <div className="flex gap-2">
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleInputKeyDown}
                rows={2}
                placeholder="Enter your question (Enter to send, Shift+Enter to newline)"
                className="flex-1 resize-none px-3 py-2 rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-900 text-sm text-slate-900 dark:text-slate-100"
              />
              <button
                onClick={() => {
                  void handleSubmit();
                }}
                disabled={sending || input.trim() === ''}
                className="px-4 py-2 rounded-md bg-cyan-600 text-white text-sm font-semibold hover:bg-cyan-700 disabled:opacity-60 disabled:cursor-not-allowed"
              >
                Send
              </button>
            </div>
          </footer>
        </section>
      )}

      <button
        onClick={() => setIsOpen((prev) => !prev)}
        className="fixed bottom-5 right-4 sm:right-6 z-50 h-14 px-5 rounded-full shadow-xl bg-cyan-600 hover:bg-cyan-700 text-white font-semibold text-sm"
      >
        {isOpen ? t('closeChat') : t('aiChat')}
      </button>
    </>
  );
};

export default FloatingChatPanel;
