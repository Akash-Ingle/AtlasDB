"use client";

import { useState } from "react";
import { AuthGuard } from "@/components/layout/auth-guard";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { EventDetail } from "@/components/events/event-detail";
import { severityBg, formatTimestamp } from "@/lib/utils";
import { useSearch } from "@/hooks/use-search";
import type { Event } from "@/lib/api";
import { Search as SearchIcon } from "lucide-react";

export default function SearchPage() {
  const [query, setQuery] = useState("");
  const [submittedQuery, setSubmittedQuery] = useState("");
  const [sourceFilter, setSourceFilter] = useState("");
  const [severityFilter, setSeverityFilter] = useState("");
  const [selected, setSelected] = useState<Event | null>(null);

  const { data, isLoading, error } = useSearch(submittedQuery, {
    source: sourceFilter,
    severity: severityFilter,
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmittedQuery(query);
  }

  const results = data?.data || [];

  return (
    <AuthGuard>
      <div className="flex h-screen">
        <div className="flex-1 p-6 overflow-auto">
          <h1 className="text-xl font-semibold text-white mb-1">Search</h1>
          <p className="text-sm text-zinc-500 mb-6">
            Search events with keywords, filters, and field syntax
          </p>

          {/* Search bar */}
          <form onSubmit={handleSubmit} className="mb-6">
            <div className="relative">
              <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder='Search events... (e.g. "connection timeout" source:payment-service severity:error)'
                className="w-full pl-10 pr-4 py-3 rounded-lg bg-zinc-900 border border-zinc-800 text-white text-sm placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </form>

          {/* Filters */}
          <div className="flex gap-3 mb-6">
            <div>
              <label className="block text-xs text-zinc-500 mb-1">Source</label>
              <input
                type="text"
                value={sourceFilter}
                onChange={(e) => setSourceFilter(e.target.value)}
                placeholder="Any"
                className="px-3 py-1.5 rounded bg-zinc-900 border border-zinc-800 text-white text-xs w-40 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-xs text-zinc-500 mb-1">
                Severity
              </label>
              <select
                value={severityFilter}
                onChange={(e) => setSeverityFilter(e.target.value)}
                className="px-3 py-1.5 rounded bg-zinc-900 border border-zinc-800 text-white text-xs w-28 focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="">Any</option>
                <option value="debug">debug</option>
                <option value="info">info</option>
                <option value="warn">warn</option>
                <option value="error">error</option>
                <option value="fatal">fatal</option>
              </select>
            </div>
          </div>

          {/* Syntax help */}
          {!submittedQuery && (
            <Card className="mb-6">
              <CardHeader>
                <CardTitle>Search Syntax</CardTitle>
              </CardHeader>
              <div className="grid grid-cols-2 gap-3 text-xs">
                <SyntaxExample
                  syntax="connection timeout"
                  description="Keyword search"
                />
                <SyntaxExample
                  syntax="source:payment-service"
                  description="Filter by source"
                />
                <SyntaxExample
                  syntax="severity:error"
                  description="Filter by severity"
                />
                <SyntaxExample
                  syntax="error OR timeout"
                  description="Boolean operators"
                />
                <SyntaxExample
                  syntax='"connection refused"'
                  description="Exact phrase"
                />
                <SyntaxExample
                  syntax="source:auth-service login"
                  description="Combined filter + keyword"
                />
              </div>
            </Card>
          )}

          {/* Results */}
          {submittedQuery && (
            <div>
              <div className="flex items-center justify-between mb-3">
                <p className="text-xs text-zinc-500">
                  {isLoading
                    ? "Searching..."
                    : `${results.length} result${results.length !== 1 ? "s" : ""}`}
                </p>
              </div>

              {error && (
                <div className="text-sm text-red-400 bg-red-400/10 border border-red-400/20 rounded-md px-3 py-2 mb-4">
                  Search failed. Please try again.
                </div>
              )}

              <Card>
                {results.length === 0 && !isLoading ? (
                  <div className="text-center py-12 text-zinc-600 text-sm">
                    No results found for &ldquo;{submittedQuery}&rdquo;
                  </div>
                ) : (
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-zinc-800">
                        <th className="text-left py-2 px-3 text-xs text-zinc-500 font-medium">
                          Rank
                        </th>
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
                      {results.map((result) => (
                        <tr
                          key={result.event.event_id}
                          onClick={() => setSelected(result.event)}
                          className="border-b border-zinc-900 hover:bg-zinc-900/50 cursor-pointer transition-colors"
                        >
                          <td className="py-2 px-3 text-zinc-600 font-mono text-xs">
                            {result.rank.toFixed(3)}
                          </td>
                          <td className="py-2 px-3 text-zinc-400 font-mono text-xs whitespace-nowrap">
                            {formatTimestamp(result.event.timestamp)}
                          </td>
                          <td className="py-2 px-3">
                            <Badge
                              className={severityBg[result.event.severity]}
                            >
                              {result.event.severity}
                            </Badge>
                          </td>
                          <td className="py-2 px-3 text-zinc-300 font-mono text-xs">
                            {result.event.source}
                          </td>
                          <td className="py-2 px-3 text-zinc-400 text-xs">
                            {result.event.event_type}
                          </td>
                          <td className="py-2 px-3 text-zinc-500 text-xs max-w-xs truncate">
                            {JSON.stringify(result.event.data).slice(0, 60)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </Card>
            </div>
          )}
        </div>

        {/* Detail panel */}
        {selected && (
          <EventDetail event={selected} onClose={() => setSelected(null)} />
        )}
      </div>
    </AuthGuard>
  );
}

function SyntaxExample({
  syntax,
  description,
}: {
  syntax: string;
  description: string;
}) {
  return (
    <div className="flex items-center gap-2">
      <code className="px-2 py-1 rounded bg-zinc-900 text-blue-400 font-mono text-xs">
        {syntax}
      </code>
      <span className="text-zinc-500">{description}</span>
    </div>
  );
}
