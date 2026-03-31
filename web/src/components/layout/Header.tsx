import { motion } from 'framer-motion';
import type { Result } from '../../types/result';

interface Props {
  result: Result;
  onReset: () => void;
  onVideoPreview: () => void;
}

export function Header({ result, onReset, onVideoPreview }: Props) {
  return (
    <motion.header
      initial={{ opacity: 0, y: -10 }}
      animate={{ opacity: 1, y: 0 }}
      className="flex items-center justify-between px-6 py-3 border-b border-[var(--color-border)]"
    >
      <div className="flex items-center gap-4 min-w-0">
        <h1 className="text-lg font-bold shrink-0">edge-checker</h1>
        <span className="text-[var(--color-text-muted)] text-sm shrink-0">|</span>
        <span className="text-sm font-medium truncate">{result.scenario_name}</span>
        <span className="text-xs px-2 py-0.5 rounded bg-[var(--color-surface)] text-[var(--color-text-muted)] shrink-0">
          {result.execution.mode}
        </span>
        <span className="text-xs text-[var(--color-text-muted)] truncate hidden md:inline">
          {result.target.method} {result.target.base_url}{result.target.path}
        </span>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        <button
          onClick={onVideoPreview}
          className="text-sm text-[var(--color-accent)] hover:text-white transition-colors"
          title="Video preview (Remotion)"
        >
          Video
        </button>
        <button
          onClick={onReset}
          className="text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
        >
          Load another
        </button>
      </div>
    </motion.header>
  );
}
