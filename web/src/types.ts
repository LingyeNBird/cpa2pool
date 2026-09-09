export interface Participant {
  id: string;
  name: string;
  note: string;
  key_scope: string;
  key_preview: string;
  enabled: boolean;
  paused: boolean;
  deleted: boolean;
  models: string[];
  efforts: string[];
  expires_at: string | null;
  created_at: string;
}
export interface Quota {
  id: string;
  participant_id: string;
  name: string;
  limit: string;
  period: string;
  starts_at: string;
  expires_at: string | null;
  enabled: boolean;
  anchor: string;
}
export interface Period {
  id: string;
  quota_id: string;
  participant_id: string;
  starts_at: string;
  ends_at: string | null;
  closed_at: string | null;
  limit: string;
  used: string;
  remaining: string;
}
export interface QuotaView extends Quota {
  current: Period;
  reason: string;
}
export interface Status {
  available: boolean;
  reasons: string[];
  remaining: string | null;
  used: string;
  quotas: QuotaView[];
}
export interface Price {
  model: string;
  input: string;
  output: string;
  cache_read: string;
  cache_write: string;
  priority_enabled: boolean;
  priority_multiplier: string;
  long_enabled: boolean;
  long_threshold: number;
  long_input_multiplier: string;
  long_output_multiplier: string;
  model_enabled: boolean;
  model_multiplier: string;
  combination: string;
  updated_at: string;
}
export interface Usage {
  input: number;
  output: number;
  cache_read: number;
  cache_write: number;
  reasoning: number;
}
export interface Bill {
  id: string;
  request_id: string;
  participant_id: string;
  model: string;
  requested_model: string;
  effort: string;
  service_tier: string;
  time: string;
  usage: Usage;
  charge: {
    base: string;
    final: string;
    factors: { name: string; input: string; output: string }[];
    price: Price;
  };
}
export interface Audit {
  id: string;
  participant_id: string;
  quota_id: string;
  action: string;
  time: string;
  before: unknown;
  after: unknown;
  note: string;
}
export interface Stat {
  group: string;
  requests: number;
  cost: string;
  input: number;
  output: number;
  cache_read: number;
}
export interface Page<T> {
  items: T[];
  total: number;
}
