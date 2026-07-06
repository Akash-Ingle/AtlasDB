const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface RequestOptions {
  method?: string;
  body?: unknown;
  token?: string | null;
}

export class APIError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string
  ) {
    super(message);
  }
}

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  const token = opts.token ?? getStoredToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method: opts.method || "GET",
    headers,
    body: opts.body ? JSON.stringify(opts.body) : undefined,
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: { code: "unknown", message: res.statusText } }));
    throw new APIError(res.status, err.error?.code || "unknown", err.error?.message || res.statusText);
  }

  if (res.status === 204) return {} as T;
  return res.json();
}

function getStoredToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("atlas_token");
}

// ── Auth ──────────────────────────────────────────────────────────
export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export function register(email: string, password: string) {
  return request<TokenPair>("/api/v1/auth/register", {
    method: "POST",
    body: { email, password },
    token: null,
  });
}

export function login(email: string, password: string) {
  return request<TokenPair>("/api/v1/auth/login", {
    method: "POST",
    body: { email, password },
    token: null,
  });
}

// ── Events ────────────────────────────────────────────────────────
export interface Event {
  event_id: string;
  source: string;
  event_type: string;
  severity: "debug" | "info" | "warn" | "error" | "fatal";
  timestamp: string;
  received_at: string;
  data: Record<string, unknown>;
  tags: string[];
  metadata?: Record<string, unknown>;
}

export interface PaginatedResponse<T> {
  data: T[];
  cursor?: string;
  has_more: boolean;
}

export interface IngestResponse {
  accepted: number;
  event_ids: string[];
}

export function getEvents(params?: {
  source?: string;
  severity?: string;
  start_time?: string;
  end_time?: string;
  limit?: number;
  cursor?: string;
}) {
  const qs = new URLSearchParams();
  if (params) {
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== "") qs.set(k, String(v));
    });
  }
  const query = qs.toString();
  return request<PaginatedResponse<Event>>(`/api/v1/events${query ? "?" + query : ""}`);
}

export function getEvent(id: string) {
  return request<Event>(`/api/v1/events/${id}`);
}

export function ingestEvents(events: Partial<Event>[]) {
  return request<IngestResponse>("/api/v1/events", {
    method: "POST",
    body: { events },
  });
}

// ── Search ────────────────────────────────────────────────────────
export interface SearchResult {
  event: Event;
  rank: number;
}

export function searchEvents(params: {
  q: string;
  source?: string;
  severity?: string;
  start_time?: string;
  end_time?: string;
  limit?: number;
  cursor?: string;
}) {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== "") qs.set(k, String(v));
  });
  return request<PaginatedResponse<SearchResult>>(`/api/v1/search?${qs.toString()}`);
}

// ── Analytics ─────────────────────────────────────────────────────
export interface TimeSeriesPoint {
  bucket: string;
  count: number;
  source?: string;
  severity?: string;
}

export interface AnalyticsSummary {
  total_events: number;
  error_count: number;
  error_rate: number;
  active_sources: number;
  top_sources: { source: string; count: number }[];
  by_severity: { severity: string; count: number }[];
}

export interface TopNResult {
  key: string;
  count: number;
}

export function getAnalyticsSummary(range_: string = "1h") {
  return request<AnalyticsSummary>(`/api/v1/analytics/summary?range=${range_}`);
}

export function getTimeSeries(params: {
  range?: string;
  resolution?: string;
  group_by?: string;
  source?: string;
}) {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v) qs.set(k, v);
  });
  return request<{ data: TimeSeriesPoint[]; resolution: string }>(
    `/api/v1/analytics/timeseries?${qs.toString()}`
  );
}

export function getTopN(params: { range?: string; group_by?: string; limit?: number }) {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined) qs.set(k, String(v));
  });
  return request<{ data: TopNResult[] }>(`/api/v1/analytics/top?${qs.toString()}`);
}

// ── Alerts ────────────────────────────────────────────────────────
export interface AlertCondition {
  metric: string;
  operator: string;
  threshold: number;
  window: string;
  source?: string;
}

export interface AlertRule {
  rule_id: string;
  name: string;
  description?: string;
  condition: AlertCondition;
  severity: string;
  channels: unknown[];
  enabled: boolean;
  cooldown_seconds: number;
  created_at: string;
  updated_at: string;
}

export interface AlertEvent {
  alert_event_id: string;
  rule_id: string;
  rule_name?: string;
  status: string;
  fired_at: string;
  resolved_at?: string;
  value: number;
  context?: Record<string, unknown>;
  notified: boolean;
}

export function getAlertRules() {
  return request<{ rules: AlertRule[] }>("/api/v1/alerts/rules");
}

export function createAlertRule(rule: {
  name: string;
  description?: string;
  condition: AlertCondition;
  severity?: string;
  channels?: unknown[];
  cooldown_seconds?: number;
}) {
  return request<AlertRule>("/api/v1/alerts/rules", { method: "POST", body: rule });
}

export function updateAlertRule(id: string, rule: Partial<AlertRule>) {
  return request<AlertRule>(`/api/v1/alerts/rules/${id}`, { method: "PUT", body: rule });
}

export function deleteAlertRule(id: string) {
  return request<void>(`/api/v1/alerts/rules/${id}`, { method: "DELETE" });
}

export function getAlertHistory(limit: number = 50) {
  return request<{ alerts: AlertEvent[] }>(`/api/v1/alerts/history?limit=${limit}`);
}

// ── System ────────────────────────────────────────────────────────
export function getHealth() {
  return request<{ status: string; checks: Record<string, string> }>("/readyz", { token: null });
}

// ── WebSocket ─────────────────────────────────────────────────────
export function createEventStream(onEvent: (event: Event) => void): WebSocket {
  const wsBase = API_BASE.replace(/^http/, "ws");
  const token = getStoredToken();
  const ws = new WebSocket(`${wsBase}/api/v1/events/stream?token=${token}`);

  ws.onmessage = (msg) => {
    try {
      const event = JSON.parse(msg.data) as Event;
      onEvent(event);
    } catch {
      // ignore malformed messages
    }
  };

  return ws;
}
