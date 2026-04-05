import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchAppSetting, updateAppSetting, fetchWebhookList, updateWebhookConfig, WebhookConfig } from '../utils/api';

interface NotificationSettingsData {
  enabled: boolean;
}

const ALL_SEVERITIES = ['info', 'warning', 'critical'] as const;

const SEVERITY_COLOR: Record<string, string> = {
  critical: 'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300',
  warning:  'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  info:     'bg-sky-100 text-sky-700 dark:bg-sky-900/40 dark:text-sky-300',
};

const NotificationSettings: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const [enabled, setEnabled] = useState(true);

  // Webhook severity routing state
  const [webhooks, setWebhooks] = useState<WebhookConfig[]>([]);
  const [webhooksLoading, setWebhooksLoading] = useState(true);
  const [severityMap, setSeverityMap] = useState<Record<number, string[]>>({});
  const [routingSaving, setRoutingSaving] = useState(false);
  const [routingMessage, setRoutingMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    fetchAppSetting<NotificationSettingsData>('notification')
      .then((data) => {
        if (data) setEnabled(data.enabled);
      })
      .catch((err) => setMessage({ type: 'error', text: err.message }))
      .finally(() => setLoading(false));

    fetchWebhookList()
      .then((list) => {
        setWebhooks(list);
        const map: Record<number, string[]> = {};
        list.forEach((w) => { map[w.id] = w.severities ?? []; });
        setSeverityMap(map);
      })
      .catch(() => {/* webhook list 실패는 무시 */})
      .finally(() => setWebhooksLoading(false));
  }, []);

  // 하나의 severity는 최대 하나의 웹훅에만 배정 가능
  const toggleSeverity = (webhookId: number, sev: string) => {
    setSeverityMap((prev) => {
      const current = prev[webhookId] ?? [];
      const isRemoving = current.includes(sev);

      if (isRemoving) {
        return { ...prev, [webhookId]: current.filter((s) => s !== sev) };
      }

      // 다른 웹훅에서 같은 severity 제거 후 현재 웹훅에 추가
      const updated: Record<number, string[]> = {};
      for (const id of Object.keys(prev)) {
        const numId = Number(id);
        updated[numId] = numId === webhookId
          ? [...current, sev]
          : (prev[numId] ?? []).filter((s) => s !== sev);
      }
      return updated;
    });
  };


  const handleSaveRouting = async () => {
    setRoutingSaving(true);
    setRoutingMessage(null);
    try {
      await Promise.all(
        webhooks.map((w) =>
          updateWebhookConfig(w.id, {
            name: w.name,
            url: w.url,
            type: w.type as 'slack' | 'teams' | 'http',
            token: w.token,
            channel: w.channel,
            severities: severityMap[w.id] ?? [],
          })
        )
      );
      setRoutingMessage({ type: 'success', text: 'Severity routing saved.' });
    } catch (err: unknown) {
      setRoutingMessage({ type: 'error', text: err instanceof Error ? err.message : 'Save failed' });
    } finally {
      setRoutingSaving(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setMessage(null);
    try {
      await updateAppSetting<NotificationSettingsData>('notification', { enabled });
      setMessage({ type: 'success', text: 'Notification settings saved.' });
    } catch (err: unknown) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Save failed' });
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="bg-white dark:bg-slate-800 rounded-lg shadow-md p-6">
        <div className="animate-pulse space-y-4">
          <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded w-1/3" />
          <div className="h-10 bg-slate-200 dark:bg-slate-700 rounded" />
        </div>
      </div>
    );
  }

  const labelClass = "block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1";

  return (
    <div className="bg-white dark:bg-slate-800 rounded-lg shadow-md p-6 transition-colors duration-300">
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => navigate('/settings')} className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">
          &larr;
        </button>
        <h1 className="text-2xl font-semibold text-slate-800 dark:text-white">Notification</h1>
      </div>

      <p className="text-sm text-slate-500 dark:text-slate-400 mb-6">
        Temporarily disable the notification pipeline.
        When disabled, even if Alertmanager webhooks are received, DB storage, AI analysis, and messenger notifications will be suspended.
        Useful during maintenance.
      </p>

      <div className="space-y-5 max-w-lg">
        <div className="flex items-center justify-between">
          <label className={labelClass}>Enable Notification Pipeline</label>
          <button
            type="button"
            onClick={() => setEnabled(!enabled)}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${enabled ? 'bg-cyan-600' : 'bg-slate-300 dark:bg-slate-600'}`}
          >
            <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${enabled ? 'translate-x-6' : 'translate-x-1'}`} />
          </button>
        </div>

        {!enabled && (
          <div className="p-3 rounded-lg text-sm bg-amber-50 dark:bg-amber-950/20 text-amber-600 dark:text-amber-400 border border-amber-200 dark:border-amber-800">
            Notifications are disabled. Alertmanager webhooks will not be processed even if received.
          </div>
        )}

        <div className="pt-2">
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-6 py-2 bg-cyan-600 hover:bg-cyan-700 text-white rounded-lg font-medium disabled:opacity-50 transition-colors"
          >
            {saving ? 'Saving...' : 'Save'}
          </button>
        </div>

        {message && (
          <div className={`p-3 rounded-lg text-sm ${message.type === 'success' ? 'bg-emerald-50 dark:bg-emerald-950/20 text-emerald-600 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800' : 'bg-rose-50 dark:bg-rose-950/20 text-rose-600 dark:text-rose-400 border border-rose-200 dark:border-rose-800'}`}>
            {message.text}
          </div>
        )}
      </div>

      {/* Webhook Severity Routing */}
      <div className="mt-8 pt-8 border-t border-slate-200 dark:border-slate-700">
        <h2 className="text-lg font-semibold text-slate-800 dark:text-white mb-1">Webhook Severity Routing</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400 mb-1">
          Configure which alert severities each webhook channel receives. Leave all unchecked to receive all severities.
        </p>
        <p className="text-xs text-amber-600 dark:text-amber-400 mb-5">
          Each severity can be assigned to at most one webhook. Assigning a severity to a new webhook will automatically remove it from the previous one.
        </p>

        {webhooksLoading ? (
          <div className="space-y-3">
            {[1, 2].map((i) => (
              <div key={i} className="animate-pulse h-14 bg-slate-100 dark:bg-slate-700/50 rounded-lg" />
            ))}
          </div>
        ) : webhooks.length === 0 ? (
          <div className="text-sm text-slate-400 dark:text-slate-500 py-4">
            No webhooks registered.{' '}
            <button onClick={() => navigate('/settings/webhooks/new')} className="text-cyan-600 dark:text-cyan-400 hover:underline">
              Add one
            </button>
          </div>
        ) : (
          <div className="space-y-3 max-w-2xl">
            {webhooks.map((w) => {
              const label = w.name || `Webhook #${w.id}`;
              const detail = w.type === 'slack'
                ? (w.channel ? `Slack · ${w.channel}` : 'Slack')
                : w.url || 'HTTP Webhook';
              const selected = severityMap[w.id] ?? [];
              return (
                <div key={w.id} className="flex flex-col sm:flex-row sm:items-center gap-3 p-4 border border-slate-200 dark:border-slate-700 rounded-lg">
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-semibold text-slate-700 dark:text-slate-200 truncate">{label}</p>
                    <p className="text-xs font-mono text-slate-500 dark:text-slate-400 truncate">{detail}</p>
                  </div>
                  <div className="flex items-center gap-3 shrink-0 flex-wrap">
                    {ALL_SEVERITIES.map((sev) => {
                      const checked = selected.includes(sev);
                      return (
                        <label key={sev} className="flex items-center gap-1.5 cursor-pointer select-none">
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={() => toggleSeverity(w.id, sev)}
                            className="w-4 h-4 rounded accent-cyan-600"
                          />
                          <span className={`text-xs font-semibold px-1.5 py-0.5 rounded ${SEVERITY_COLOR[sev]}`}>
                            {sev}
                          </span>
                        </label>
                      );
                    })}
                  </div>
                </div>
              );
            })}

            <div className="pt-2 flex items-center gap-3">
              <button
                onClick={handleSaveRouting}
                disabled={routingSaving}
                className="px-6 py-2 bg-cyan-600 hover:bg-cyan-700 text-white rounded-lg font-medium disabled:opacity-50 transition-colors"
              >
                {routingSaving ? 'Saving...' : 'Save Routing'}
              </button>
              {routingMessage && (
                <span className={`text-sm ${routingMessage.type === 'success' ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}`}>
                  {routingMessage.text}
                </span>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default NotificationSettings;
