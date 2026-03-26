import { motion } from 'framer-motion';
import type { Result } from '../../types/result';

interface Props {
  result: Result;
  onReset: () => void;
}

export function Header({ result, onReset }: Props) {
  return (
    <motion.header
      initial={{ opacity: 0, y: -10 }}
      animate={{ opacity: 1, y: 0 }}
      className="flex items-center justify-between px-6 py-3 border-b border-[var(--color-border)]"
    >
      <div className="flex items-center gap-4">
        <h1 className="text-lg font-bold">edge-checker</h1>
        <span className="text-[var(--color-text-muted)] text-sm">|</span>
        <span className="text-sm font-medium">{result.scenario_name}</span>
        <span className="text-xs px-2 py-0.5 rounded bg-[var(--color-surface)] text-[var(--color-text-muted)]">
          {result.execution.mode}
        </span>
        <span className="text-xs text-[var(--color-text-muted)]">
          {result.target.method} {result.target.base_url}{result.target.path}
        </span>
      </div>
      <button
        onClick={onReset}
        className="text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
      >
        Load another
      </button>
    </motion.header>
  );
}
