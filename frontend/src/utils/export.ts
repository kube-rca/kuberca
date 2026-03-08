export type ExportFormat = 'csv' | 'excel';

export interface ExportColumn<T> {
  key: string;
  header: string;
  value: (row: T) => string | number | null | undefined;
}

export interface ExportSummary {
  total: number;
  firing: number;
  resolved: number;
}

const sanitizeCell = (value: unknown): string => {
  if (value === null || value === undefined) return '';
  return String(value);
};

const escapeCsvCell = (value: string): string => {
  if (/[",\n\r]/.test(value)) {
    return `"${value.replace(/"/g, '""')}"`;
  }
  return value;
};

const buildCsv = <T>(rows: T[], columns: ExportColumn<T>[], summary?: ExportSummary): string => {
  const headers = columns.map((col) => escapeCsvCell(col.header)).join(',');
  const lines = rows.map((row) =>
    columns
      .map((col) => escapeCsvCell(sanitizeCell(col.value(row))))
      .join(',')
  );

  if (!summary) {
    return [headers, ...lines].join('\r\n');
  }

  const summaryLines = [
    'Metric,Count',
    `Total,${summary.total}`,
    `Firing,${summary.firing}`,
    `Resolved,${summary.resolved}`,
    '',
  ];

  return [...summaryLines, headers, ...lines].join('\r\n');
};

const escapeHtml = (value: string): string =>
  value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');

const buildExcelHtml = <T>(rows: T[], columns: ExportColumn<T>[], summary?: ExportSummary): string => {
  const head = columns.map((col) => `<th>${escapeHtml(col.header)}</th>`).join('');
  const colgroup = columns
    .map((col) => {
      const width = col.key.includes('title') ? 360 : col.key.includes('at') ? 180 : 140;
      return `<col style="width:${width}px;">`;
    })
    .join('');
  const body = rows
    .map((row) => {
      const tds = columns
        .map((col) => `<td style="mso-number-format:'\\@';">${escapeHtml(sanitizeCell(col.value(row)))}</td>`)
        .join('');
      return `<tr>${tds}</tr>`;
    })
    .join('');

  const summaryHtml = summary
    ? `<table style="margin-bottom:12px;"><thead><tr><th>Metric</th><th>Count</th></tr></thead><tbody><tr><td>Total</td><td>${summary.total}</td></tr><tr><td>Firing</td><td>${summary.firing}</td></tr><tr><td>Resolved</td><td>${summary.resolved}</td></tr></tbody></table>`
    : '';

  return `<!doctype html><html><head><meta charset="utf-8"></head><body>${summaryHtml}<table><colgroup>${colgroup}</colgroup><thead><tr>${head}</tr></thead><tbody>${body}</tbody></table></body></html>`;
};

const downloadBlob = (blob: Blob, filename: string) => {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
};

export const exportRows = <T>(
  rows: T[],
  columns: ExportColumn<T>[],
  baseFilename: string,
  format: ExportFormat,
  summary?: ExportSummary
) => {
  if (!rows.length) return;

  const now = new Date();
  const stamp = `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(
    now.getDate()
  ).padStart(2, '0')}_${String(now.getHours()).padStart(2, '0')}${String(now.getMinutes()).padStart(
    2,
    '0'
  )}`;
  const filename = `${baseFilename}_${stamp}.${format === 'csv' ? 'csv' : 'xls'}`;

  if (format === 'csv') {
    const csv = buildCsv(rows, columns, summary);
    downloadBlob(new Blob([`\uFEFF${csv}`], { type: 'text/csv;charset=utf-8;' }), filename);
    return;
  }

  const html = buildExcelHtml(rows, columns, summary);
  downloadBlob(new Blob([html], { type: 'application/vnd.ms-excel;charset=utf-8;' }), filename);
};
