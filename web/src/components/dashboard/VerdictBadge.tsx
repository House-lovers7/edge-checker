import { motion, AnimatePresence } from 'framer-motion';
import type { Verdict } from '../../types/result';

interface Props {
  verdict: Verdict;
  show: boolean;
}

export function VerdictBadge({ verdict, show }: Props) {
  const isPASS = verdict.overall === 'PASS';

  return (
    <div className="p-4 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)]">
      <div className="text-xs text-[var(--color-text-muted)] mb-3">Verdict</div>

      <AnimatePresence>
        {show ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.5 }}
            animate={{
              opacity: 1,
              scale: 1,
              x: isPASS ? 0 : [0, -4, 4, -4, 4, 0],
            }}
            transition={{
              type: 'spring',
              stiffness: 200,
              damping: 10,
              x: { duration: 0.5, delay: 0.3 },
            }}
          >
            {/* Badge */}
            <div
              className={`text-center py-3 px-4 rounded-lg text-xl font-bold ${
                isPASS
                  ? 'bg-green-500/20 text-green-400 border border-green-500/30'
                  : 'bg-red-500/20 text-red-400 border border-red-500/30'
              }`}
            >
              {isPASS ? '✓ PASS' : '✗ FAIL'}
            </div>

            {/* Rules */}
            <div className="mt-3 space-y-2">
              {verdict.rules.map((rule) => (
                <div key={rule.name} className="flex items-start gap-2 text-xs">
                  <span className={
                    rule.status === 'PASS' ? 'text-green-400' :
                    rule.status === 'FAIL' ? 'text-red-400' : 'text-gray-400'
                  }>
                    {rule.status === 'PASS' ? '✓' : rule.status === 'FAIL' ? '✗' : '-'}
                  </span>
                  <div>
                    <div className="font-medium">{rule.name}</div>
                    {rule.actual && (
                      <div className="text-[var(--color-text-muted)]">{rule.actual}</div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </motion.div>
        ) : (
          <div className="text-center py-6 text-[var(--color-text-muted)] text-sm">
            Play to end to see verdict
          </div>
        )}
      </AnimatePresence>
    </div>
  );
}
