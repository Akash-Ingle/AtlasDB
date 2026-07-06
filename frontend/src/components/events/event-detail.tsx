"use client";

import type { Event } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { severityBg, formatDate } from "@/lib/utils";
import { X } from "lucide-react";

interface Props {
  event: Event;
  onClose: () => void;
}

export function EventDetail({ event, onClose }: Props) {
  return (
    <div className="border-l border-zinc-800 w-96 bg-zinc-950 p-4 overflow-y-auto">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-white">Event Detail</h3>
        <button
          onClick={onClose}
          className="text-zinc-500 hover:text-white transition-colors"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      <div className="space-y-4">
        <Field label="Event ID" value={event.event_id} mono />
        <Field label="Timestamp" value={formatDate(event.timestamp)} />
        <div>
          <FieldLabel>Severity</FieldLabel>
          <Badge className={severityBg[event.severity]}>{event.severity}</Badge>
        </div>
        <Field label="Source" value={event.source} mono />
        <Field label="Type" value={event.event_type} />

        {event.tags.length > 0 && (
          <div>
            <FieldLabel>Tags</FieldLabel>
            <div className="flex flex-wrap gap-1">
              {event.tags.map((tag) => (
                <Badge
                  key={tag}
                  className="bg-zinc-800 text-zinc-400 border-zinc-700"
                >
                  {tag}
                </Badge>
              ))}
            </div>
          </div>
        )}

        <div>
          <FieldLabel>Data</FieldLabel>
          <pre className="mt-1 p-3 rounded bg-zinc-900 border border-zinc-800 text-xs text-zinc-300 overflow-x-auto">
            {JSON.stringify(event.data, null, 2)}
          </pre>
        </div>

        {event.metadata && Object.keys(event.metadata).length > 0 && (
          <div>
            <FieldLabel>Metadata</FieldLabel>
            <pre className="mt-1 p-3 rounded bg-zinc-900 border border-zinc-800 text-xs text-zinc-300 overflow-x-auto">
              {JSON.stringify(event.metadata, null, 2)}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}

function FieldLabel({ children }: { children: React.ReactNode }) {
  return (
    <p className="text-xs text-zinc-500 uppercase tracking-wider mb-1">
      {children}
    </p>
  );
}

function Field({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div>
      <FieldLabel>{label}</FieldLabel>
      <p className={`text-sm text-zinc-300 ${mono ? "font-mono" : ""}`}>
        {value}
      </p>
    </div>
  );
}
