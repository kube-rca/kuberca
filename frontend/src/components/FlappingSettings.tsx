import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchAppSetting, updateAppSetting, FlappingSettings as FlappingSettingsType } from '../utils/api';

const FlappingSettings: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const [enabled, setEnabled] = useState(true);
  const [detectionWindowMinutes, setDetectionWindowMinutes] = useState(30);
  const [cycleThreshold, setCycleThreshold] = useState(3);
  const [clearanceWindowMinutes, setClearanceWindowMinutes] = useState(30);

  useEffect(() => {
    fetchAppSetting<FlappingSettingsType>('flapping')
      .then((data) => {
        if (data) {
          setEnabled(data.enabled);
          setDetectionWindowMinutes(data.detectionWindowMinutes);
          setCycleThreshold(data.cycleThreshold);
          setClearanceWindowMinutes(data.clearanceWindowMinutes);
        }
      })
      .catch((err) => setMessage({ type: 'error', text: err.message }))
      .finally(() => setLoading(false));
  }, []);

  const handleSave = async () => {
    setSaving(true);
    setMessage(null);
    try {
      await updateAppSetting<FlappingSettingsType>('flapping', {
        enabled,
        detectionWindowMinutes,
        cycleThreshold,
        clearanceWindowMinutes,
      });
      setMessage({ type: 'success', text: 'Flapping Detection settings saved.' });
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
        <h1 className="text-2xl font-semibold text-slate-800 dark:text-white">Flapping Detection</h1>
      </div>

      <p className="text-sm text-slate-500 dark:text-slate-400 mb-6">
        This setting suppresses notifications when an Alert repeatedly flaps between firing/resolved within a short period.
      </p>

      <div className="space-y-5 max-w-lg">
        {/* Enabled Toggle */}
        <div className="flex items-center justify-between">
          <label className={labelClass}>Enable Flapping Detection</label>
          <button
            type="button"
            onClick={() => setEnabled(!enabled)}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${enabled ? 'bg-cyan-600' : 'bg-slate-300 dark:bg-slate-600'}`}
          >
            <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${enabled ? 'translate-x-6' : 'translate-x-1'}`} />
          </button>
        </div>

        {/* Detection Window */}
        <div>
          <label className={labelClass}>Detection Window (minutes)</label>
          <input
            type="number"
            min={1}
            value={detectionWindowMinutes}
            onChange={(e) => setDetectionWindowMinutes(Number(e.target.value))}
            className={inputClass}
            disabled={!enabled}
          />
          <p className="text-xs text-slate-400 dark:text-slate-500 mt-1">Counts the number of state transitions within this time.</p>
        </div>

        {/* Cycle Threshold */}
        <div>
          <label className={labelClass}>Cycle Threshold</label>
          <input
            type="number"
            min={1}
            value={cycleThreshold}
            onChange={(e) => setCycleThreshold(Number(e.target.value))}
            className={inputClass}
            disabled={!enabled}
          />
          <p className="text-xs text-slate-400 dark:text-slate-500 mt-1">Determines flapping if firing→resolved transitions exceed this count.</p>
        </div>

        {/* Clearance Window */}
        <div>
          <label className={labelClass}>Clearance Window (minutes)</label>
          <input
            type="number"
            min={1}
            value={clearanceWindowMinutes}
            onChange={(e) => setClearanceWindowMinutes(Number(e.target.value))}
            className={inputClass}
            disabled={!enabled}
          />
          <p className="text-xs text-slate-400 dark:text-slate-500 mt-1">Releases flapping state if stable for this duration after the last resolved.</p>
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

export default FlappingSettings;
