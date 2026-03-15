import React, { useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { fetchWebhookById, createWebhookConfig, updateWebhookConfig, WebhookConfigPayload } from '../utils/api';

type WebhookType = 'slack' | 'teams' | 'http';

const SLACK_POST_MESSAGE_URL = 'https://slack.com/api/chat.postMessage';

const buildConfigPayload = (type: WebhookType, url: string, token: string, channel: string, severities: string[]): WebhookConfigPayload => {
  const trimmedToken = token.trim();
  const trimmedChannel = channel.trim();

  return {
    type,
    url: type === 'slack' ? SLACK_POST_MESSAGE_URL : url,
    token: trimmedToken || undefined,
    channel: trimmedChannel || undefined,
    severities,
  };
};

const detectSavedWebhookType = (cfg: Awaited<ReturnType<typeof fetchWebhookById>>): WebhookType => {
  if (cfg.type === 'slack' || cfg.type === 'teams' || cfg.type === 'http') {
    return cfg.type;
  }
  return 'http';
};

const WebhookSettings: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id?: string }>();
  const isEditMode = Boolean(id);

  const [url, setUrl] = useState('');
  const [webhookType, setWebhookType] = useState<WebhookType>('slack');
  const [token, setToken] = useState('');
  const [channel, setChannel] = useState('');
  const [severities, setSeverities] = useState<string[]>([]);
  const [isSaving, setIsSaving] = useState(false);
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [showToken, setShowToken] = useState(false);

  React.useEffect(() => {
    if (!isEditMode || !id) return;

    fetchWebhookById(Number(id))
      .then((cfg) => {
        setUrl(cfg.url);
        setWebhookType(detectSavedWebhookType(cfg));

        const authValue = cfg.token ?? '';
        const bearerPrefix = 'Bearer ';
        if (authValue.startsWith(bearerPrefix)) {
          setToken(authValue.slice(bearerPrefix.length));
        } else {
          setToken(authValue);
        }

        const channelValue = cfg.channel ?? '';
        setChannel(channelValue);
        setSeverities(cfg.severities ?? []);
      })
      .catch((err) => {
        setSaveMessage({ type: 'error', text: `Failed to load settings: ${err instanceof Error ? err.message : err}` });
      });
  }, [id, isEditMode]);

  const urlError = useMemo(() => {
    if (webhookType === 'slack') {
      return null;
    }

    const trimmedUrl = url.trim();
    if (!trimmedUrl) return 'Please enter the Webhook receiver URL.';

    try {
      const parsed = new URL(trimmedUrl);
      if (!['http:', 'https:'].includes(parsed.protocol)) {
        return 'Webhook URL must be http or https protocol.';
      }
      return null;
    } catch {
      return 'Invalid URL format.';
    }
  }, [url, webhookType]);

  const slackError = useMemo(() => {
    if (webhookType !== 'slack') {
      return null;
    }
    if (!token.trim()) {
      return 'Please enter the Slack Bot Token.';
    }
    if (!channel.trim()) {
      return 'Please enter the Slack Channel ID.';
    }
    return null;
  }, [token, channel, webhookType]);

  const handleSave = async () => {
    if (isSaving) return;

    const trimmedUrl = url.trim();
    if (webhookType !== 'slack' && (!trimmedUrl || urlError)) {
      setSaveMessage({ type: 'error', text: urlError ?? 'Please enter the Webhook receiver URL.' });
      return;
    }

    if (slackError) {
      setSaveMessage({ type: 'error', text: slackError });
      return;
    }

    setIsSaving(true);
    setSaveMessage(null);

    try {
      const payload = buildConfigPayload(webhookType, trimmedUrl, token, channel, severities);

      if (isEditMode && id) {
        await updateWebhookConfig(Number(id), payload);
      } else {
        await createWebhookConfig(payload);
      }

      // 신규 생성 시 severity 배정 화면으로 이동
      navigate(isEditMode ? '/settings/webhooks' : '/settings/notifications');
    } catch (err) {
      setSaveMessage({
        type: 'error',
        text: err instanceof Error ? err.message : 'Failed to save.',
      });
      setIsSaving(false);
    }
  };

  return (
    <div className="bg-white dark:bg-slate-800 rounded-lg shadow-md p-6 transition-colors duration-300">
      <div className="flex items-center mb-6">
        <button
          onClick={() => navigate('/settings/webhooks')}
          className="mr-4 text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200"
        >
          &larr; Back
        </button>
        <h1 className="text-2xl font-semibold text-slate-800 dark:text-white">
          {isEditMode ? 'Edit Webhook' : 'New Webhook'}
        </h1>
      </div>

      <div className="space-y-6 max-w-3xl">
        <div>
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Webhook Type</label>
          <select
            value={webhookType}
            onChange={(e) => setWebhookType(e.target.value as WebhookType)}
            className="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 focus:ring-cyan-500 focus:border-cyan-500"
          >
            <option value="slack">Slack</option>
            <option value="teams">Microsoft Teams</option>
            <option value="http">Generic HTTP</option>
          </select>
        </div>

        {webhookType === 'slack' ? (
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Slack API Endpoint</label>
            <input
              type="text"
              value={SLACK_POST_MESSAGE_URL}
              readOnly
              className="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-md bg-slate-100 dark:bg-slate-700/60 text-slate-700 dark:text-slate-300"
            />
          </div>
        ) : (
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Webhook Receiver URL</label>
            <input
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://..."
              className="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 focus:ring-cyan-500 focus:border-cyan-500"
            />
            {urlError && <p className="mt-1 text-sm text-rose-500 dark:text-rose-400">{urlError}</p>}
          </div>
        )}

        {webhookType === 'slack' && (
          <>
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Slack Bot Token</label>
              <div className="relative">
                <input
                  type={showToken ? 'text' : 'password'}
                  value={token}
                  onChange={(e) => setToken(e.target.value)}
                  placeholder="xoxb-..."
                  className="w-full px-3 py-2 pr-10 border border-slate-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 focus:ring-cyan-500 focus:border-cyan-500"
                />
                <button
                  type="button"
                  onClick={() => setShowToken((v) => !v)}
                  className="absolute inset-y-0 right-0 flex items-center px-3 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 focus:outline-none"
                  aria-label={showToken ? 'Hide token' : 'Show token'}
                >
                  {showToken ? (
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                    </svg>
                  ) : (
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path strokeLinecap="round" strokeLinejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                  )}
                </button>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Slack Channel ID</label>
              <input
                type="text"
                value={channel}
                onChange={(e) => setChannel(e.target.value)}
                placeholder="C0123456789"
                className="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 focus:ring-cyan-500 focus:border-cyan-500"
              />
            </div>
            {slackError && <p className="mt-1 text-sm text-rose-500 dark:text-rose-400">{slackError}</p>}
          </>
        )}

        {webhookType === 'http' && (
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Bearer Token (Optional)
            </label>
            <div className="relative">
              <input
                type={showToken ? 'text' : 'password'}
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder="Enter only if a token is required"
                className="w-full px-3 py-2 pr-10 border border-slate-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 focus:ring-cyan-500 focus:border-cyan-500"
              />
              <button
                type="button"
                onClick={() => setShowToken((v) => !v)}
                className="absolute inset-y-0 right-0 flex items-center px-3 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 focus:outline-none"
                aria-label={showToken ? 'Hide token' : 'Show token'}
              >
                {showToken ? (
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                  </svg>
                ) : (
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path strokeLinecap="round" strokeLinejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                )}
              </button>
            </div>
          </div>
        )}

        {!isEditMode && (
          <div className="p-3 rounded-lg bg-sky-50 dark:bg-sky-950/20 border border-sky-200 dark:border-sky-800 text-sm text-sky-700 dark:text-sky-300">
            After saving, you'll be taken to <strong>Notification Settings</strong> to assign which alert severities this webhook receives.
          </div>
        )}

        <div className="pt-2">
          {saveMessage && (
            <p
              className={`mb-3 text-sm ${
                saveMessage.type === 'success' ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'
              }`}
            >
              {saveMessage.text}
            </p>
          )}
          <div className="flex gap-3">
            <button
              onClick={handleSave}
              disabled={isSaving || !!urlError || !!slackError}
              className="px-4 py-2 bg-cyan-600 text-white rounded-md hover:bg-cyan-700 focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:ring-offset-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSaving ? 'Saving...' : isEditMode ? 'Update Settings' : 'Save Settings'}
            </button>
            <button
              onClick={() => navigate('/settings/webhooks')}
              className="px-4 py-2 border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300 rounded-md hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default WebhookSettings;
