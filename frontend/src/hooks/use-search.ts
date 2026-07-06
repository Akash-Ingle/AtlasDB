"use client";

import { useQuery } from "@tanstack/react-query";
import {
  searchEvents,
  type SearchResult,
  type PaginatedResponse,
} from "@/lib/api";

export function useSearch(query: string, params?: { source?: string; severity?: string }) {
  return useQuery<PaginatedResponse<SearchResult>>({
    queryKey: ["search", query, params?.source, params?.severity],
    queryFn: () =>
      searchEvents({
        q: query,
        source: params?.source,
        severity: params?.severity,
      }),
    enabled: query.length > 0,
  });
}
