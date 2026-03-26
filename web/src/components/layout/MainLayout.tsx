import type { Result } from '../../types/result';
import type { TimelinePlayer } from '../../hooks/useTimelinePlayer';
import { NetworkFlow } from '../flow/NetworkFlow';
import { PlaybackControls } from '../controls/PlaybackControls';
import { SummaryCards } from '../dashboard/SummaryCards';
import { StatusDonut } from '../dashboard/StatusDonut';
import { TimelineChart } from '../dashboard/TimelineChart';
import { VerdictBadge } from '../dashboard/VerdictBadge';

interface Props {
  result: Result;
  player: TimelinePlayer;
}

export function MainLayout({ result, player }: Props) {
  const currentBucket = result.timeline[player.currentSecond] ?? result.timeline[0];

  // Accumulate stats up to currentSecond
  const accumulated = accumulateUntil(result.timeline, player.currentSecond);

  return (
    <div className="flex-1 flex flex-col">
      {/* Network Flow - Hero */}
      <div className="border-b border-[var(--color-border)]">
        <NetworkFlow
          bucket={currentBucket}
          second={player.currentSecond}
        />
        <PlaybackControls player={player} />
      </div>

      {/* Dashboard */}
      <div className="flex-1 grid grid-cols-1 lg:grid-cols-3 gap-4 p-4">
        <div className="lg:col-span-2 space-y-4">
          <SummaryCards accumulated={accumulated} />
          <TimelineChart
            timeline={result.timeline}
            currentSecond={player.currentSecond}
          />
        </div>
        <div className="space-y-4">
          <StatusDonut statusCounts={accumulated.statusCounts} />
          {result.verdict && (
            <VerdictBadge
              verdict={result.verdict}
              show={player.isFinished}
            />
          )}
        </div>
      </div>
    </div>
  );
}

interface AccumulatedStats {
  totalRequests: number;
  successCount: number;
  blockedCount: number;
  errorCount: number;
  statusCounts: Record<string, number>;
  avgLatency: number;
}

function accumulateUntil(timeline: Result['timeline'], untilSecond: number): AccumulatedStats {
  const statusCounts: Record<string, number> = {};
  let total = 0;
  let errors = 0;
  let latencySum = 0;
  let count = 0;

  for (let i = 0; i <= untilSecond && i < timeline.length; i++) {
    const b = timeline[i];
    total += b.request_count;
    errors += b.error_count;
    latencySum += b.avg_latency_ms;
    count++;
    for (const [code, n] of Object.entries(b.status_counts)) {
      statusCounts[code] = (statusCounts[code] ?? 0) + n;
    }
  }

  const success = Object.entries(statusCounts)
    .filter(([c]) => parseInt(c) >= 200 && parseInt(c) < 300)
    .reduce((sum, [, n]) => sum + n, 0);

  const blocked = (statusCounts['403'] ?? 0) + (statusCounts['429'] ?? 0);

  return {
    totalRequests: total,
    successCount: success,
    blockedCount: blocked,
    errorCount: errors,
    statusCounts,
    avgLatency: count > 0 ? latencySum / count : 0,
  };
}
