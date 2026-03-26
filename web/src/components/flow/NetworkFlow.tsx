import { motion } from 'framer-motion';
import type { SecondBucket } from '../../types/result';
import { NodeIcon } from './NodeIcon';
import { ConnectionPath } from './ConnectionPath';
import { ShieldEffect } from './ShieldEffect';
import { PacketSpawner } from './PacketSpawner';

interface Props {
  bucket: SecondBucket;
  second: number;
}

const CLIENT_X = 120;
const EDGE_X = 500;
const ORIGIN_X = 880;
const CENTER_Y = 140;

export function NetworkFlow({ bucket, second }: Props) {
  const totalInBucket = Object.values(bucket.status_counts).reduce((a, b) => a + b, 0) + bucket.error_count;
  const blockedInBucket = (bucket.status_counts['403'] ?? 0) + (bucket.status_counts['429'] ?? 0);
  const blockRatio = totalInBucket > 0 ? blockedInBucket / totalInBucket : 0;

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="w-full overflow-hidden"
      style={{ height: 280 }}
    >
      <svg viewBox="0 0 1000 280" className="w-full h-full" preserveAspectRatio="xMidYMid meet">
        {/* Connection paths */}
        <ConnectionPath x1={CLIENT_X} x2={EDGE_X} y={CENTER_Y} />
        <ConnectionPath x1={EDGE_X} x2={ORIGIN_X} y={CENTER_Y} />

        {/* Shield effect */}
        <ShieldEffect x={EDGE_X} y={CENTER_Y} blockRatio={blockRatio} />

        {/* Packets */}
        <PacketSpawner
          bucket={bucket}
          second={second}
          clientX={CLIENT_X + 40}
          edgeX={EDGE_X}
          originX={ORIGIN_X - 40}
          centerY={CENTER_Y}
        />

        {/* Nodes */}
        <NodeIcon
          type="client"
          x={CLIENT_X}
          y={CENTER_Y}
          label="Client"
          sublabel={`${bucket.request_count} req/s`}
        />
        <NodeIcon
          type="edge"
          x={EDGE_X}
          y={CENTER_Y}
          label="Edge (WAF)"
          sublabel={blockedInBucket > 0 ? `${blockedInBucket} blocked` : 'passing'}
          pulseIntensity={blockRatio}
        />
        <NodeIcon
          type="origin"
          x={ORIGIN_X}
          y={CENTER_Y}
          label="Origin"
          sublabel={`${bucket.status_counts['200'] ?? 0} passed`}
        />

        {/* Second counter */}
        <text x={500} y={268} textAnchor="middle" fill="var(--color-text-muted)" fontSize={12}>
          t = {second}s
        </text>
      </svg>
    </motion.div>
  );
}
