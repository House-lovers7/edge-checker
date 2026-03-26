import { AnimatedCounter } from '../common/AnimatedCounter';

interface Props {
  p50: number;
  p95: number;
  p99: number;
}

export function LatencyDisplay({ p50, p95, p99 }: Props) {
  const items = [
    { label: 'p50', value: p50 },
    { label: 'p95', value: p95 },
    { label: 'p99', value: p99 },
  ];

  return (
    <div className="p-4 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)]">
      <div className="text-xs text-[var(--color-text-muted)] mb-3">Latency Percentiles</div>
      <div className="flex gap-4 justify-around">
        {items.map((item) => (
          <div key={item.label} className="text-center">
            <div className="text-lg font-bold text-purple-400">
              <AnimatedCounter value={item.value} format={(v) => `${v.toFixed(0)}`} />
              <span className="text-xs font-normal text-[var(--color-text-muted)]">ms</span>
            </div>
            <div className="text-[10px] text-[var(--color-text-muted)] uppercase">{item.label}</div>
          </div>
        ))}
      </div>
    </div>
  );
}
