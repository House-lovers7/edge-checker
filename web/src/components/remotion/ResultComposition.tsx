import { AbsoluteFill, useCurrentFrame, interpolate } from 'remotion';
import type { Result } from '../../types/result';
import { getStatusColor } from '../../utils/colors';

interface Props {
  result: Result;
}

const FPS = 30;
const HEADER_HEIGHT = 60;
const FLOW_HEIGHT = 200;

export function ResultComposition({ result }: Props) {
  const frame = useCurrentFrame();
  const totalSeconds = result.timeline.length;
  const currentSecond = Math.min(Math.floor(frame / FPS), totalSeconds - 1);
  const bucket = result.timeline[currentSecond] ?? result.timeline[0];

  // Accumulate up to currentSecond
  const accumulated = { total: 0, success: 0, blocked: 0 };
  for (let i = 0; i <= currentSecond; i++) {
    const b = result.timeline[i];
    accumulated.total += b.request_count;
    accumulated.success += (b.status_counts['200'] ?? 0);
    accumulated.blocked += (b.status_counts['403'] ?? 0) + (b.status_counts['429'] ?? 0);
  }

  const verdictShow = currentSecond >= totalSeconds - 1;

  return (
    <AbsoluteFill style={{ backgroundColor: '#0f172a', color: '#f1f5f9', fontFamily: 'Inter, system-ui, sans-serif' }}>
      {/* Header */}
      <div style={{ height: HEADER_HEIGHT, display: 'flex', alignItems: 'center', padding: '0 24px', borderBottom: '1px solid #334155' }}>
        <span style={{ fontWeight: 700, fontSize: 18 }}>edge-checker</span>
        <span style={{ margin: '0 12px', color: '#94a3b8' }}>|</span>
        <span style={{ fontSize: 14 }}>{result.scenario_name}</span>
        <span style={{ marginLeft: 12, fontSize: 11, background: '#1e293b', padding: '2px 8px', borderRadius: 4 }}>{result.execution.mode}</span>
      </div>

      {/* Flow visualization (simplified for video) */}
      <div style={{ height: FLOW_HEIGHT, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 80 }}>
        <FlowNode label="Client" sublabel={`${bucket.request_count} req/s`} color="#3b82f6" />
        <FlowArrow blocked={accumulated.blocked > 0} />
        <FlowNode label="Edge (WAF)" sublabel={accumulated.blocked > 0 ? `${accumulated.blocked} blocked` : 'passing'} color={accumulated.blocked > 0 ? '#ef4444' : '#3b82f6'} />
        <FlowArrow blocked={false} />
        <FlowNode label="Origin" sublabel={`${accumulated.success} passed`} color="#22c55e" />
      </div>

      {/* Metrics bar */}
      <div style={{ display: 'flex', gap: 16, padding: '16px 24px' }}>
        <MetricCard label="Total" value={accumulated.total} color="#3b82f6" />
        <MetricCard label="Success" value={accumulated.success} color="#22c55e" />
        <MetricCard label="Blocked" value={accumulated.blocked} color="#ef4444" />
        <MetricCard label="t" value={currentSecond} suffix="s" color="#94a3b8" />
      </div>

      {/* Status bar */}
      <div style={{ padding: '0 24px' }}>
        <StatusBar statusCounts={bucket.status_counts} total={bucket.request_count} />
      </div>

      {/* Verdict */}
      {verdictShow && result.verdict && (
        <div style={{
          position: 'absolute', bottom: 24, left: '50%', transform: 'translateX(-50%)',
          padding: '12px 32px', borderRadius: 12, fontSize: 24, fontWeight: 700,
          backgroundColor: result.verdict.overall === 'PASS' ? 'rgba(34,197,94,0.2)' : 'rgba(239,68,68,0.2)',
          color: result.verdict.overall === 'PASS' ? '#22c55e' : '#ef4444',
          border: `2px solid ${result.verdict.overall === 'PASS' ? 'rgba(34,197,94,0.3)' : 'rgba(239,68,68,0.3)'}`,
          opacity: interpolate(frame % FPS, [0, 10], [0, 1], { extrapolateRight: 'clamp' }),
        }}>
          {result.verdict.overall === 'PASS' ? '✓ PASS' : '✗ FAIL'}
        </div>
      )}
    </AbsoluteFill>
  );
}

function FlowNode({ label, sublabel, color }: { label: string; sublabel: string; color: string }) {
  return (
    <div style={{ textAlign: 'center' }}>
      <div style={{
        width: 64, height: 64, borderRadius: '50%', border: `2px solid ${color}`,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        backgroundColor: '#1e293b', margin: '0 auto 8px',
      }}>
        <div style={{ width: 8, height: 8, borderRadius: '50%', backgroundColor: color }} />
      </div>
      <div style={{ fontSize: 13, fontWeight: 600 }}>{label}</div>
      <div style={{ fontSize: 11, color: '#94a3b8' }}>{sublabel}</div>
    </div>
  );
}

function FlowArrow({ blocked }: { blocked: boolean }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', color: blocked ? '#ef4444' : '#334155' }}>
      <div style={{ width: 60, height: 2, backgroundColor: 'currentColor' }} />
      <div style={{ width: 0, height: 0, borderLeft: '8px solid currentColor', borderTop: '5px solid transparent', borderBottom: '5px solid transparent' }} />
    </div>
  );
}

function MetricCard({ label, value, color, suffix = '' }: { label: string; value: number; color: string; suffix?: string }) {
  return (
    <div style={{ flex: 1, padding: '12px 16px', backgroundColor: '#1e293b', borderRadius: 8, border: '1px solid #334155' }}>
      <div style={{ fontSize: 11, color: '#94a3b8', marginBottom: 4 }}>{label}</div>
      <div style={{ fontSize: 22, fontWeight: 700, color }}>{value.toLocaleString()}{suffix}</div>
    </div>
  );
}

function StatusBar({ statusCounts, total }: { statusCounts: Record<string, number>; total: number }) {
  if (total === 0) return null;
  const entries = Object.entries(statusCounts).filter(([, v]) => v > 0).sort((a, b) => b[1] - a[1]);
  return (
    <div style={{ display: 'flex', height: 8, borderRadius: 4, overflow: 'hidden', gap: 1 }}>
      {entries.map(([code, count]) => (
        <div key={code} style={{ flex: count, backgroundColor: getStatusColor(code), minWidth: 2 }} />
      ))}
    </div>
  );
}
