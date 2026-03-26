import { motion } from 'framer-motion';
import { AnimatedCounter } from '../common/AnimatedCounter';

interface Props {
  accumulated: {
    totalRequests: number;
    successCount: number;
    blockedCount: number;
    avgLatency: number;
  };
}

export function SummaryCards({ accumulated }: Props) {
  const cards = [
    { label: 'Total Requests', value: accumulated.totalRequests, color: 'var(--color-accent)' },
    { label: 'Success (2xx)', value: accumulated.successCount, color: 'var(--color-success)' },
    { label: 'Blocked', value: accumulated.blockedCount, color: 'var(--color-blocked)' },
    { label: 'Avg Latency', value: accumulated.avgLatency, color: 'var(--color-rate-limit)', format: (v: number) => `${v.toFixed(1)}ms` },
  ];

  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
      {cards.map((card, i) => (
        <motion.div
          key={card.label}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: i * 0.05 }}
          className="p-4 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)]"
        >
          <div className="text-xs text-[var(--color-text-muted)] mb-1">{card.label}</div>
          <div className="text-2xl font-bold" style={{ color: card.color }}>
            <AnimatedCounter value={card.value} format={card.format} />
          </div>
        </motion.div>
      ))}
    </div>
  );
}
