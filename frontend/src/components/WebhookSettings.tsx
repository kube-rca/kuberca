import React, { useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { fetchWebhookById, createWebhookConfig, updateWebhookConfig, WebhookConfigPayload } from '../utils/api';

type WebhookType = 'slack' | 'teams' | 'http';

const WEBHOOK_LABEL: Record<WebhookType, string> = {
  slack: 'Slack',
  teams: 'Microsoft Teams',
  http: 'Generic HTTP',
};

const SLACK_POST_MESSAGE_URL = 'https://slack.com/api/chat.postMessage';

const buildConfigPayload = (type: WebhookType, url: string, token: string, channel: string): WebhookConfigPayload => {
  const trimmedToken = token.trim();
  const trimmedChannel = channel.trim();

  return {
    type,
    url: type === 'slack' ? SLACK_POST_MESSAGE_URL : url,
    token: trimmedToken || undefined,
    channel: trimmedChannel || undefined,
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
  const [isSaving, setIsSaving] = useState(false);
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

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
      })
      .catch((err) => {
        setSaveMessage({ type: 'error', text: `설정 로드 실패: ${err instanceof Error ? err.message : err}` });
      });
  }, [id, isEditMode]);

  const urlError = useMemo(() => {
    if (webhookType === 'slack') {
      return null;
    }

    const trimmedUrl = url.trim();
    if (!trimmedUrl) return 'Webhook receiver URL을 입력해 주세요.';

    try {
      const parsed = new URL(trimmedUrl);
      if (!['http:', 'https:'].includes(parsed.protocol)) {
        return 'Webhook URL은 http 또는 https 프로토콜이어야 합니다.';
      }
      return null;
    } catch {
      return '유효한 URL 형식이 아닙니다.';
    }
  }, [url, webhookType]);

  const slackError = useMemo(() => {
    if (webhookType !== 'slack') {
      return null;
    }
    if (!token.trim()) {
      return 'Slack Bot Token을 입력해 주세요.';
    }
    if (!channel.trim()) {
      return 'Slack Channel ID를 입력해 주세요.';
    }
    return null;
  }, [token, channel, webhookType]);

  const handleSave = async () => {
    if (isSaving) return;

    const trimmedUrl = url.trim();
    if (webhookType !== 'slack' && (!trimmedUrl || urlError)) {
      setSaveMessage({ type: 'error', text: urlError ?? 'Webhook receiver URL을 입력해 주세요.' });
      return;
    }

    if (slackError) {
      setSaveMessage({ type: 'error', text: slackError });
      return;
    }

    setIsSaving(true);
    setSaveMessage(null);

    try {
      const payload = buildConfigPayload(webhookType, trimmedUrl, token, channel);

      if (isEditMode && id) {
        await updateWebhookConfig(Number(id), payload);
      } else {
        await createWebhookConfig(payload);
      }

      navigate('/settings/webhooks');
    } catch (err) {
      setSaveMessage({
        type: 'error',
        text: err instanceof Error ? err.message : '저장에 실패했습니다.',
      });
      setIsSaving(false);
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
      <div className="flex items-center mb-6">
        <button
          onClick={() => navigate('/settings/webhooks')}
          className="mr-4 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
        >
          &larr; Back
        </button>
        <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">
          {isEditMode ? 'Edit Webhook' : 'New Webhook'}
        </h1>
      </div>

      <div className="space-y-6 max-w-3xl">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Webhook Type</label>
          <select
            value={webhookType}
            onChange={(e) => setWebhookType(e.target.value as WebhookType)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="slack">Slack</option>
            <option value="teams">Microsoft Teams</option>
            <option value="http">Generic HTTP</option>
          </select>
        </div>

        {webhookType === 'slack' ? (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Slack API Endpoint</label>
            <input
              type="text"
              value={SLACK_POST_MESSAGE_URL}
              readOnly
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-gray-100 dark:bg-gray-700/60 text-gray-700 dark:text-gray-300"
            />
          </div>
        ) : (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Webhook Receiver URL</label>
            <input
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://..."
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
            />
            {urlError && <p className="mt-1 text-sm text-red-500 dark:text-red-400">{urlError}</p>}
          </div>
        )}

        {webhookType === 'slack' && (
          <>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Slack Bot Token</label>
              <input
                type="password"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder="xoxb-..."
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Slack Channel ID</label>
              <input
                type="text"
                value={channel}
                onChange={(e) => setChannel(e.target.value)}
                placeholder="C0123456789"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            {slackError && <p className="mt-1 text-sm text-red-500 dark:text-red-400">{slackError}</p>}
          </>
        )}

        {webhookType === 'http' && (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Bearer Token (Optional)
            </label>
            <input
              type="password"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder="토큰이 필요한 경우만 입력"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
        )}

        <div className="rounded-md border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 px-4 py-3">
          <h2 className="text-sm font-semibold text-blue-800 dark:text-blue-300 mb-1">자동 적용 설정</h2>
          <p className="text-sm text-blue-700 dark:text-blue-200">
            {WEBHOOK_LABEL[webhookType]} 연동 정보만 frontend에서 전달하고, 알림 본문은 backend가 생성합니다.
          </p>
          <p className="text-xs text-blue-600 dark:text-blue-300 mt-1">
            Slack 타입은 Bot Token + Channel ID를 저장해 `chat.postMessage` 기반으로 알림을 전송합니다.
          </p>
        </div>

        <div className="pt-2">
          {saveMessage && (
            <p
              className={`mb-3 text-sm ${
                saveMessage.type === 'success' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
              }`}
            >
              {saveMessage.text}
            </p>
          )}
          <div className="flex gap-3">
            <button
              onClick={handleSave}
              disabled={isSaving || !!urlError || !!slackError}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSaving ? 'Saving...' : isEditMode ? 'Update Settings' : 'Save Settings'}
            </button>
            <button
              onClick={() => navigate('/settings/webhooks')}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
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
