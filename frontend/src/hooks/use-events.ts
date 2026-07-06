"use client";

import { useQuery } from "@tanstack/react-query";
import { getEvents, type Event, type PaginatedResponse } from "@/lib/api";

export function useEvents(params?: {
  source?: string;
  severity?: string;
  limit?: number;
  refreshInterval?: number;
}) {
  return useQuery<PaginatedResponse<Event>>({
    queryKey: ["events", params?.source, params?.severity, params?.limit],
    queryFn: () =>
      getEvents({
        source: params?.source,
        severity: params?.severity,
        limit: params?.limit || 50,
      }),
    refetchInterval: params?.refreshInterval ?? 5000,
  });
}
