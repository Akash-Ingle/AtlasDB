"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { AuthGuard } from "@/components/layout/auth-guard";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { EventDetail } from "@/components/events/event-detail";
import { severityBg, formatTimestamp } from "@/lib/utils";
import { createEventStream, type Event } from "@/lib/api";
import { useEvents } from "@/hooks/use-events";
import { Pause, Play, Radio } from "lucide-react";

export default function EventsPage() {
  const [liveEvents, setLiveEvents] = useState<Event[]>([]);
  const [paused, setPaused] = useState(false);
  const [selected, setSelected] = useState<Event | null>(null);
  const [severityFilter, setSeverityFilter] = useState<string>("");
  const wsRef = useRef<WebSocket | null>(null);

  // Historical events as fallback
  const { data: historicalData } = useEvents({ limit: 100 });

  // WebSocket for live events
  useEffect(() => {
    const ws = createEventStream((event) => {
      if (!paused) {
        setLiveEvents((prev) => [event, ...prev].slice(0, 500));
      }
    });
    wsRef.current = ws;

    return () => {
      ws.close();
    };
  }, [paused]);

  // Merge live + historical, deduplicate
  const allEvents = (() => {
    const map = new Map<string, Event>();
    for (const e of liveEvents) map.set(e.event_id, e);
    for (const e of historicalData?.data || []) {
      if (!map.has(e.event_id)) map.set(e.event_id, e);
    }
    let events = Array.from(map.values()).sort(
      (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
    );
    if (severityFilter) {
      events = events.filter((e) => e.severity === severityFilter);
    }
    return events;
  })();

  const togglePause = useCallback(() => setPaused((p) => !p), []);

  const severities = ["", "debug", "info", "warn", "error", "fatal"];

  return (
    <AuthGuard>
      <div className="flex h-screen">
        <div className="flex-1 flex flex-col overflow-hidden">
          <div className="p-6 pb-0">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <h1 className="text-xl font-semibold text-white">
                  Live Events
                </h1>
                <div className="flex items-center gap-1.5">
                  <Radio
                    className={`h-3 w-3 ${paused ? "text-zinc-600" : "text-emerald-400 animate-pulse"}`}
                  />
                  <span className="text-xs text-zinc-500">
                    {paused ? "Paused" : "Live"}
                  </span>
                </div>
              </div>

              <div className="flex items-center gap-3">
                {/* Severity filter */}
                <div className="flex gap-1">
                  {severities.map((s) => (
                    <button
                      key={s || "all"}
                      onClick={() => setSeverityFilter(s)}
                      className={`px-2 py-1 text-xs rounded transition-colors ${
                        severityFilter === s
                          ? "bg-zinc-700 text-white"
                          : "text-zinc-500 hover:text-zinc-300"
                      }`}
                    >
                      {s || "All"}
                    </button>
                  ))}
                </div>

                {/* Pause/Resume */}
                <button
                  onClick={togglePause}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-zinc-800 text-zinc-300 text-xs hover:bg-zinc-700 transition-colors"
                >
                  {paused ? (
                    <Play className="h-3 w-3" />
                  ) : (
                    <Pause className="h-3 w-3" />
                  )}
                  {paused ? "Resume" : "Pause"}
                </button>
              </div>
            </div>
          </div>

          {/* Event stream */}
          <Card className="mx-6 mb-6 flex-1 overflow-hidden flex flex-col">
            <div className="flex-1 overflow-y-auto">
              {allEvents.length === 0 ? (
                <div className="flex items-center justify-center h-full text-zinc-600 text-sm">
                  Waiting for events...
                </div>
              ) : (
                <table className="w-full text-sm">
                  <tbody>
                    {allEvents.map((event) => (
                      <tr
                        key={event.event_id}
                        onClick={() => setSelected(event)}
                        className="border-b border-zinc-900 hover:bg-zinc-900/50 cursor-pointer transition-colors"
                      >
                        <td className="py-1.5 px-3 text-zinc-500 font-mono text-xs w-20">
                          {formatTimestamp(event.timestamp)}
                        </td>
                        <td className="py-1.5 px-2 w-16">
                          <Badge className={severityBg[event.severity]}>
                            {event.severity}
                          </Badge>
                        </td>
                        <td className="py-1.5 px-2 text-zinc-400 font-mono text-xs w-40">
                          {event.source}
                        </td>
                        <td className="py-1.5 px-2 text-zinc-500 text-xs truncate">
                          {event.event_type}
                          {event.data &&
                            typeof event.data === "object" &&
                            "path" in event.data &&
                            ` ${event.data.method || ""} ${event.data.path}`}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </Card>
        </div>

        {/* Detail panel */}
        {selected && (
          <EventDetail
            event={selected}
            onClose={() => setSelected(null)}
          />
        )}
      </div>
    </AuthGuard>
  );
}
