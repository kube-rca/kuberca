import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchWebhookList, deleteWebhookConfig, WebhookConfig } from '../utils/api';

type WebhookType = 'Slack' | 'Teams' | 'HTTP';

const TYPE_BADGE: Record<WebhookType, string> = {
  Slack: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300',
  Teams: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/40 dark:text-indigo-300',
  HTTP: 'bg-slate-100 text-slate-700 dark:bg-slate-700 dark:text-slate-200',
};

const getWebhookType = (cfg: WebhookConfig): WebhookType => {
  const value = cfg.type?.toLowerCase();
  if (value === 'slack') return 'Slack';
  if (value === 'teams') return 'Teams';
  return 'HTTP';
};

const getSlackChannel = (cfg: WebhookConfig): string => {
  return cfg.channel?.trim() ?? '';
};

const WebhookList: React.FC = () => {
  const navigate = useNavigate();
  const [configs, setConfigs] = useState<WebhookConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);

  const load = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await fetchWebhookList();
      setConfigs(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : '목록을 불러오는데 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  const handleDelete = async (e: React.MouseEvent, id: number) => {
    e.stopPropagation();
    if (!confirm('이 웹훅 설정을 삭제하시겠습니까?')) return;
    setDeletingId(id);
    try {
      await deleteWebhookConfig(id);
      setConfigs((prev) => prev.filter((c) => c.id !== id));
    } catch (e) {
      alert(e instanceof Error ? e.message : '삭제에 실패했습니다.');
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate('/settings')}
            className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
          >
            &larr; Back
          </button>
          <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">Webhook Management</h1>
        </div>
        <button
          onClick={() => navigate('/settings/webhooks/new')}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors"
        >
          <span className="text-lg leading-none">+</span>
          New Webhook
        </button>
      </div>

      {/* Body */}
      {loading ? (
        <div className="flex justify-center items-center py-16 text-gray-500 dark:text-gray-400">
          불러오는 중...
        </div>
      ) : error ? (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 text-red-600 dark:text-red-400">
          {error}
        </div>
      ) : configs.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-gray-400 dark:text-gray-500">
          <svg className="w-12 h-12 mb-4 opacity-40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
              d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
          <p className="text-sm">등록된 웹훅 설정이 없습니다.</p>
          <button
            onClick={() => navigate('/settings/webhooks/new')}
            className="mt-4 text-blue-600 dark:text-blue-400 text-sm hover:underline"
          >
            + 새 웹훅 설정 추가
          </button>
        </div>
      ) : (
        <div className="space-y-3">
          {configs.map((cfg) => {
            const type = getWebhookType(cfg);
            const slackChannel = type === 'Slack' ? getSlackChannel(cfg) : '';
            const primaryText =
              type === 'Slack'
                ? (slackChannel ? `Channel: ${slackChannel}` : 'Slack Bot')
                : cfg.url;

            return (
            <div
              key={cfg.id}
              onClick={() => navigate(`/settings/webhooks/${cfg.id}`)}
              className="group flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg hover:border-blue-400 dark:hover:border-blue-500 hover:shadow-sm cursor-pointer transition-all duration-150"
            >
              <div className="flex items-center gap-3 min-w-0">
                <span
                  className={`shrink-0 inline-block px-2 py-0.5 text-xs font-semibold rounded ${TYPE_BADGE[type]}`}
                >
                  {type}
                </span>
                <span className="text-sm font-mono text-gray-800 dark:text-gray-100 truncate">
                  {primaryText || <span className="text-gray-400 italic">정보 없음</span>}
                </span>
              </div>
              <div className="flex items-center gap-3 shrink-0 ml-4">
                <span className="text-xs text-gray-400 dark:text-gray-500 hidden sm:block">
                  {new Date(cfg.updated_at).toLocaleString('ko-KR', {
                    year: 'numeric', month: '2-digit', day: '2-digit',
                    hour: '2-digit', minute: '2-digit',
                  })}
                </span>
                <button
                  onClick={(e) => handleDelete(e, cfg.id)}
                  disabled={deletingId === cfg.id}
                  className="px-2 py-1 text-xs text-red-500 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300 opacity-0 group-hover:opacity-100 transition-opacity disabled:opacity-50"
                >
                  {deletingId === cfg.id ? '삭제 중...' : '삭제'}
                </button>
              </div>
            </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default WebhookList;
