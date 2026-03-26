import type { SecondBucket } from '../types/result';

export type PacketStatus = 'pass' | 'block' | 'rate-limit' | 'error';

export interface PacketData {
  id: string;
  status: PacketStatus;
  yOffset: number;
}

const MAX_DISPLAY_PACKETS = 20;

export function samplePackets(bucket: SecondBucket, second: number): PacketData[] {
  const entries: Array<{ code: string; count: number }> = [];
  for (const [code, count] of Object.entries(bucket.status_counts)) {
    if (count > 0) entries.push({ code, count });
  }
  if (bucket.error_count > 0) {
    entries.push({ code: 'error', count: bucket.error_count });
  }

  const total = entries.reduce((sum, e) => sum + e.count, 0);
  if (total === 0) return [];

  const displayCount = Math.min(total, MAX_DISPLAY_PACKETS);
  const packets: PacketData[] = [];

  for (const entry of entries) {
    const share = Math.max(1, Math.round((entry.count / total) * displayCount));
    for (let i = 0; i < share && packets.length < displayCount; i++) {
      packets.push({
        id: `${second}-${entry.code}-${i}`,
        status: codeToStatus(entry.code),
        yOffset: (Math.random() - 0.5) * 40,
      });
    }
  }

  return packets;
}

function codeToStatus(code: string): PacketStatus {
  if (code === 'error') return 'error';
  const num = parseInt(code, 10);
  if (num === 403) return 'block';
  if (num === 429) return 'rate-limit';
  if (num >= 200 && num < 400) return 'pass';
  return 'error';
}

export function statusToColor(status: PacketStatus): string {
  switch (status) {
    case 'pass': return '#22c55e';
    case 'block': return '#ef4444';
    case 'rate-limit': return '#f97316';
    case 'error': return '#6b7280';
  }
}
