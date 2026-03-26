import { motion, useMotionValue, useTransform } from 'framer-motion';

interface Props {
  x: number;
  y: number;
  blockRatio: number; // 0.0 to 1.0
}

export function ShieldEffect({ x, y, blockRatio }: Props) {
  const intensity = useMotionValue(blockRatio);
  const glowOpacity = useTransform(intensity, [0, 1], [0, 0.6]);
  const glowScale = useTransform(intensity, [0, 1], [1, 1.4]);

  return (
    <g>
      {/* Shield arc */}
      <motion.path
        d={`M ${x - 30} ${y - 50} Q ${x - 45} ${y} ${x - 30} ${y + 50}`}
        fill="none"
        stroke={blockRatio > 0.3 ? '#ef4444' : '#3b82f6'}
        strokeWidth={3}
        strokeLinecap="round"
        opacity={0.4 + blockRatio * 0.5}
      />

      {/* Glow effect when blocking */}
      {blockRatio > 0 && (
        <motion.circle
          cx={x - 30}
          cy={y}
          r={20}
          fill={blockRatio > 0.3 ? '#ef4444' : '#f97316'}
          style={{ opacity: glowOpacity, scale: glowScale }}
          animate={{
            opacity: [blockRatio * 0.2, blockRatio * 0.5, blockRatio * 0.2],
          }}
          transition={{ duration: 1.5, repeat: Infinity }}
        />
      )}
    </g>
  );
}
