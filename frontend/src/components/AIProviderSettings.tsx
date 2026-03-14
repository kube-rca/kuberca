import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchAppSetting, updateAppSetting, AISettings } from '../utils/api';

const MODEL_PLACEHOLDERS: Record<string, string> = {
  gemini: 'gemini-3-flash-preview',
  openai: 'gpt-5-mini',
  anthropic: 'claude-haiku-4-5',
};

const AIProviderSettings: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const [provider, setProvider] = useState('gemini');
  const [modelId, setModelId] = useState('');

  useEffect(() => {
    fetchAppSetting<AISettings>('ai')
      .then((data) => {
        if (data) {
          setProvider(data.provider || 'gemini');
          setModelId(data.modelId || '');
        }
      })
      .catch((err) => setMessage({ type: 'error', text: err.message }))
      .finally(() => setLoading(false));
  }, []);

  const handleSave = async () => {
    setSaving(true);
    setMessage(null);
    try {
      await updateAppSetting<AISettings>('ai', { provider, modelId });
      setMessage({ type: 'success', text: 'AI Provider settings saved. Reflected in Agent.' });
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

  const inputClass = "w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:ring-2 focus:ring-cyan-500 focus:border-transparent";
  const labelClass = "block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1";

  return (
    <div className="bg-white dark:bg-slate-800 rounded-lg shadow-md p-6 transition-colors duration-300">
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => navigate('/settings')} className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">
          &larr;
        </button>
        <h1 className="text-2xl font-semibold text-slate-800 dark:text-white">AI Provider</h1>
      </div>

      <p className="text-sm text-slate-500 dark:text-slate-400 mb-6">
        Change the AI Provider and Model to be used for RCA analysis.
        Changes are immediately reflected in the Agent and will be applied to the next analysis.
      </p>

      <div className="space-y-5 max-w-lg">
        {/* Provider Select */}
        <div>
          <label className={labelClass}>Provider</label>
          <select
            value={provider}
            onChange={(e) => {
              setProvider(e.target.value);
              setModelId('');
            }}
            className={inputClass}
          >
            <option value="gemini">Gemini</option>
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
          </select>
        </div>

        {/* Model ID */}
        <div>
          <label className={labelClass}>Model ID</label>
          <input
            type="text"
            value={modelId}
            onChange={(e) => setModelId(e.target.value)}
            placeholder={MODEL_PLACEHOLDERS[provider] || ''}
            className={inputClass}
          />
          <p className="text-xs text-slate-400 dark:text-slate-500 mt-1">
            If left empty, the default model set in Helm Values will be used.
          </p>
        </div>

        {/* API Key Notice */}
        <div className="bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-800 rounded-lg p-3">
          <p className="text-sm text-amber-700 dark:text-amber-400">
            For security reasons, the API Key cannot be changed on this screen.
            Manage it using Helm Values or K8s Secret.
          </p>
        </div>

        {/* Save */}
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
    </div>
  );
};

export default AIProviderSettings;
