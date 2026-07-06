import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

const errorRate = new Rate("errors");
const ingestDuration = new Trend("ingest_duration");

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
let TOKEN = "";

export const options = {
  scenarios: {
    sustained_load: {
      executor: "constant-arrival-rate",
      rate: 100,
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 20,
      maxVUs: 50,
    },
    spike: {
      executor: "ramping-arrival-rate",
      startRate: 10,
      timeUnit: "1s",
      stages: [
        { target: 10, duration: "30s" },
        { target: 500, duration: "10s" },
        { target: 500, duration: "30s" },
        { target: 10, duration: "10s" },
        { target: 10, duration: "30s" },
      ],
      preAllocatedVUs: 50,
      maxVUs: 200,
      startTime: "2m",
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<200", "p(99)<500"],
    errors: ["rate<0.01"],
    ingest_duration: ["p(95)<150"],
  },
};

export function setup() {
  // Register and login
  const email = `loadtest-${Date.now()}@atlas.dev`;
  http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
    email,
    password: "loadtest123!",
  }), { headers: { "Content-Type": "application/json" } });

  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email,
    password: "loadtest123!",
  }), { headers: { "Content-Type": "application/json" } });

  const body = JSON.parse(loginRes.body);
  return { token: body.access_token };
}

const SOURCES = [
  "api-gateway", "user-service", "payment-service",
  "order-service", "inventory-service", "notification-service",
];
const EVENT_TYPES = [
  "http_request", "db_query", "cache_miss", "authentication",
  "deployment", "error", "metric_report",
];
const SEVERITIES = ["debug", "info", "info", "info", "warn", "error"];

function randomEvent() {
  return {
    source: SOURCES[Math.floor(Math.random() * SOURCES.length)],
    event_type: EVENT_TYPES[Math.floor(Math.random() * EVENT_TYPES.length)],
    severity: SEVERITIES[Math.floor(Math.random() * SEVERITIES.length)],
    data: {
      method: "POST",
      path: `/api/v1/orders/${Math.floor(Math.random() * 10000)}`,
      status_code: Math.random() > 0.95 ? 500 : 200,
      duration_ms: Math.random() * 500,
      user_id: `user_${Math.floor(Math.random() * 1000)}`,
    },
    tags: ["load-test", `run-${Date.now()}`],
  };
}

export default function (data) {
  const batchSize = Math.floor(Math.random() * 9) + 2; // 2-10 events
  const events = Array.from({ length: batchSize }, randomEvent);

  const res = http.post(`${BASE_URL}/api/v1/events`, JSON.stringify({ events }), {
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${data.token}`,
    },
  });

  check(res, {
    "status is 202": (r) => r.status === 202,
    "has accepted count": (r) => {
      const body = JSON.parse(r.body);
      return body.accepted > 0;
    },
  });

  errorRate.add(res.status !== 202);
  ingestDuration.add(res.timings.duration);

  sleep(0.01);
}
