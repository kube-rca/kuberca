import React, { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import Editor from 'react-simple-code-editor';
import Prism from 'prismjs';
import 'prismjs/components/prism-json';
import 'prismjs/themes/prism.css';
import { useUndo } from '../hooks/useUndo';
import { fetchWebhookById, createWebhookConfig, updateWebhookConfig, WebhookHeaderItem } from '../utils/api';

// Custom styler to highlight variables
const highlightWithPrism = (code: string) => {
  let highlighted = Prism.highlight(code, Prism.languages.json, 'json');
  highlighted = highlighted.replace(
    /{{[\w.]+}}/g,
    (match) => `<span class="bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded px-0.5 font-bold">${match}</span>`
  );
  return highlighted;
};

interface WebhookHeader {
  key: string;
  value: string;
}

const WebhookSettings: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id?: string }>();
  const isEditMode = Boolean(id);

  const [url, setUrl] = useState('');
  const [method, setMethod] = useState('POST');
  const [headers, setHeaders] = useState<WebhookHeader[]>([{ key: '', value: '' }]);
  const { state: body, setState: setBody, undo, redo } = useUndo('{\n  "text": "Hello World"\n}');
  const [jsonError, setJsonError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  // JSON 유효성 검사
  React.useEffect(() => {
    try {
      JSON.parse(body);
      setJsonError(null);
    } catch (e) {
      setJsonError(e instanceof Error ? e.message : 'Invalid JSON');
    }
  }, [body]);

  // 편집 모드: 기존 설정 로드
  React.useEffect(() => {
    if (!isEditMode || !id) return;
    fetchWebhookById(Number(id))
      .then((cfg) => {
        setUrl(cfg.url);
        setMethod(cfg.method);
        if (Array.isArray(cfg.headers) && cfg.headers.length > 0) {
          setHeaders(cfg.headers);
        }
        if (cfg.body) setBody(cfg.body);
      })
      .catch((err) => {
        setSaveMessage({ type: 'error', text: `설정 로드 실패: ${err instanceof Error ? err.message : err}` });
      });
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleHeaderChange = (index: number, field: 'key' | 'value', value: string) => {
    const newHeaders = [...headers];
    newHeaders[index][field] = value;
    setHeaders(newHeaders);
  };

  const addHeader = () => setHeaders([...headers, { key: '', value: '' }]);

  const removeHeader = (index: number) => {
    setHeaders(headers.filter((_: WebhookHeader, i: number) => i !== index));
  };

  const handleSave = async () => {
    if (isSaving || jsonError) return;
    setIsSaving(true);
    setSaveMessage(null);
    try {
      const filteredHeaders: WebhookHeaderItem[] = headers.filter((h) => h.key.trim() !== '');
      const payload = { url, method, headers: filteredHeaders, body };

      if (isEditMode && id) {
        await updateWebhookConfig(Number(id), payload);
      } else {
        await createWebhookConfig(payload);
      }

      // 저장 성공 → 목록으로 이동
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

      <div className="space-y-6 max-w-4xl">
        {/* URL and Method */}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Webhook URL</label>
          <div className="flex gap-2">
            <select
              value={method}
              onChange={(e) => setMethod(e.target.value)}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="DELETE">DELETE</option>
            </select>
            <input
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://example.com/webhook"
              className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
        </div>

        {/* Headers */}
        <div>
          <div className="flex justify-between items-center mb-2">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Headers</label>
            <button
              onClick={addHeader}
              className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300"
            >
              + Add Header
            </button>
          </div>
          <div className="space-y-2">
            {headers.map((header, index) => (
              <div key={index} className="flex gap-2">
                <input
                  type="text"
                  placeholder="Key"
                  value={header.key}
                  onChange={(e) => handleHeaderChange(index, 'key', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
                />
                <input
                  type="text"
                  placeholder="Value"
                  value={header.value}
                  onChange={(e) => handleHeaderChange(index, 'value', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
                />
                {headers.length > 1 && (
                  <button
                    onClick={() => removeHeader(index)}
                    className="px-2 text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300"
                  >
                    &times;
                  </button>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Body */}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Body (JSON)</label>
          <div className="flex gap-4">
            <div className="flex-1">
              <div className={`relative w-full border rounded-md bg-white dark:bg-gray-700 overflow-hidden ${
                  jsonError
                    ? 'border-red-500 ring-1 ring-red-500'
                    : 'border-gray-300 dark:border-gray-600 focus-within:ring-2 focus-within:ring-blue-500 focus-within:border-blue-500'
                }`}>
                <Editor
                  value={body}
                  onValueChange={(code) => setBody(code)}
                  highlight={highlightWithPrism}
                  padding={10}
                  textareaId="webhook-body-editor"
                  className="font-mono text-sm"
                  style={{
                    fontFamily: '"Fira code", "Fira Mono", monospace',
                    fontSize: 14,
                    minHeight: '300px',
                  }}
                  onKeyDown={(e) => {
                    const { key, ctrlKey, metaKey, shiftKey } = e;
                    if ((ctrlKey || metaKey) && key === 'z') {
                      e.preventDefault();
                      shiftKey ? redo() : undo();
                      return;
                    }
                    if (key === 'Tab') {
                      e.preventDefault();
                      const textarea = e.target as HTMLTextAreaElement;
                      const start = textarea.selectionStart;
                      const end = textarea.selectionEnd;
                      const newBody = body.substring(0, start) + '  ' + body.substring(end);
                      setBody(newBody);
                      setTimeout(() => {
                        textarea.selectionStart = textarea.selectionEnd = start + 2;
                      }, 0);
                    }
                  }}
                />
              </div>
              {jsonError && (
                <p className="mt-1 text-sm text-red-500 dark:text-red-400">{jsonError}</p>
              )}
            </div>
            <div className="w-1/3 bg-slate-100 dark:bg-slate-800 p-4 rounded-lg border border-slate-200 dark:border-slate-700">
              <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3 uppercase tracking-wider">Available Variables</h3>
              <div className="space-y-2">
                {[
                  { label: 'Incident ID',  value: '{{incident.id}}' },
                  { label: 'Title',        value: '{{incident.title}}' },
                  { label: 'Severity',     value: '{{incident.severity}}' },
                  { label: 'Status',       value: '{{incident.status}}' },
                  { label: 'Created At',   value: '{{incident.created_at}}' },
                  { label: 'Summary',      value: '{{incident.summary}}' },
                ].map((variable) => (
                  <button
                    key={variable.value}
                    onClick={() => {
                      const textarea = document.getElementById('webhook-body-editor') as HTMLTextAreaElement;
                      if (!textarea) return;
                      const start = textarea.selectionStart;
                      const end = textarea.selectionEnd;
                      const newBody = body.substring(0, start) + variable.value + body.substring(end);
                      setBody(newBody);
                      setTimeout(() => {
                        textarea.focus();
                        textarea.setSelectionRange(start + variable.value.length, start + variable.value.length);
                      }, 0);
                    }}
                    className="w-full text-left px-3 py-2 text-sm bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-md shadow-sm hover:bg-blue-50 dark:hover:bg-blue-900/50 hover:border-blue-300 dark:hover:border-blue-500 transition-all duration-200 group"
                  >
                    <div className="flex flex-col">
                      <span className="font-mono text-blue-600 dark:text-blue-400 text-xs font-semibold group-hover:text-blue-700 dark:group-hover:text-blue-300">{variable.value}</span>
                      <span className="text-slate-500 dark:text-slate-400 text-xs mt-0.5">{variable.label}</span>
                    </div>
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Save Button */}
        <div className="pt-4">
          {saveMessage && (
            <p className={`mb-3 text-sm ${
              saveMessage.type === 'success'
                ? 'text-green-600 dark:text-green-400'
                : 'text-red-600 dark:text-red-400'
            }`}>
              {saveMessage.text}
            </p>
          )}
          <div className="flex gap-3">
            <button
              onClick={handleSave}
              disabled={isSaving || !!jsonError}
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
