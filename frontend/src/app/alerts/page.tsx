"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { AuthGuard } from "@/components/layout/auth-guard";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  getAlertRules,
  getAlertHistory,
  createAlertRule,
  deleteAlertRule,
  updateAlertRule,
  type AlertRule,
  type AlertCondition,
} from "@/lib/api";

type Tab = "rules" | "history";

const METRIC_OPTIONS = ["error_rate", "event_count", "error_count"];
const OPERATOR_OPTIONS = [">", ">=", "<", "<=", "=="];
const WINDOW_OPTIONS = ["1m", "5m", "15m", "1h"];
const SEVERITY_OPTIONS = ["info", "warning", "error", "critical"];

export default function AlertsPage() {
  const [tab, setTab] = useState<Tab>("rules");
  const [showCreate, setShowCreate] = useState(false);

  return (
    <AuthGuard>
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold text-white">Alerts</h1>
            <p className="text-sm text-zinc-500 mt-1">
              Manage alert rules and view firing history
            </p>
          </div>
          <button
            onClick={() => setShowCreate(!showCreate)}
            className="px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-500 transition-colors"
          >
            {showCreate ? "Cancel" : "Create Rule"}
          </button>
        </div>

        {showCreate && <CreateRuleForm onDone={() => setShowCreate(false)} />}

        {/* Tabs */}
        <div className="flex gap-1 bg-zinc-900 rounded-md p-1 w-fit">
          <button
            onClick={() => setTab("rules")}
            className={`px-4 py-1.5 text-sm rounded transition-colors ${
              tab === "rules"
                ? "bg-zinc-700 text-white"
                : "text-zinc-500 hover:text-zinc-300"
            }`}
          >
            Rules
          </button>
          <button
            onClick={() => setTab("history")}
            className={`px-4 py-1.5 text-sm rounded transition-colors ${
              tab === "history"
                ? "bg-zinc-700 text-white"
                : "text-zinc-500 hover:text-zinc-300"
            }`}
          >
            History
          </button>
        </div>

        {tab === "rules" && <RulesTab />}
        {tab === "history" && <HistoryTab />}
      </div>
    </AuthGuard>
  );
}

function RulesTab() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ["alert-rules"],
    queryFn: getAlertRules,
    refetchInterval: 15000,
  });

  const deleteMutation = useMutation({
    mutationFn: deleteAlertRule,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["alert-rules"] }),
  });

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      updateAlertRule(id, { enabled } as Partial<AlertRule>),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["alert-rules"] }),
  });

  const rules = data?.rules || [];

  if (isLoading) {
    return <div className="text-zinc-500 text-sm py-8 text-center">Loading rules...</div>;
  }

  if (rules.length === 0) {
    return (
      <Card>
        <div className="py-12 text-center text-zinc-500 text-sm">
          No alert rules configured. Create one to get started.
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-3">
      {rules.map((rule) => (
        <Card key={rule.rule_id}>
          <div className="p-4 flex items-start justify-between">
            <div className="space-y-1 flex-1">
              <div className="flex items-center gap-2">
                <h3 className="text-sm font-medium text-white">{rule.name}</h3>
                <Badge
                  variant={rule.enabled ? "default" : "outline"}
                  className={rule.enabled ? "bg-emerald-900/50 text-emerald-400 border-emerald-800" : ""}
                >
                  {rule.enabled ? "Active" : "Disabled"}
                </Badge>
                <Badge className={severityColor(rule.severity)}>{rule.severity}</Badge>
              </div>
              {rule.description && (
                <p className="text-xs text-zinc-500">{rule.description}</p>
              )}
              <p className="text-xs text-zinc-400 font-mono">
                {rule.condition.metric} {rule.condition.operator} {rule.condition.threshold}
                {" "}(window: {rule.condition.window})
                {rule.condition.source ? ` [source: ${rule.condition.source}]` : ""}
              </p>
              <p className="text-xs text-zinc-600">
                Cooldown: {rule.cooldown_seconds}s
              </p>
            </div>
            <div className="flex gap-2 ml-4">
              <button
                onClick={() =>
                  toggleMutation.mutate({
                    id: rule.rule_id,
                    enabled: !rule.enabled,
                  })
                }
                className="px-3 py-1.5 text-xs rounded bg-zinc-800 text-zinc-400 hover:text-white transition-colors"
              >
                {rule.enabled ? "Disable" : "Enable"}
              </button>
              <button
                onClick={() => {
                  if (confirm("Delete this rule?")) {
                    deleteMutation.mutate(rule.rule_id);
                  }
                }}
                className="px-3 py-1.5 text-xs rounded bg-zinc-800 text-red-400 hover:text-red-300 transition-colors"
              >
                Delete
              </button>
            </div>
          </div>
        </Card>
      ))}
    </div>
  );
}

