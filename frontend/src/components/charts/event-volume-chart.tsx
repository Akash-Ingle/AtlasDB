"use client";

import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import type { Event } from "@/lib/api";

interface Props {
  events: Event[];
}

export function EventVolumeChart({ events }: Props) {
  // Group events into 5-minute buckets
  const buckets = new Map<string, { info: number; warn: number; error: number }>();

  const now = Date.now();
  // Create empty buckets for the last hour
  for (let i = 12; i >= 0; i--) {
    const t = new Date(now - i * 5 * 60 * 1000);
    const key = `${t.getHours().toString().padStart(2, "0")}:${(Math.floor(t.getMinutes() / 5) * 5).toString().padStart(2, "0")}`;
    buckets.set(key, { info: 0, warn: 0, error: 0 });
  }

  for (const event of events) {
    const d = new Date(event.timestamp);
    const key = `${d.getHours().toString().padStart(2, "0")}:${(Math.floor(d.getMinutes() / 5) * 5).toString().padStart(2, "0")}`;
    const bucket = buckets.get(key);
    if (bucket) {
      if (event.severity === "error" || event.severity === "fatal") {
        bucket.error++;
      } else if (event.severity === "warn") {
        bucket.warn++;
      } else {
        bucket.info++;
      }
    }
  }

  const data = Array.from(buckets.entries()).map(([time, counts]) => ({
    time,
    ...counts,
  }));

  return (
    <ResponsiveContainer width="100%" height={240}>
      <AreaChart data={data}>
        <defs>
          <linearGradient id="infoGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#3b82f6" stopOpacity={0.3} />
            <stop offset="100%" stopColor="#3b82f6" stopOpacity={0} />
          </linearGradient>
          <linearGradient id="warnGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#f59e0b" stopOpacity={0.3} />
            <stop offset="100%" stopColor="#f59e0b" stopOpacity={0} />
          </linearGradient>
          <linearGradient id="errorGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#ef4444" stopOpacity={0.3} />
            <stop offset="100%" stopColor="#ef4444" stopOpacity={0} />
          </linearGradient>
        </defs>
        <XAxis
          dataKey="time"
          tick={{ fill: "#71717a", fontSize: 11 }}
          axisLine={{ stroke: "#27272a" }}
          tickLine={false}
        />
        <YAxis
          tick={{ fill: "#71717a", fontSize: 11 }}
          axisLine={false}
          tickLine={false}
          width={35}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: "#18181b",
            border: "1px solid #27272a",
            borderRadius: "6px",
            fontSize: "12px",
          }}
          labelStyle={{ color: "#a1a1aa" }}
        />
        <Area
          type="monotone"
          dataKey="info"
          stroke="#3b82f6"
          fill="url(#infoGrad)"
          strokeWidth={1.5}
          stackId="1"
        />
        <Area
          type="monotone"
          dataKey="warn"
          stroke="#f59e0b"
          fill="url(#warnGrad)"
          strokeWidth={1.5}
          stackId="1"
        />
        <Area
          type="monotone"
          dataKey="error"
          stroke="#ef4444"
          fill="url(#errorGrad)"
          strokeWidth={1.5}
          stackId="1"
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
