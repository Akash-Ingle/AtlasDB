"use client";

import type { Event } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { severityBg, formatTimestamp } from "@/lib/utils";

interface Props {
  events: Event[];
  onSelect?: (event: Event) => void;
}

export function EventTable({ events, onSelect }: Props) {
  if (events.length === 0) {
    return (
      <div className="text-center py-12 text-zinc-600 text-sm">
        No events found
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-zinc-800">
            <th className="text-left py-2 px-3 text-xs text-zinc-500 font-medium">
              Time
            </th>
            <th className="text-left py-2 px-3 text-xs text-zinc-500 font-medium">
              Severity
            </th>
            <th className="text-left py-2 px-3 text-xs text-zinc-500 font-medium">
              Source
            </th>
            <th className="text-left py-2 px-3 text-xs text-zinc-500 font-medium">
              Type
            </th>
            <th className="text-left py-2 px-3 text-xs text-zinc-500 font-medium">
              Details
            </th>
          </tr>
        </thead>
        <tbody>
          {events.map((event) => (
            <tr
              key={event.event_id}
              onClick={() => onSelect?.(event)}
              className="border-b border-zinc-900 hover:bg-zinc-900/50 cursor-pointer transition-colors"
            >
              <td className="py-2 px-3 text-zinc-400 font-mono text-xs whitespace-nowrap">
                {formatTimestamp(event.timestamp)}
              </td>
              <td className="py-2 px-3">
                <Badge className={severityBg[event.severity]}>
                  {event.severity}
                </Badge>
              </td>
              <td className="py-2 px-3 text-zinc-300 font-mono text-xs">
                {event.source}
              </td>
              <td className="py-2 px-3 text-zinc-400 text-xs">
                {event.event_type}
              </td>
              <td className="py-2 px-3 text-zinc-500 text-xs max-w-xs truncate">
                {summarizeData(event.data)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function summarizeData(data: Record<string, unknown>): string {
  const parts: string[] = [];
  if (data.method) parts.push(String(data.method));
  if (data.path) parts.push(String(data.path));
  if (data.status) parts.push(`→ ${data.status}`);
  if (data.duration_ms) parts.push(`${data.duration_ms}ms`);
  if (data.error) parts.push(String(data.error));
  if (data.result) parts.push(String(data.result));
  if (parts.length === 0) {
    return JSON.stringify(data).slice(0, 80);
  }
  return parts.join(" ");
}
