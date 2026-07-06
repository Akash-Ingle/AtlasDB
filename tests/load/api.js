import http from "k6/http";
import { check, group, sleep } from "k6";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

export const options = {
  stages: [
    { duration: "30s", target: 10 },
    { duration: "1m", target: 30 },
    { duration: "30s", target: 50 },
    { duration: "1m", target: 50 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<300"],
    "http_req_duration{name:list_events}": ["p(95)<200"],
    "http_req_duration{name:search}": ["p(95)<300"],
    "http_req_duration{name:analytics}": ["p(95)<250"],
  },
};

export function setup() {
  const email = `apitest-${Date.now()}@atlas.dev`;
  http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
    email,
    password: "apitest123!",
  }), { headers: { "Content-Type": "application/json" } });

  const res = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email,
    password: "apitest123!",
  }), { headers: { "Content-Type": "application/json" } });

  return { token: JSON.parse(res.body).access_token };
}

const headers = (token) => ({
  "Content-Type": "application/json",
  Authorization: `Bearer ${token}`,
});

export default function (data) {
  const h = headers(data.token);

  group("List Events", () => {
    const res = http.get(`${BASE_URL}/api/v1/events?limit=20`, {
      headers: h,
      tags: { name: "list_events" },
    });
    check(res, { "events 200": (r) => r.status === 200 });
  });

  group("Search", () => {
    const queries = ["error", "timeout", "payment", "login"];
    const q = queries[Math.floor(Math.random() * queries.length)];
    const res = http.get(`${BASE_URL}/api/v1/search?q=${q}&limit=10`, {
      headers: h,
      tags: { name: "search" },
    });
    check(res, { "search 200": (r) => r.status === 200 });
  });

  group("Analytics", () => {
    const res = http.get(`${BASE_URL}/api/v1/analytics/summary?range=1h`, {
      headers: h,
      tags: { name: "analytics" },
    });
    check(res, { "analytics 200": (r) => r.status === 200 });
  });

  group("Timeseries", () => {
    const res = http.get(
      `${BASE_URL}/api/v1/analytics/timeseries?range=1h&group_by=severity`,
      { headers: h, tags: { name: "timeseries" } }
    );
    check(res, { "timeseries 200": (r) => r.status === 200 });
  });

  group("Health", () => {
    const res = http.get(`${BASE_URL}/readyz`);
    check(res, { "health 200": (r) => r.status === 200 });
  });

  sleep(0.5);
}
