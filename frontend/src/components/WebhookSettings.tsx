import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface WebhookHeader {
  key: string;
  value: string;
}

const WebhookSettings: React.FC = () => {
  const navigate = useNavigate();
  const [url, setUrl] = useState('');
  const [method, setMethod] = useState('POST');
  const [headers, setHeaders] = useState<WebhookHeader[]>([{ key: '', value: '' }]);
  const [body, setBody] = useState('{\n  "text": "Hello World"\n}');
  const [jsonError, setJsonError] = useState<string | null>(null);

  React.useEffect(() => {
    try {
      JSON.parse(body);
      setJsonError(null);
    } catch (e) {
      if (e instanceof Error) {
        setJsonError(e.message);
      } else {
        setJsonError('Invalid JSON');
      }
    }
  }, [body]);

  const handleHeaderChange = (index: number, field: 'key' | 'value', value: string) => {
    const newHeaders = [...headers];
    newHeaders[index][field] = value;
    setHeaders(newHeaders);
  };

  const addHeader = () => {
    setHeaders([...headers, { key: '', value: '' }]);
  };

  const removeHeader = (index: number) => {
    const newHeaders = headers.filter((_, i) => i !== index);
    setHeaders(newHeaders);
  };

  const handleSave = () => {
    // Mock save functionality
    console.log('Saved Webhook Settings:', { url, method, headers, body });
    alert('Webhook settings saved (Check console for details)');
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
       <div className="flex items-center mb-6">
        <button 
          onClick={() => navigate('/settings')}
          className="mr-4 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
        >
          &larr; Back
        </button>
        <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">Webhook Management</h1>
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
              <textarea
                id="webhook-body"
                value={body}
                onChange={(e) => setBody(e.target.value)}
                rows={12}
                className={`w-full px-3 py-2 border rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 font-mono text-sm focus:ring-blue-500 focus:border-blue-500 ${
                  jsonError 
                    ? 'border-red-500 focus:border-red-500 focus:ring-red-500' 
                    : 'border-gray-300 dark:border-gray-600'
                }`}
              />
              {jsonError && (
                <p className="mt-1 text-sm text-red-500 dark:text-red-400">
                  {jsonError}
                </p>
              )}
            </div>
            <div className="w-1/3 bg-gray-50 dark:bg-gray-900/50 p-4 rounded-lg border border-gray-200 dark:border-gray-700">
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Insert Variables</h3>
              <div className="space-y-2">
                {[
                  { label: 'Incident ID', value: '{{incident.id}}' },
                  { label: 'Title', value: '{{incident.title}}' },
                  { label: 'Severity', value: '{{incident.severity}}' },
                  { label: 'Status', value: '{{incident.status}}' },
                  { label: 'Created At', value: '{{incident.created_at}}' },
                  { label: 'Summary', value: '{{incident.summary}}' },
                ].map((variable) => (
                  <button
                    key={variable.value}
                    onClick={() => {
                      const textarea = document.getElementById('webhook-body') as HTMLTextAreaElement;
                      if (!textarea) return;

                      const start = textarea.selectionStart;
                      const end = textarea.selectionEnd;
                      const text = body;
                      const before = text.substring(0, start);
                      const after = text.substring(end, text.length);
                      
                      const newBody = before + variable.value + after;
                      setBody(newBody);
                      
                      // Restore focus and cursor position next tick
                      setTimeout(() => {
                        textarea.focus();
                        textarea.setSelectionRange(start + variable.value.length, start + variable.value.length);
                      }, 0);
                    }}
                    className="w-full text-left px-3 py-2 text-sm bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded hover:bg-blue-50 dark:hover:bg-blue-900/30 hover:border-blue-300 dark:hover:border-blue-700 transition-colors text-gray-700 dark:text-gray-200"
                  >
                    <span className="font-mono text-blue-600 dark:text-blue-400 text-xs mr-2">{variable.value}</span>
                    <span className="text-gray-500 dark:text-gray-400 text-xs">- {variable.label}</span>
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        <div className="pt-4">
          <button
            onClick={handleSave}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
          >
            Save Settings
          </button>
        </div>
      </div>
    </div>
  );
};

export default WebhookSettings;
