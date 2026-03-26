import { useState, useEffect, useRef } from 'react';
import type { SecondBucket } from '../../types/result';
import type { PacketData } from '../../utils/packetSampler';
import { samplePackets } from '../../utils/packetSampler';
import { Packet } from './Packet';

interface Props {
  bucket: SecondBucket;
  second: number;
  clientX: number;
  edgeX: number;
  originX: number;
  centerY: number;
}

export function PacketSpawner({ bucket, second, clientX, edgeX, originX, centerY }: Props) {
  const [activePackets, setActivePackets] = useState<PacketData[]>([]);
  const prevSecondRef = useRef(-1);

  useEffect(() => {
    if (second !== prevSecondRef.current) {
      prevSecondRef.current = second;
      const newPackets = samplePackets(bucket, second);
      setActivePackets((prev) => [...prev, ...newPackets]);
    }
  }, [bucket, second]);

  const handleComplete = (id: string) => {
    setActivePackets((prev) => prev.filter((p) => p.id !== id));
  };

  // Limit active packets to prevent memory issues
  const visiblePackets = activePackets.slice(-60);

  return (
    <>
      {visiblePackets.map((p, i) => (
        <Packet
          key={p.id}
          status={p.status}
          startX={clientX}
          y={centerY + p.yOffset}
          clientX={clientX}
          edgeX={edgeX}
          originX={originX}
          delay={i * 0.05}
          onComplete={() => handleComplete(p.id)}
        />
      ))}
    </>
  );
}
