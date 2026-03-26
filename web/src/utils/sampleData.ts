import type { Result, SecondBucket } from '../types/result';

function generateTimeline(totalSeconds: number, baseRps: number): SecondBucket[] {
  const timeline: SecondBucket[] = [];
  for (let s = 0; s < totalSeconds; s++) {
    const blockStart = 8;
    const isBlocking = s >= blockStart;
    const blockRatio = isBlocking ? Math.min(0.6, (s - blockStart) * 0.08) : 0;
    const total = baseRps + Math.floor(Math.random() * 5 - 2);
    const blocked = Math.floor(total * blockRatio);
    const rateLimited = isBlocking ? Math.floor(total * blockRatio * 0.3) : 0;
    const success = total - blocked - rateLimited;
    const errorCount = Math.floor(Math.random() * 2);

    const statusCounts: Record<string, number> = {};
    if (success > 0) statusCounts['200'] = success;
    if (blocked > 0) statusCounts['403'] = blocked;
    if (rateLimited > 0) statusCounts['429'] = rateLimited;

    timeline.push({
      second: s,
      request_count: total,
      status_counts: statusCounts,
      error_count: errorCount,
      avg_latency_ms: 20 + Math.random() * 30 + (isBlocking ? 15 : 0),
    });
  }
  return timeline;
}

function summarizeTimeline(timeline: SecondBucket[]) {
  const statusCounts: Record<string, number> = {};
  let total = 0;
  let errors = 0;
  let latencySum = 0;
  const latencies: number[] = [];

  for (const b of timeline) {
    total += b.request_count;
    errors += b.error_count;
    latencySum += b.avg_latency_ms;
    latencies.push(b.avg_latency_ms);
    for (const [code, count] of Object.entries(b.status_counts)) {
      statusCounts[code] = (statusCounts[code] ?? 0) + count;
    }
  }

  latencies.sort((a, b) => a - b);
  const p = (pct: number) => latencies[Math.floor(latencies.length * pct)] ?? 0;
  const success = (statusCounts['200'] ?? 0) + (statusCounts['201'] ?? 0);

  return {
    total_requests: total,
    success_count: success,
    error_count: errors,
    status_counts: statusCounts,
    avg_latency_ms: latencySum / timeline.length,
    p50_latency_ms: p(0.5),
    p95_latency_ms: p(0.95),
    p99_latency_ms: p(0.99),
  };
}

const timeline = generateTimeline(30, 60);

export const sampleResult: Result = {
  scenario_name: 'assets-threshold-check',
  description: 'Verify Akamai rate policy triggers on sustained static asset access',
  started_at: '2026-03-27T10:30:00+09:00',
  ended_at: '2026-03-27T10:30:30+09:00',
  duration: '30s',
  target: {
    base_url: 'https://staging.example.com',
    path: '/assets/images/medium/sample.jpg',
    method: 'GET',
    profile: 'bot-like',
  },
  execution: {
    mode: 'constant',
    duration: '30s',
    concurrency: 5,
    rps: 60,
    environment: 'staging',
  },
  summary: summarizeTimeline(timeline),
  timeline,
  verdict: {
    overall: 'PASS',
    rules: [
      {
        name: 'Block detection',
        status: 'PASS',
        expected: 'Block status [403 429] within 60s',
        actual: 'First block at 8s',
        message: 'Defense triggered within expected window (at 8s)',
      },
      {
        name: 'Block ratio',
        status: 'PASS',
        expected: '>= 20%',
        actual: '32.5% (585/1800)',
        message: 'Block ratio meets the expected minimum',
      },
    ],
  },
};
