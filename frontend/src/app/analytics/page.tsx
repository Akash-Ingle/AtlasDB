"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  AreaChart, Area, BarChart, Bar,
  XAxis, YAxis, Tooltip, ResponsiveContainer, Cell,
} from "recharts";
import { AuthGuard } from "@/components/layout/auth-guard";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { Stat } from "@/components/ui/stat";
import { getAnalyticsSummary, getTimeSeries, getTopN } from "@/lib/api";

const RANGES = ["15m", "1h", "6h", "24h", "7d"] as const;
const BAR_COLORS = ["#3b82f6", "#8b5cf6", "#06b6d4", "#10b981", "#f59e0b", "#ef4444", "#ec4899", "#6366f1"];

export default function AnalyticsPage() {
  const [range_, setRange] = useState<string>("1h");

  const { data: summary } = useQuery({
    queryKey: ["analytics-summary", range_],
    queryFn: () => getAnalyticsSummary(range_),
    refetchInterval: 10000,
  });

  const { data: timeseries } = useQuery({
    queryKey: ["analytics-timeseries", range_],
    queryFn: () => getTimeSeries({ range: range_, group_by: "severity" }),
    refetchInterval: 10000,
  });

  const { data: topSources } = useQuery({
    queryKey: ["analytics-top-sources", range_],
    queryFn: () => getTopN({ range: range_, group_by: "source", limit: 8 }),
    refetchInterval: 10000,
  });

  const { data: topTypes } = useQuery({
    queryKey: ["analytics-top-types", range_],
    queryFn: () => getTopN({ range: range_, group_by: "event_type", limit: 8 }),
    refetchInterval: 10000,
  });

  const tsData = (timeseries?.data || []).map((p) => ({
    time: new Date(p.bucket).toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    }),
    count: p.count,
    severity: p.severity || "info",
  }));

  // Pivot time-series data by severity
  const pivoted = new Map<string, Record<string, number>>();
  for (const p of tsData) {
    const existing = pivoted.get(p.time) || {};
    existing[p.severity] = (existing[p.severity] || 0) + p.count;
    pivoted.set(p.time, existing);
  }
  const chartData = Array.from(pivoted.entries()).map(([time, counts]) => ({
    time,
    ...counts,
  }));

  const errorRate = summary ? (summary.error_rate * 100).toFixed(1) : "0";

  return (
    <AuthGuard>
      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold text-white">Analytics</h1>
            <p className="text-sm text-zinc-500 mt-1">
              Event metrics and trends
            </p>
          </div>

          {/* Time range selector */}
          <div className="flex gap-1 bg-zinc-900 rounded-md p-1">
            {RANGES.map((r) => (
              <button
                key={r}
                onClick={() => setRange(r)}
                className={`px-3 py-1.5 text-xs rounded transition-colors ${
                  range_ === r
                    ? "bg-zinc-700 text-white"
                    : "text-zinc-500 hover:text-zinc-300"
                }`}
              >
                {r}
              </button>
            ))}
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-4 gap-4">
          <Stat label="Total Events" value={summary?.total_events ?? 0} />
          <Stat
            label="Error Rate"
            value={`${errorRate}%`}
            changeType={Number(errorRate) > 5 ? "negative" : "positive"}
          />
          <Stat label="Error Count" value={summary?.error_count ?? 0}
            changeType={(summary?.error_count ?? 0) > 0 ? "negative" : "positive"}
          />
          <Stat label="Active Sources" value={summary?.active_sources ?? 0} />
        </div>

        {/* Time-series chart */}
        <Card>
          <CardHeader>
            <CardTitle>Event Volume by Severity</CardTitle>
          </CardHeader>
          <ResponsiveContainer width="100%" height={280}>
            <AreaChart data={chartData}>
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
                width={40}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: "#18181b",
                  border: "1px solid #27272a",
                  borderRadius: "6px",
                  fontSize: "12px",
                }}
              />
              <Area type="monotone" dataKey="info" stroke="#3b82f6" fill="#3b82f6" fillOpacity={0.15} stackId="1" />
              <Area type="monotone" dataKey="warn" stroke="#f59e0b" fill="#f59e0b" fillOpacity={0.15} stackId="1" />
              <Area type="monotone" dataKey="error" stroke="#ef4444" fill="#ef4444" fillOpacity={0.15} stackId="1" />
              <Area type="monotone" dataKey="debug" stroke="#71717a" fill="#71717a" fillOpacity={0.1} stackId="1" />
            </AreaChart>
          </ResponsiveContainer>
        </Card>

        {/* Top-N charts row */}
        <div className="grid grid-cols-2 gap-4">
          <Card>
            <CardHeader>
              <CardTitle>Top Sources</CardTitle>
            </CardHeader>
            <ResponsiveContainer width="100%" height={240}>
              <BarChart data={topSources?.data || []} layout="vertical">
                <XAxis type="number" tick={{ fill: "#71717a", fontSize: 11 }} axisLine={false} tickLine={false} />
                <YAxis
                  type="category" dataKey="key"
                  tick={{ fill: "#a1a1aa", fontSize: 11 }}
                  axisLine={false} tickLine={false}
                  width={120}
                />
                <Tooltip
                  contentStyle={{ backgroundColor: "#18181b", border: "1px solid #27272a", borderRadius: "6px", fontSize: "12px" }}
                />
                <Bar dataKey="count" radius={[0, 4, 4, 0]}>
                  {(topSources?.data || []).map((_, i) => (
                    <Cell key={i} fill={BAR_COLORS[i % BAR_COLORS.length]} fillOpacity={0.8} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Top Event Types</CardTitle>
            </CardHeader>
            <ResponsiveContainer width="100%" height={240}>
              <BarChart data={topTypes?.data || []} layout="vertical">
                <XAxis type="number" tick={{ fill: "#71717a", fontSize: 11 }} axisLine={false} tickLine={false} />
                <YAxis
                  type="category" dataKey="key"
                  tick={{ fill: "#a1a1aa", fontSize: 11 }}
                  axisLine={false} tickLine={false}
                  width={120}
                />
                <Tooltip
                  contentStyle={{ backgroundColor: "#18181b", border: "1px solid #27272a", borderRadius: "6px", fontSize: "12px" }}
                />
                <Bar dataKey="count" radius={[0, 4, 4, 0]}>
                  {(topTypes?.data || []).map((_, i) => (
                    <Cell key={i} fill={BAR_COLORS[(i + 3) % BAR_COLORS.length]} fillOpacity={0.8} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </Card>
        </div>

        {/* Severity breakdown table */}
        {summary?.by_severity && summary.by_severity.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>Breakdown by Severity</CardTitle>
            </CardHeader>
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-zinc-800">
                  <th className="text-left py-2 px-3 text-xs text-zinc-500 font-medium">Severity</th>
                  <th className="text-right py-2 px-3 text-xs text-zinc-500 font-medium">Count</th>
                  <th className="text-right py-2 px-3 text-xs text-zinc-500 font-medium">Percentage</th>
                </tr>
              </thead>
              <tbody>
                {summary.by_severity.map((s) => (
                  <tr key={s.severity} className="border-b border-zinc-900">
                    <td className="py-2 px-3 text-zinc-300 capitalize">{s.severity}</td>
                    <td className="py-2 px-3 text-zinc-400 text-right font-mono">{s.count.toLocaleString()}</td>
                    <td className="py-2 px-3 text-zinc-500 text-right font-mono">
                      {summary.total_events > 0
                        ? ((s.count / summary.total_events) * 100).toFixed(1)
                        : "0"}%
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        )}
      </div>
    </AuthGuard>
  );
}
