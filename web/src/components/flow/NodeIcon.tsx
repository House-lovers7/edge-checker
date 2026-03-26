import { motion } from 'framer-motion';

interface Props {
  type: 'client' | 'edge' | 'origin';
  x: number;
  y: number;
  label: string;
  sublabel?: string;
  pulseIntensity?: number;
}

export function NodeIcon({ type, x, y, label, sublabel, pulseIntensity = 0 }: Props) {
  const iconSize = 48;

  return (
    <g>
      {/* Pulse glow behind node */}
      {pulseIntensity > 0 && (
        <motion.circle
          cx={x}
          cy={y}
          r={iconSize}
          fill="none"
          stroke={type === 'edge' ? '#ef4444' : '#3b82f6'}
          strokeWidth={2}
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{
            opacity: [0, pulseIntensity * 0.4, 0],
            scale: [0.8, 1.3, 0.8],
          }}
          transition={{ duration: 1.5, repeat: Infinity }}
        />
      )}

      {/* Node background */}
      <motion.circle
        cx={x}
        cy={y}
        r={36}
        fill="var(--color-surface)"
        stroke="var(--color-border)"
        strokeWidth={2}
        whileHover={{ scale: 1.05 }}
      />

      {/* Icon */}
      {type === 'client' && <ClientIcon x={x} y={y} />}
      {type === 'edge' && <EdgeIcon x={x} y={y} />}
      {type === 'origin' && <OriginIcon x={x} y={y} />}

      {/* Label */}
      <text
        x={x}
        y={y + 55}
        textAnchor="middle"
        fill="var(--color-text)"
        fontSize={14}
        fontWeight={600}
      >
        {label}
      </text>
      {sublabel && (
        <text
          x={x}
          y={y + 72}
          textAnchor="middle"
          fill="var(--color-text-muted)"
          fontSize={11}
        >
          {sublabel}
        </text>
      )}
    </g>
  );
}

function ClientIcon({ x, y }: { x: number; y: number }) {
  return (
    <g transform={`translate(${x - 14}, ${y - 14})`}>
      <rect x={2} y={2} width={24} height={16} rx={2} fill="none" stroke="var(--color-text)" strokeWidth={1.5} />
      <line x1={8} y1={22} x2={20} y2={22} stroke="var(--color-text)" strokeWidth={1.5} />
      <line x1={14} y1={18} x2={14} y2={22} stroke="var(--color-text)" strokeWidth={1.5} />
    </g>
  );
}

function EdgeIcon({ x, y }: { x: number; y: number }) {
  // Shield shape
  return (
    <g transform={`translate(${x - 12}, ${y - 16})`}>
      <path
        d="M12 0L24 8V16C24 22 18 28 12 30C6 28 0 22 0 16V8L12 0Z"
        fill="none"
        stroke="#3b82f6"
        strokeWidth={1.5}
      />
      <path
        d="M8 14L11 17L17 11"
        fill="none"
        stroke="#3b82f6"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </g>
  );
}

function OriginIcon({ x, y }: { x: number; y: number }) {
  return (
    <g transform={`translate(${x - 10}, ${y - 14})`}>
      <rect x={0} y={0} width={20} height={28} rx={2} fill="none" stroke="var(--color-text)" strokeWidth={1.5} />
      <line x1={4} y1={6} x2={16} y2={6} stroke="var(--color-text)" strokeWidth={1} />
      <line x1={4} y1={12} x2={16} y2={12} stroke="var(--color-text)" strokeWidth={1} />
      <line x1={4} y1={18} x2={16} y2={18} stroke="var(--color-text)" strokeWidth={1} />
      <circle cx={10} cy={24} r={2} fill="var(--color-success)" />
    </g>
  );
}
