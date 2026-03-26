export const STATUS_COLORS: Record<string, string> = {
  '200': '#22c55e',
  '201': '#22c55e',
  '204': '#22c55e',
  '301': '#3b82f6',
  '302': '#3b82f6',
  '304': '#3b82f6',
  '403': '#ef4444',
  '429': '#f97316',
  '500': '#a855f7',
  '502': '#a855f7',
  '503': '#a855f7',
  error: '#6b7280',
};

export function getStatusColor(code: string | number): string {
  return STATUS_COLORS[String(code)] ?? '#94a3b8';
}

export function getStatusCategory(code: number): 'success' | 'redirect' | 'blocked' | 'server_error' | 'error' {
  if (code >= 200 && code < 300) return 'success';
  if (code >= 300 && code < 400) return 'redirect';
  if (code === 403 || code === 429) return 'blocked';
  if (code >= 500) return 'server_error';
  return 'error';
}
