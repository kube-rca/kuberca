import React, { useEffect, useState } from 'react';
import { NavLink } from 'react-router-dom';
import { requestWithAuth } from '../utils/api';

interface AnalysisConfig {
  manualAnalyzeSeverities: string;
}

const SEVERITIES = ['warning', 'critical'] as const;

const AnalysisSettings: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  // Set of severities that require manual analysis
  const [manualSeverities, setManualSeverities] = useState<Set<string>>(new Set());

  useEffect(() => {
    const loadSettings = async () => {
      try {
        setLoading(true);
        const response = await requestWithAuth('/api/v1/settings/app/analysis', {
          method: 'GET',
        });
        if (response.ok) {
          const data = await response.json();
          const config: AnalysisConfig = data.data?.value ?? data.data ?? data;
          if (config.manualAnalyzeSeverities) {
            const parsed = config.manualAnalyzeSeverities
              .split(',')
              .map((s: string) => s.trim())
              .filter(Boolean);
            setManualSeverities(new Set(parsed));
          }
        }
      } catch (err) {
        console.error('분석 설정 로드 실패:', err);
        setError('설정을 불러오는데 실패했습니다.');
      } finally {
        setLoading(false);
      }
    };
    loadSettings();
  }, []);

  const handleSave = async () => {
    try {
      setSaving(true);
      setError(null);
      setSuccess(false);

      const payload: AnalysisConfig = {
        manualAnalyzeSeverities: Array.from(manualSeverities).join(','),
      };

      const response = await requestWithAuth('/api/v1/settings/app/analysis', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        throw new Error(`저장 실패 (${response.status})`);
      }

      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      console.error('분석 설정 저장 실패:', err);
      setError('설정 저장에 실패했습니다.');
    } finally {
      setSaving(false);
    }
  };

  const handleSeverityToggle = (severity: string) => {
    setManualSeverities((prev) => {
      const next = new Set(prev);
      if (next.has(severity)) {
        next.delete(severity);
      } else {
        next.add(severity);
      }
      return next;
    });
  };

  if (loading) {
    return (
      <div className="bg-white dark:bg-slate-800 rounded-lg shadow-md p-6 transition-colors duration-300 max-w-2xl mx-auto">
        <div className="space-y-4">
          <div className="skeleton h-8 w-48" />
          <div className="skeleton h-12 w-full rounded-lg" />
          <div className="skeleton h-12 w-full rounded-lg" />
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-slate-800 rounded-lg shadow-md p-6 transition-colors duration-300 max-w-2xl mx-auto">
      <div className="mb-6">
        <NavLink
          to="/settings"
          className="text-sm text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200 font-medium"
        >
          &larr; Settings
        </NavLink>
      </div>

      <h1 className="text-2xl font-semibold text-slate-800 dark:text-white mb-2">Analysis Mode</h1>
      <p className="text-sm text-slate-500 dark:text-slate-400 mb-6">
        기본적으로 모든 알림은 자동 분석됩니다. 수동 분석이 필요한 severity를 선택하세요.
      </p>

      {/* Manual Severities Selection */}
      <div className="mb-6">
        <div className="p-4 border border-slate-200 dark:border-slate-700 rounded-lg">
          <div className="font-medium text-slate-800 dark:text-slate-200 mb-3">
            수동 분석 대상 Severity
          </div>
          <div className="text-sm text-slate-500 dark:text-slate-400 mb-4">
            체크한 severity는 자동 분석되지 않으며, UI에서 직접 분석 버튼을 눌러야 합니다.
            아무것도 선택하지 않으면 모든 알림이 자동 분석됩니다.
          </div>
          <div className="flex gap-6">
            {SEVERITIES.map((severity) => (
              <label key={severity} className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={manualSeverities.has(severity)}
                  onChange={() => handleSeverityToggle(severity)}
                  className="w-4 h-4 text-violet-600 bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600 rounded focus:ring-violet-500"
                />
                <span className={`text-sm font-medium ${
                  severity === 'warning'
                    ? 'text-amber-700 dark:text-amber-400'
                    : 'text-rose-700 dark:text-rose-400'
                }`}>
                  {severity.charAt(0).toUpperCase() + severity.slice(1)}
                </span>
                <span className="text-xs text-slate-400 dark:text-slate-500">
                  {manualSeverities.has(severity) ? '(Manual)' : '(Auto)'}
                </span>
              </label>
            ))}
          </div>
        </div>
      </div>

      {/* Current Status Summary */}
      <div className="mb-6 p-3 bg-slate-50 dark:bg-slate-700/50 rounded-lg text-sm text-slate-600 dark:text-slate-300">
        {manualSeverities.size === 0 ? (
          <span>모든 알림이 자동 분석됩니다.</span>
        ) : manualSeverities.size === SEVERITIES.length ? (
          <span>모든 알림에 수동 분석이 필요합니다.</span>
        ) : (
          <span>
            <strong>{Array.from(manualSeverities).join(', ')}</strong> 알림은 수동 분석,
            나머지는 자동 분석됩니다.
          </span>
        )}
      </div>

      {/* Feedback Messages */}
      {error && (
        <div className="mb-4 p-3 bg-rose-50 dark:bg-rose-900/20 border border-rose-200 dark:border-rose-800 rounded-lg text-rose-600 dark:text-rose-400 text-sm">
          {error}
        </div>
      )}
      {success && (
        <div className="mb-4 p-3 bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-lg text-emerald-600 dark:text-emerald-400 text-sm">
          설정이 저장되었습니다.
        </div>
      )}

      {/* Save Button */}
      <button
        onClick={handleSave}
        disabled={saving}
        className="px-6 py-2.5 bg-violet-600 text-white text-sm font-semibold rounded-lg hover:bg-violet-700 transition-colors shadow-sm disabled:opacity-50"
      >
        {saving ? '저장 중...' : '저장'}
      </button>
    </div>
  );
};

export default AnalysisSettings;
