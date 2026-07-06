"use client";

import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { AuthGuard } from "@/components/layout/auth-guard";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { Stat } from "@/components/ui/stat";
import { EventTable } from "@/components/events/event-table";
import { EventVolumeChart } from "@/components/charts/event-volume-chart";
import { getEvents, getHealth } from "@/lib/api";

export default function OverviewPage() {
  const { data: eventsData } = useQuery({
    queryKey: ["overview-events"],
    queryFn: () => getEvents({ limit: 100 }),
    refetchInterval: 5000,
  });

  const { data: health } = useQuery({
    queryKey: ["health"],
    queryFn: getHealth,
    refetchInterval: 10000,
  });

  const events = eventsData?.data || [];

  const stats = useMemo(() => {
    const total = events.length;
    const errors = events.filter(
      (e) => e.severity === "error" || e.severity === "fatal"
    ).length;
    const sources = new Set(events.map((e) => e.source)).size;
    const errorRate = total > 0 ? ((errors / total) * 100).toFixed(1) : "0";
    return { total, errors, sources, errorRate };
  }, [events]);

  const healthStatus =
    health?.status === "ready" ? "Healthy" : health?.status || "Unknown";

  return (
    <AuthGuard>
      <div className="p-6 space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-xl font-semibold text-white">Overview</h1>
          <p className="text-sm text-zinc-500 mt-1">
            System health and event activity
          </p>
        </div>

        {/* Stats Row */}
        <div className="grid grid-cols-4 gap-4">
          <Stat label="Total Events" value={stats.total} />
          <Stat
            label="Error Rate"
            value={`${stats.errorRate}%`}
            change={stats.errors > 0 ? `${stats.errors} errors` : "No errors"}
            changeType={stats.errors > 0 ? "negative" : "positive"}
          />
          <Stat label="Active Sources" value={stats.sources} />
          <Stat
            label="System Health"
            value={healthStatus}
            changeType={healthStatus === "Healthy" ? "positive" : "negative"}
          />
        </div>

        {/* Event Volume Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Event Volume (last hour)</CardTitle>
          </CardHeader>
          <EventVolumeChart events={events} />
        </Card>

        {/* Recent Events */}
        <Card>
          <CardHeader>
            <CardTitle>Recent Events</CardTitle>
          </CardHeader>
          <EventTable events={events.slice(0, 20)} />
        </Card>
      </div>
    </AuthGuard>
  );
}
