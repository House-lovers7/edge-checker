import {
  ComposedChart,
  Bar,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ReferenceLine,
  ResponsiveContainer,
} from 'recharts';
import type { SecondBucket } from '../../types/result';

interface Props {
  timeline: SecondBucket[];
  currentSecond: number;
}

export function TimelineChart({ timeline, currentSecond }: Props) {
  const data = timeline.map((b) => ({
    second: b.second,
    success: (b.status_counts['200'] ?? 0) + (b.status_counts['201'] ?? 0),
    blocked403: b.status_counts['403'] ?? 0,
    blocked429: b.status_counts['429'] ?? 0,
    errors: b.error_count,
    latency: Math.round(b.avg_latency_ms),
  }));

  return (
    <div className="p-4 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)]">
      <div className="text-xs text-[var(--color-text-muted)] mb-3">Timeline (per second)</div>
      <div className="h-52">
        <ResponsiveContainer width="100%" height="100%">
          <ComposedChart data={data} margin={{ top: 5, right: 10, bottom: 5, left: 0 }}>
            <XAxis
              dataKey="second"
              tick={{ fontSize: 10, fill: '#94a3b8' }}
              tickFormatter={(v) => `${v}s`}
              interval="preserveStartEnd"
            />
            <YAxis
              yAxisId="left"
              tick={{ fontSize: 10, fill: '#94a3b8' }}
              width={35}
            />
            <YAxis
              yAxisId="right"
              orientation="right"
              tick={{ fontSize: 10, fill: '#94a3b8' }}
              width={35}
              tickFormatter={(v) => `${v}ms`}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: '#1e293b',
                border: '1px solid #334155',
                borderRadius: '8px',
                fontSize: '12px',
              }}
              labelFormatter={(v) => `${v}s`}
            />
            <Bar yAxisId="left" dataKey="success" stackId="a" fill="#22c55e" radius={[0, 0, 0, 0]} />
            <Bar yAxisId="left" dataKey="blocked403" stackId="a" fill="#ef4444" />
            <Bar yAxisId="left" dataKey="blocked429" stackId="a" fill="#f97316" />
            <Bar yAxisId="left" dataKey="errors" stackId="a" fill="#6b7280" radius={[2, 2, 0, 0]} />
            <Line
              yAxisId="right"
              dataKey="latency"
              stroke="#a855f7"
              strokeWidth={2}
              dot={false}
              name="Latency (ms)"
            />
            <ReferenceLine
              yAxisId="left"
              x={currentSecond}
              stroke="#3b82f6"
              strokeWidth={2}
              strokeDasharray="4 2"
            />
          </ComposedChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
