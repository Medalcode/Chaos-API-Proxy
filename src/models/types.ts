export interface ChaosRules {
  latency_ms?: number;
  jitter?: number;
  inject_failure_rate?: number;
  error_code?: number;
  error_body?: string;
  drop_connection?: boolean;
  bandwidth_limit_kbps?: number;
  response_fuzzing?: {
    enabled: boolean;
    probability: number;
    mutation_rate: number;
  };
  modify_headers?: Record<string, string>;
  remove_headers?: string[];
}

export interface ChaosConfig {
  id: string;
  name: string;
  description?: string;
  target: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
  rules: ChaosRules;
}

export interface RequestLog {
  id: string;
  timestamp: string;
  config_id: string;
  method: string;
  path: string;
  status_code: number;
  duration_ms: number;
  chaos_type: string;
}
