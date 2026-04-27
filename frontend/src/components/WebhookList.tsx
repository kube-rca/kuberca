import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchWebhookList, deleteWebhookConfig, WebhookConfig } from '../utils/api';

type WebhookType = 'Slack' | 'Teams' | 'HTTP';

const TYPE_BADGE: Record<WebhookType, string> = {
  Slack: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300',
  Teams: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/40 dark:text-indigo-300',
  HTTP: 'bg-slate-100 text-slate-700 dark:bg-slate-700 dark:text-slate-200',
};

const SEVERITY_BADGE: Record<string, string> = {
  critical: 'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300',
  warning:  'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  info:     'bg-sky-100 text-sky-700 dark:bg-sky-900/40 dark:text-sky-300',
};

const ALL_SEVERITIES = ['info', 'warning', 'critical'] as const;

// severities=[] 웹훅이 실제로 수신하는 severity 목록 계산.
// 다른 웹훅이 claim한 severity는 제외된다.
const getEffectiveSeverities = (cfg: WebhookConfig, allConfigs: WebhookConfig[]): string[] => {
  if (cfg.severities.length > 0) return cfg.severities;
  const claimed = new Set<string>();
  for (const other of allConfigs) {
    if (other.id !== cfg.id) {
      other.severities.forEach((s) => claimed.add(s));
    }
  }
  return ALL_SEVERITIES.filter((s) => !claimed.has(s));
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
    setError(null);
    const data = await fetchWebhookList();
    setConfigs(data);
  };

  useEffect(() => {
    void (async () => {
      setLoading(true);
      try {
        await load();
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Failed to load the list.');
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const handleDelete = async (e: React.MouseEvent, id: number) => {
    e.stopPropagation();
    if (!confirm('Are you sure you want to delete this webhook setting?')) return;
    setDeletingId(id);
    try {
      await deleteWebhookConfig(id);
      setConfigs((prev) => prev.filter((c) => c.id !== id));
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Failed to delete.');
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div className="bg-white dark:bg-slate-800 rounded-lg shadow-md p-6 transition-colors duration-300">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate('/settings')}
            className="text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200"
          >
            &larr; Back
          </button>
          <h1 className="text-2xl font-semibold text-slate-800 dark:text-white">Webhook Management</h1>
        </div>
        <button
          onClick={() => navigate('/settings/webhooks/new')}
          className="flex items-center gap-2 px-4 py-2 bg-cyan-600 text-white text-sm font-medium rounded-md hover:bg-cyan-700 focus:outline-none focus:ring-2 focus:ring-cyan-500 transition-colors"
        >
          <span className="text-lg leading-none">+</span>
          New Webhook
        </button>
      </div>

      {/* Body */}
      {loading ? (
        <div className="flex justify-center items-center py-16 text-slate-500 dark:text-slate-400">
          Loading...
        </div>
      ) : error ? (
        <div className="bg-rose-50 dark:bg-rose-900/20 border border-rose-200 dark:border-rose-800 rounded-md p-4 text-rose-600 dark:text-rose-400">
          {error}
        </div>
      ) : configs.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-slate-400 dark:text-slate-500">
          <svg className="w-12 h-12 mb-4 opacity-40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
              d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
          <p className="text-sm">No webhook settings registered.</p>
          <button
            onClick={() => navigate('/settings/webhooks/new')}
            className="mt-4 text-cyan-600 dark:text-cyan-400 text-sm hover:underline"
          >
            + Add new webhook setting
          </button>
        </div>
      ) : (
        <div className="space-y-3">
          {configs.map((cfg) => {
            const type = getWebhookType(cfg);
            const slackChannel = type === 'Slack' ? getSlackChannel(cfg) : '';
            const secondaryText =
              type === 'Slack'
                ? (slackChannel ? `Channel: ${slackChannel}` : 'Slack Bot')
                : cfg.url;

            return (
            <div
              key={cfg.id}
              onClick={() => navigate(`/settings/webhooks/${cfg.id}`)}
              className="group flex items-center justify-between p-4 border border-slate-200 dark:border-slate-700 rounded-lg hover:border-cyan-400 dark:hover:border-cyan-500 hover:shadow-sm cursor-pointer transition-all duration-150"
            >
              <div className="flex items-center gap-3 min-w-0">
                <span
                  className={`shrink-0 inline-block px-2 py-0.5 text-xs font-semibold rounded ${TYPE_BADGE[type]}`}
                >
                  {type}
                </span>
                <div className="min-w-0">
                  <p className="text-sm font-semibold text-slate-800 dark:text-slate-100 truncate">
                    {cfg.name || 'Unnamed Webhook'}
                  </p>
                  <p className="text-xs font-mono text-slate-500 dark:text-slate-400 truncate">
                    {secondaryText || 'No info'}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-3 shrink-0 ml-4">
                {(() => {
                  const effective = getEffectiveSeverities(cfg, configs);
                  if (effective.length === 0) {
                    return <span className="hidden sm:inline text-xs text-slate-400 dark:text-slate-500 italic">—</span>;
                  }
                  if (effective.length === ALL_SEVERITIES.length && cfg.severities.length === 0) {
                    return <span className="hidden sm:inline text-xs text-slate-400 dark:text-slate-500 italic">all</span>;
                  }
                  return (
                    <div className="hidden sm:flex items-center gap-1">
                      {effective.map((sev) => (
                        <span key={sev} className={`inline-block px-1.5 py-0.5 text-xs font-semibold rounded ${SEVERITY_BADGE[sev] ?? 'bg-slate-100 text-slate-600'}`}>
                          {sev}
                        </span>
                      ))}
                    </div>
                  );
                })()}
                <span className="text-xs text-slate-400 dark:text-slate-500 hidden sm:block">
                  {new Date(cfg.updated_at).toLocaleString('ko-KR', {
                    year: 'numeric', month: '2-digit', day: '2-digit',
                    hour: '2-digit', minute: '2-digit',
                  })}
                </span>
                <button
                  onClick={(e) => handleDelete(e, cfg.id)}
                  disabled={deletingId === cfg.id}
                  className="px-2 py-1 text-xs text-rose-500 hover:text-rose-700 dark:text-rose-400 dark:hover:text-rose-300 opacity-0 group-hover:opacity-100 transition-opacity disabled:opacity-50"
                >
                  {deletingId === cfg.id ? 'Deleting...' : 'Delete'}
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
