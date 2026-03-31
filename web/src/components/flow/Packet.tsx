import { motion } from 'framer-motion';
import type { PacketStatus } from '../../utils/packetSampler';
import { statusToColor } from '../../utils/packetSampler';

interface Props {
  status: PacketStatus;
  startX: number;
  y: number;
  clientX: number;
  edgeX: number;
  originX: number;
  delay: number;
  onComplete: () => void;
}

export function Packet({ status, y, clientX, edgeX, originX, delay, onComplete }: Props) {
  const color = statusToColor(status);

  if (status === 'pass') {
    return (
      <motion.circle
        cx={clientX}
        cy={y}
        r={4}
        fill={color}
        initial={{ cx: clientX, opacity: 0, scale: 0.5 }}
        animate={{
          cx: originX,
          cy: [y, y + (Math.random() - 0.5) * 10, y],
          opacity: [0, 0.9, 0.9, 0],
          scale: [0.5, 1, 1, 0.3],
        }}
        transition={{
          cx: { type: 'spring', stiffness: 25, damping: 10, delay },
          cy: { duration: 2.5, delay, ease: 'easeInOut' },
          opacity: { duration: 2.8, delay, times: [0, 0.1, 0.75, 1] },
          scale: { duration: 2.8, delay, times: [0, 0.1, 0.75, 1] },
        }}
        onAnimationComplete={onComplete}
      />
    );
  }

  if (status === 'block' || status === 'rate-limit') {
    return (
      <g>
        {/* Impact flash at shield */}
        <motion.circle
          cx={edgeX - 30}
          cy={y}
          r={8}
          fill={color}
          initial={{ opacity: 0, scale: 0 }}
          animate={{ opacity: [0, 0.6, 0], scale: [0, 1.5, 0] }}
          transition={{ duration: 0.4, delay: delay + 0.6 }}
        />
        {/* Packet */}
        <motion.circle
          cx={clientX}
          cy={y}
          r={4}
          fill={color}
          initial={{ cx: clientX, opacity: 0, scale: 0.5 }}
          animate={{
            cx: [clientX, edgeX - 35, clientX + 40],
            opacity: [0, 0.9, 1, 0],
            scale: [0.5, 1, 1.3, 0],
          }}
          transition={{
            duration: 1.6,
            delay,
            times: [0, 0.4, 0.5, 1],
            ease: ['easeOut', 'easeOut', 'easeIn'],
          }}
          onAnimationComplete={onComplete}
        />
      </g>
    );
  }

  // error — flicker and fade
  return (
    <motion.circle
      cx={clientX}
      cy={y}
      r={4}
      fill={color}
      initial={{ cx: clientX, opacity: 0 }}
      animate={{
        cx: clientX + 80,
        opacity: [0, 0.8, 0.3, 0.7, 0],
        scale: [0.5, 1, 0.8, 1, 0],
      }}
      transition={{
        duration: 0.8,
        delay,
      }}
      onAnimationComplete={onComplete}
    />
  );
}
