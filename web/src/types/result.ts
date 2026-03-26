export interface Result {
  scenario_name: string;
  description: string;
  started_at: string;
  ended_at: string;
  duration: string;
  interrupted?: boolean;
  target: TargetInfo;
  execution: ExecutionInfo;
  summary: Summary;
  timeline: SecondBucket[];
  verdict?: Verdict;
}

export interface TargetInfo {
  base_url: string;
  path: string;
  method: string;
  profile: string;
}

export interface ExecutionInfo {
  mode: string;
  duration: string;
  concurrency: number;
  rps: number;
  environment: string;
}

export interface Summary {
  total_requests: number;
  success_count: number;
  error_count: number;
  status_counts: Record<string, number>;
  avg_latency_ms: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
}

export interface SecondBucket {
  second: number;
  request_count: number;
  status_counts: Record<string, number>;
  error_count: number;
  avg_latency_ms: number;
}

export type VerdictStatus = 'PASS' | 'FAIL' | 'SKIP';

export interface RuleResult {
  name: string;
  status: VerdictStatus;
  expected: string;
  actual: string;
  message: string;
}

export interface Verdict {
  overall: VerdictStatus;
  rules: RuleResult[];
}
