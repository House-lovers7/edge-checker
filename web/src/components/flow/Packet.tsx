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
  const r = 5;

  if (status === 'pass') {
    return (
      <motion.circle
        cx={clientX}
        cy={y}
        r={r}
        fill={color}
        initial={{ cx: clientX, opacity: 0.8 }}
        animate={{ cx: originX, opacity: [0.8, 1, 0.8, 0] }}
        transition={{
          cx: { type: 'spring', stiffness: 30, damping: 12, delay },
          opacity: { duration: 2.5, delay, times: [0, 0.3, 0.8, 1] },
        }}
        onAnimationComplete={onComplete}
      />
    );
  }

  if (status === 'block' || status === 'rate-limit') {
    return (
      <>
        <motion.circle
          cx={clientX}
          cy={y}
          r={r}
          fill={color}
          initial={{ cx: clientX, opacity: 0.8 }}
          animate={{
            cx: [clientX, edgeX - 30, clientX + 50],
            opacity: [0.8, 1, 0],
          }}
          transition={{
            duration: 1.8,
            delay,
            times: [0, 0.4, 1],
            ease: ['easeOut', 'easeIn'],
          }}
          onAnimationComplete={onComplete}
        />
      </>
    );
  }

  // error
  return (
    <motion.circle
      cx={clientX}
      cy={y}
      r={r}
      fill={color}
      initial={{ cx: clientX, opacity: 0.8 }}
      animate={{
        cx: clientX + 100,
        opacity: [0.8, 1, 0.5, 0],
      }}
      transition={{
        duration: 1,
        delay,
      }}
      onAnimationComplete={onComplete}
    />
  );
}
