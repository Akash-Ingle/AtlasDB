"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { AuthGuard } from "@/components/layout/auth-guard";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

function useApiKeys() {
  const token = typeof window !== "undefined" ? localStorage.getItem("atlas_token") : null;
  return useQuery({
    queryKey: ["api-keys"],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/v1/auth/api-keys`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) throw new Error("Failed to fetch API keys");
      return res.json();
    },
  });
}

export default function SettingsPage() {
  return (
    <AuthGuard>
      <div className="p-6 space-y-6 max-w-3xl">
        <div>
          <h1 className="text-xl font-semibold text-white">Settings</h1>
          <p className="text-sm text-zinc-500 mt-1">
            Manage API keys, preferences, and system configuration
          </p>
        </div>

        <APIKeysSection />
        <DLQSection />
        <AboutSection />
      </div>
    </AuthGuard>
  );
}

function APIKeysSection() {
  const queryClient = useQueryClient();
  const { data } = useApiKeys();
  const [name, setName] = useState("");
  const [showCreate, setShowCreate] = useState(false);

  const token = typeof window !== "undefined" ? localStorage.getItem("atlas_token") : null;

  const createMutation = useMutation({
    mutationFn: async (keyName: string) => {
      const res = await fetch(`${API_BASE}/api/v1/auth/api-keys`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ name: keyName, scopes: ["events:write", "events:read", "search:read"] }),
      });
      if (!res.ok) throw new Error("Failed to create API key");
      return res.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
      setName("");
      setShowCreate(false);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      await fetch(`${API_BASE}/api/v1/auth/api-keys/${id}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      });
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["api-keys"] }),
  });

  const keys = data?.keys || [];

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>API Keys</CardTitle>
          <button
            onClick={() => setShowCreate(!showCreate)}
            className="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-500 transition-colors"
          >
            {showCreate ? "Cancel" : "Create Key"}
          </button>
        </div>
      </CardHeader>

      {showCreate && (
        <div className="px-4 pb-4 flex gap-2">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Key name (e.g. CI Pipeline)"
            className="flex-1 px-3 py-2 bg-zinc-900 border border-zinc-800 rounded text-sm text-zinc-300 focus:outline-none focus:border-zinc-600"
          />
          <button
            onClick={() => name && createMutation.mutate(name)}
            disabled={!name || createMutation.isPending}
            className="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-500 disabled:opacity-50"
          >
            Create
          </button>
        </div>
      )}

      {keys.length === 0 ? (
        <div className="px-4 pb-4 text-sm text-zinc-500">
          No API keys configured.
        </div>
      ) : (
        <div className="px-4 pb-4 space-y-2">
          {keys.map((key: { key_id: string; name: string; key_prefix: string; scopes: string[]; created_at: string }) => (
            <div
              key={key.key_id}
              className="flex items-center justify-between bg-zinc-900 rounded px-3 py-2"
            >
              <div>
                <span className="text-sm text-zinc-300">{key.name}</span>
                <span className="text-xs text-zinc-600 ml-2 font-mono">{key.key_prefix}...</span>
                <div className="flex gap-1 mt-1">
                  {key.scopes?.map((s: string) => (
                    <Badge key={s} className="bg-zinc-800 text-zinc-500 border-zinc-700 text-[10px]">
                      {s}
                    </Badge>
                  ))}
                </div>
              </div>
              <button
                onClick={() => { if (confirm("Delete this key?")) deleteMutation.mutate(key.key_id); }}
                className="text-xs text-red-400 hover:text-red-300"
              >
                Delete
              </button>
            </div>
          ))}
        </div>
      )}
    </Card>
  );
}

function DLQSection() {
  const token = typeof window !== "undefined" ? localStorage.getItem("atlas_token") : null;
  const { data } = useQuery({
    queryKey: ["dlq-stats"],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/v1/admin/dlq/stats`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) return { streams: [] };
      return res.json();
    },
    refetchInterval: 10000,
  });

  const streams = data?.streams || [];
  const totalMessages = streams.reduce((sum: number, s: { length: number }) => sum + s.length, 0);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Dead Letter Queue</CardTitle>
      </CardHeader>
      <div className="px-4 pb-4">
        {totalMessages === 0 ? (
          <p className="text-sm text-zinc-500">No messages in DLQ. All events processed successfully.</p>
        ) : (
          <div className="space-y-2">
            {streams.map((s: { stream: string; length: number }) => (
              <div key={s.stream} className="flex items-center justify-between bg-zinc-900 rounded px-3 py-2">
                <span className="text-sm text-zinc-300 font-mono">{s.stream}</span>
                <Badge className="bg-red-900/50 text-red-400 border-red-800">
                  {s.length} messages
                </Badge>
              </div>
            ))}
          </div>
        )}
      </div>
    </Card>
  );
}

function AboutSection() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>About AtlasDB</CardTitle>
      </CardHeader>
      <div className="px-4 pb-4 space-y-2 text-sm text-zinc-400">
        <p>
          <span className="text-zinc-300 font-medium">AtlasDB</span> is a production-grade
          distributed event streaming, search, and AI analytics platform.
        </p>
        <div className="grid grid-cols-2 gap-2 mt-3 text-xs">
          <div className="bg-zinc-900 rounded px-3 py-2">
            <span className="text-zinc-500">Backend</span>
            <p className="text-zinc-300">Go 1.23</p>
          </div>
          <div className="bg-zinc-900 rounded px-3 py-2">
            <span className="text-zinc-500">Frontend</span>
            <p className="text-zinc-300">Next.js + TypeScript</p>
          </div>
          <div className="bg-zinc-900 rounded px-3 py-2">
            <span className="text-zinc-500">Database</span>
            <p className="text-zinc-300">PostgreSQL 17 + pgvector</p>
          </div>
          <div className="bg-zinc-900 rounded px-3 py-2">
            <span className="text-zinc-500">Queue</span>
            <p className="text-zinc-300">Redis Streams</p>
          </div>
        </div>
      </div>
    </Card>
  );
}
