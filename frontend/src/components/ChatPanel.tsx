import { useState, useRef, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { LLM_API_KEY } from '../utils/config';

type Message = { role: 'user' | 'assistant'; content: string };

const SYSTEM_PROMPT =
  'You are a helpful assistant for an Incident and Alert monitoring dashboard. Answer briefly in Korean when possible.';

async function chatWithOpenAI(messages: Message[], apiKey: string): Promise<string> {
  const res = await fetch('https://api.openai.com/v1/chat/completions', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${apiKey}`,
    },
    body: JSON.stringify({
      model: 'gpt-4o-mini',
      messages: [
        { role: 'system', content: SYSTEM_PROMPT },
        ...messages.map((m) => ({ role: m.role, content: m.content })),
      ],
    }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: { message?: string } })?.error?.message || res.statusText);
  }
  const data = (await res.json()) as { choices?: Array<{ message?: { content?: string } }> };
  const text = data.choices?.[0]?.message?.content?.trim();
  return text || '(응답 없음)';
}

export default function ChatPanel() {
  const [messages, setMessages] = useState<Message[]>([
    {
      role: 'assistant',
      content: 'Incident·Alert 대시보드 관련해서 궁금한 점을 물어보세요.',
    },
  ]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: 'smooth' });
  }, [messages]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const text = input.trim();
    if (!text || !LLM_API_KEY || loading) return;

    setInput('');
    setError(null);
    const userMessage: Message = { role: 'user', content: text };
    setMessages((prev) => [...prev, userMessage]);
    setLoading(true);

    try {
      const nextMessages = [...messages, userMessage];
      const reply = await chatWithOpenAI(nextMessages, LLM_API_KEY);
      setMessages((prev) => [...prev, { role: 'assistant', content: reply }]);
    } catch (err) {
      const message = err instanceof Error ? err.message : '요청에 실패했습니다.';
      setError(message);
      setMessages((prev) => [...prev, { role: 'assistant', content: `오류: ${message}` }]);
    } finally {
      setLoading(false);
      inputRef.current?.focus();
    }
  };

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div
        ref={scrollRef}
        className="flex-1 overflow-y-auto px-4 py-3"
        style={{ minHeight: 0 }}
      >
        {messages.map((msg, i) => (
          <div
            key={i}
            className={`mb-3 ${msg.role === 'user' ? 'flex justify-end' : ''}`}
          >
            <div
              className={`max-w-[85%] rounded-lg px-3 py-2 text-sm ${
                msg.role === 'user'
                  ? 'bg-blue-600 text-white dark:bg-blue-500'
                  : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
              }`}
            >
              {msg.role === 'assistant' ? (
                <div className="prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-ul:my-1 prose-li:my-0">
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>{msg.content}</ReactMarkdown>
                </div>
              ) : (
                <span className="whitespace-pre-wrap">{msg.content}</span>
              )}
            </div>
          </div>
        ))}
        {loading && (
          <div className="mb-3 flex">
            <div className="rounded-lg bg-gray-100 px-3 py-2 text-sm text-gray-500 dark:bg-gray-700 dark:text-gray-400">
              입력 중...
            </div>
          </div>
        )}
      </div>

      {error && (
        <div className="shrink-0 px-4 py-1 text-xs text-red-600 dark:text-red-400">{error}</div>
      )}

      <form onSubmit={handleSubmit} className="shrink-0 border-t border-gray-200 p-3 dark:border-gray-700">
        <div className="flex gap-2">
          <input
            ref={inputRef}
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="메시지를 입력하세요..."
            disabled={loading || !LLM_API_KEY}
            className="flex-1 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 dark:placeholder-gray-400"
            autoComplete="off"
          />
          <button
            type="submit"
            disabled={loading || !input.trim() || !LLM_API_KEY}
            className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50 dark:bg-blue-500 dark:hover:bg-blue-600"
          >
            전송
          </button>
        </div>
      </form>
    </div>
  );
}
