"use client";

import { useQuery } from "@tanstack/react-query";
import { AuthGuard } from "@/components/layout/auth-guard";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { getHealth } from "@/lib/api";

const EXTERNAL_LINKS = [
  { name: "Grafana", url: "http://localhost:3001", desc: "Dashboards & metrics" },
  { name: "Prometheus", url: "http://localhost:9090", desc: "Metric queries" },
  { name: "Jaeger", url: "http://localhost:16686", desc: "Distributed traces" },
];

export default function SystemPage() {
  const { data: health, isLoading, error } = useQuery({
    queryKey: ["system-health"],
    queryFn: getHealth,
    refetchInterval: 5000,
    retry: 2,
  });

  const isHealthy = health?.status === "ok";
  const checks = health?.checks || {};

  return (
    <AuthGuard>
      <div className="p-6 space-y-6">
        <div>
          <h1 className="text-xl font-semibold text-white">System Health</h1>
          <p className="text-sm text-zinc-500 mt-1">
            Service status, infrastructure, and observability links
          </p>
        </div>

        {/* Overall status */}
        <Card>
          <div className="p-6 flex items-center justify-between">
            <div>
              <h2 className="text-sm text-zinc-400 font-medium">Overall Status</h2>
              <p className="text-2xl font-bold text-white mt-1">
                {isLoading ? "Checking..." : isHealthy ? "All Systems Operational" : "Degraded"}
              </p>
            </div>
            {!isLoading && (
              <div
                className={`w-4 h-4 rounded-full ${
                  isHealthy ? "bg-emerald-500 shadow-lg shadow-emerald-500/30" : "bg-red-500 shadow-lg shadow-red-500/30"
                }`}
              />
            )}
            {error && (
              <Badge className="bg-red-900/50 text-red-400 border-red-800">
                Unreachable
              </Badge>
            )}
          </div>
        </Card>

        {/* Service checks */}
        <div className="grid grid-cols-2 gap-4">
          {Object.entries(checks).map(([name, status]) => (
            <Card key={name}>
              <div className="p-4 flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-medium text-zinc-300 capitalize">
                    {name.replace(/_/g, " ")}
                  </h3>
                </div>
                <Badge
                  className={
                    status === "ok"
                      ? "bg-emerald-900/50 text-emerald-400 border-emerald-800"
                      : "bg-red-900/50 text-red-400 border-red-800"
                  }
                >
                  {status as string}
                </Badge>
              </div>
            </Card>
          ))}

          {Object.keys(checks).length === 0 && !isLoading && (
            <Card>
              <div className="p-4 text-sm text-zinc-500">
                No health checks available
              </div>
            </Card>
          )}
        </div>

        {/* Observability links */}
        <div>
          <h2 className="text-sm font-medium text-zinc-400 mb-3">
            Observability Stack
          </h2>
          <div className="grid grid-cols-3 gap-4">
            {EXTERNAL_LINKS.map((link) => (
              <a
                key={link.name}
                href={link.url}
                target="_blank"
                rel="noopener noreferrer"
                className="block"
              >
                <Card className="hover:border-zinc-600 transition-colors cursor-pointer">
                  <div className="p-4">
                    <h3 className="text-sm font-medium text-blue-400">
                      {link.name} &rarr;
                    </h3>
                    <p className="text-xs text-zinc-500 mt-1">{link.desc}</p>
                  </div>
                </Card>
              </a>
            ))}
          </div>
        </div>

        {/* Architecture info */}
        <Card>
          <CardHeader>
            <CardTitle>Service Architecture</CardTitle>
          </CardHeader>
          <div className="px-4 pb-4">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-zinc-800">
                  <th className="text-left py-2 text-xs text-zinc-500 font-medium">Service</th>
                  <th className="text-left py-2 text-xs text-zinc-500 font-medium">Role</th>
                  <th className="text-left py-2 text-xs text-zinc-500 font-medium">Port</th>
                </tr>
              </thead>
              <tbody className="text-zinc-400">
                <tr className="border-b border-zinc-900">
                  <td className="py-2 text-zinc-300">api-server</td>
                  <td className="py-2">HTTP API, event ingestion, WebSocket</td>
                  <td className="py-2 font-mono">8080</td>
                </tr>
                <tr className="border-b border-zinc-900">
                  <td className="py-2 text-zinc-300">processor</td>
                  <td className="py-2">Stream consumer, writes to DB + aggregations</td>
                  <td className="py-2 font-mono">-</td>
                </tr>
                <tr className="border-b border-zinc-900">
                  <td className="py-2 text-zinc-300">worker</td>
                  <td className="py-2">Background scheduler: alerts, rollups, partitions</td>
                  <td className="py-2 font-mono">-</td>
                </tr>
                <tr className="border-b border-zinc-900">
                  <td className="py-2 text-zinc-300">dashboard</td>
                  <td className="py-2">Next.js frontend</td>
                  <td className="py-2 font-mono">3000</td>
                </tr>
                <tr className="border-b border-zinc-900">
                  <td className="py-2 text-zinc-300">postgres</td>
                  <td className="py-2">Primary data store (pgvector)</td>
                  <td className="py-2 font-mono">5432</td>
                </tr>
                <tr className="border-b border-zinc-900">
                  <td className="py-2 text-zinc-300">redis</td>
                  <td className="py-2">Queue (Streams), cache, pub/sub</td>
                  <td className="py-2 font-mono">6379</td>
                </tr>
                <tr className="border-b border-zinc-900">
                  <td className="py-2 text-zinc-300">otel-collector</td>
                  <td className="py-2">Trace collection, export to Jaeger</td>
                  <td className="py-2 font-mono">4317</td>
                </tr>
                <tr className="border-b border-zinc-900">
                  <td className="py-2 text-zinc-300">jaeger</td>
                  <td className="py-2">Trace visualization</td>
                  <td className="py-2 font-mono">16686</td>
                </tr>
                <tr className="border-b border-zinc-900">
                  <td className="py-2 text-zinc-300">prometheus</td>
                  <td className="py-2">Metrics collection</td>
                  <td className="py-2 font-mono">9090</td>
                </tr>
                <tr>
                  <td className="py-2 text-zinc-300">grafana</td>
                  <td className="py-2">Metrics dashboards</td>
                  <td className="py-2 font-mono">3001</td>
                </tr>
              </tbody>
            </table>
          </div>
        </Card>
      </div>
    </AuthGuard>
  );
}