function HistoryTab() {
  const { data, isLoading } = useQuery({
    queryKey: ["alert-history"],
    queryFn: () => getAlertHistory(50),
    refetchInterval: 10000,
  });

  const events = data?.alerts || [];

  if (isLoading) {
    return <div className="text-zinc-500 text-sm py-8 text-center">Loading history...</div>;
  }

  if (events.length === 0) {
    return (
      <Card>
        <div className="py-12 text-center text-zinc-500 text-sm">
          No alerts have fired yet.
        </div>
      </Card>
    );
  }

  return (
    <Card>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-zinc-800">
            <th className="text-left py-2 px-3 text-xs text-zinc-500 font-medium">Rule</th>
            <th className="text-left py-2 px-3 text-xs text-zinc-500 font-medium">Status</th>
            <th className="text-right py-2 px-3 text-xs text-zinc-500 font-medium">Value</th>
            <th className="text-left py-2 px-3 text-xs text-zinc-500 font-medium">Fired At</th>
            <th className="text-center py-2 px-3 text-xs text-zinc-500 font-medium">Notified</th>
          </tr>
        </thead>
        <tbody>
          {events.map((e) => (
            <tr key={e.alert_event_id} className="border-b border-zinc-900 hover:bg-zinc-900/50">
              <td className="py-2 px-3 text-zinc-300">{e.rule_name || e.rule_id.slice(0, 8)}</td>
              <td className="py-2 px-3">
                <Badge
                  className={
                    e.status === "firing"
                      ? "bg-red-900/50 text-red-400 border-red-800"
                      : "bg-emerald-900/50 text-emerald-400 border-emerald-800"
                  }
                >
                  {e.status}
                </Badge>
              </td>
              <td className="py-2 px-3 text-zinc-400 text-right font-mono">{e.value.toFixed(4)}</td>
              <td className="py-2 px-3 text-zinc-500 text-xs">
                {new Date(e.fired_at).toLocaleString()}
              </td>
              <td className="py-2 px-3 text-center">
                {e.notified ? (
                  <span className="text-emerald-500 text-xs">Yes</span>
                ) : (
                  <span className="text-zinc-600 text-xs">No</span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </Card>
  );
}

function CreateRuleForm({ onDone }: { onDone: () => void }) {
  const queryClient = useQueryClient();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [metric, setMetric] = useState("error_rate");
  const [operator, setOperator] = useState(">");
  const [threshold, setThreshold] = useState("0.05");
  const [window, setWindow] = useState("5m");
  const [severity, setSeverity] = useState("warning");
  const [source, setSource] = useState("");
  const [cooldown, setCooldown] = useState("300");
  const [webhookURL, setWebhookURL] = useState("");

  const mutation = useMutation({
    mutationFn: createAlertRule,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["alert-rules"] });
      onDone();
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const condition: AlertCondition = {
      metric,
      operator,
      threshold: parseFloat(threshold),
      window,
    };
    if (source) condition.source = source;

    const channels = webhookURL
      ? [{ type: "webhook", url: webhookURL }]
      : [];

    mutation.mutate({
      name,
      description,
      condition,
      severity,
      channels,
      cooldown_seconds: parseInt(cooldown) || 300,
    });
  };

  const inputClass =
    "w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-md text-sm text-zinc-300 focus:outline-none focus:border-zinc-600";
  const selectClass =
    "w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-md text-sm text-zinc-300 focus:outline-none focus:border-zinc-600";

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create Alert Rule</CardTitle>
      </CardHeader>
      <form onSubmit={handleSubmit} className="p-4 pt-0 space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs text-zinc-500 mb-1">Name</label>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              className={inputClass}
              placeholder="High Error Rate"
              required
            />
          </div>
          <div>
            <label className="block text-xs text-zinc-500 mb-1">Severity</label>
            <select value={severity} onChange={(e) => setSeverity(e.target.value)} className={selectClass}>
              {SEVERITY_OPTIONS.map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </select>
          </div>
        </div>

        <div>
          <label className="block text-xs text-zinc-500 mb-1">Description (optional)</label>
          <input
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className={inputClass}
            placeholder="Alert when error rate exceeds threshold"
          />
        </div>

        <div className="grid grid-cols-4 gap-4">
          <div>
            <label className="block text-xs text-zinc-500 mb-1">Metric</label>
            <select value={metric} onChange={(e) => setMetric(e.target.value)} className={selectClass}>
              {METRIC_OPTIONS.map((m) => (
                <option key={m} value={m}>{m}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs text-zinc-500 mb-1">Operator</label>
            <select value={operator} onChange={(e) => setOperator(e.target.value)} className={selectClass}>
              {OPERATOR_OPTIONS.map((o) => (
                <option key={o} value={o}>{o}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs text-zinc-500 mb-1">Threshold</label>
            <input
              value={threshold}
              onChange={(e) => setThreshold(e.target.value)}
              className={inputClass}
              type="number"
              step="any"
              required
            />
          </div>
          <div>
            <label className="block text-xs text-zinc-500 mb-1">Window</label>
            <select value={window} onChange={(e) => setWindow(e.target.value)} className={selectClass}>
              {WINDOW_OPTIONS.map((w) => (
                <option key={w} value={w}>{w}</option>
              ))}
            </select>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs text-zinc-500 mb-1">Source filter (optional)</label>
            <input
              value={source}
              onChange={(e) => setSource(e.target.value)}
              className={inputClass}
              placeholder="payment-service"
            />
          </div>
          <div>
            <label className="block text-xs text-zinc-500 mb-1">Cooldown (seconds)</label>
            <input
              value={cooldown}
              onChange={(e) => setCooldown(e.target.value)}
              className={inputClass}
              type="number"
            />
          </div>
        </div>

        <div>
          <label className="block text-xs text-zinc-500 mb-1">Webhook URL (optional)</label>
          <input
            value={webhookURL}
            onChange={(e) => setWebhookURL(e.target.value)}
            className={inputClass}
            placeholder="https://hooks.slack.com/services/..."
          />
        </div>

        <div className="flex gap-3 pt-2">
          <button
            type="submit"
            disabled={mutation.isPending}
            className="px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-500 disabled:opacity-50 transition-colors"
          >
            {mutation.isPending ? "Creating..." : "Create Rule"}
          </button>
          <button
            type="button"
            onClick={onDone}
            className="px-4 py-2 bg-zinc-800 text-zinc-400 text-sm rounded-md hover:text-white transition-colors"
          >
            Cancel
          </button>
        </div>

        {mutation.isError && (
          <p className="text-xs text-red-400">Failed to create rule. Check your inputs.</p>
        )}
      </form>
    </Card>
  );
}

function severityColor(severity: string): string {
  switch (severity) {
    case "critical":
      return "bg-red-900/50 text-red-400 border-red-800";
    case "error":
      return "bg-orange-900/50 text-orange-400 border-orange-800";
    case "warning":
      return "bg-yellow-900/50 text-yellow-400 border-yellow-800";
    default:
      return "bg-zinc-800 text-zinc-400 border-zinc-700";
  }
}
