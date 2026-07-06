"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { AuthGuard } from "@/components/layout/auth-guard";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface ChatMessage {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  confidence?: number;
  steps?: InvestigationStep[];
  query?: Record<string, unknown>;
  timestamp: Date;
}

interface InvestigationStep {
  step_number: number;
  type: string;
  content: string;
  tool_name?: string;
  tool_args?: string;
}

type Mode = "query" | "investigate";

const EXAMPLE_QUESTIONS = [
  "Show me errors from the last hour",
  "What's the error rate for payment-service?",
  "Why are there more errors than usual?",
  "Which service has the most events today?",
  "Find login failures from yesterday",
  "Are there any anomalies right now?",
];

export default function CopilotPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [mode, setMode] = useState<Mode>("query");
  const [isStreaming, setIsStreaming] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  const queryMutation = useMutation({
    mutationFn: async (question: string) => {
      const token = localStorage.getItem("atlas_token");
      const res = await fetch(`${API_BASE}/api/v1/ai/query`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ question }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error?.message || "AI query failed");
      }
      return res.json();
    },
    onSuccess: (data) => {
      const msg: ChatMessage = {
        id: crypto.randomUUID(),
        role: "assistant",
        content: data.answer,
        confidence: data.confidence,
        query: data.query,
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, msg]);
    },
    onError: (err: Error) => {
      setMessages((prev) => [
        ...prev,
        {
          id: crypto.randomUUID(),
          role: "assistant",
          content: `Error: ${err.message}. Make sure AI is enabled (AI_ENABLED=true) and an API key is configured.`,
          timestamp: new Date(),
        },
      ]);
    },
  });

  const investigate = useCallback(
    async (question: string) => {
      setIsStreaming(true);

      const assistantId = crypto.randomUUID();
      setMessages((prev) => [
        ...prev,
        {
          id: assistantId,
          role: "assistant",
          content: "Investigating...",
          steps: [],
          timestamp: new Date(),
        },
      ]);

      try {
        const token = localStorage.getItem("atlas_token");
        const res = await fetch(`${API_BASE}/api/v1/ai/investigate`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({ question, stream: true }),
        });

        if (!res.ok) {
          throw new Error("Investigation request failed");
        }

        const reader = res.body?.getReader();
        if (!reader) throw new Error("No response stream");

        const decoder = new TextDecoder();
        let buffer = "";
        const allSteps: InvestigationStep[] = [];

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n");
          buffer = lines.pop() || "";

          for (const line of lines) {
            if (line.startsWith("data: ")) {
              try {
                const data = JSON.parse(line.slice(6));
                if (data.type) {
                  allSteps.push(data as InvestigationStep);

                  const lastAnswer = allSteps.find(
                    (s) => s.type === "answer"
                  );
                  setMessages((prev) =>
                    prev.map((m) =>
                      m.id === assistantId
                        ? {
                            ...m,
                            content: lastAnswer?.content || "Investigating...",
                            steps: [...allSteps],
                          }
                        : m
                    )
                  );
                }
              } catch {
                // skip malformed SSE data
              }
            }
          }
        }

        // Final update
        const finalAnswer = allSteps.find((s) => s.type === "answer");
        setMessages((prev) =>
          prev.map((m) =>
            m.id === assistantId
              ? {
                  ...m,
                  content: finalAnswer?.content || "Investigation complete.",
                  steps: allSteps,
                }
              : m
          )
        );
      } catch (err) {
        setMessages((prev) =>
          prev.map((m) =>
            m.id === assistantId
              ? {
                  ...m,
                  content: `Investigation failed: ${err instanceof Error ? err.message : "Unknown error"}. Ensure AI_ENABLED=true.`,
                }
              : m
          )
        );
      } finally {
        setIsStreaming(false);
      }
    },
    []
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || queryMutation.isPending || isStreaming) return;

    const userMsg: ChatMessage = {
      id: crypto.randomUUID(),
      role: "user",
      content: input.trim(),
      timestamp: new Date(),
    };
    setMessages((prev) => [...prev, userMsg]);

    const question = input.trim();
    setInput("");

    if (mode === "investigate") {
      investigate(question);
    } else {
      queryMutation.mutate(question);
    }
  };

  const { data: anomalies } = useQuery({
    queryKey: ["anomalies"],
    queryFn: async () => {
      const token = localStorage.getItem("atlas_token");
      const res = await fetch(`${API_BASE}/api/v1/ai/anomalies`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) return null;
      return res.json();
    },
    refetchInterval: 60000,
    retry: false,
  });

  const isLoading = queryMutation.isPending || isStreaming;

  return (
    <AuthGuard>
      <div className="flex flex-col h-[calc(100vh-0px)]">
        {/* Header */}
        <div className="border-b border-zinc-800 px-6 py-4 flex items-center justify-between shrink-0">
          <div>
            <h1 className="text-xl font-semibold text-white">AI Copilot</h1>
            <p className="text-sm text-zinc-500 mt-0.5">
              Ask questions about your events, metrics, and alerts
            </p>
          </div>
          <div className="flex gap-1 bg-zinc-900 rounded-md p-1">
            <button
              onClick={() => setMode("query")}
              className={cn(
                "px-3 py-1.5 text-xs rounded transition-colors",
                mode === "query"
                  ? "bg-zinc-700 text-white"
                  : "text-zinc-500 hover:text-zinc-300"
              )}
            >
              Query
            </button>
            <button
              onClick={() => setMode("investigate")}
              className={cn(
                "px-3 py-1.5 text-xs rounded transition-colors",
                mode === "investigate"
                  ? "bg-violet-700 text-white"
                  : "text-zinc-500 hover:text-zinc-300"
              )}
            >
              Investigate
            </button>
          </div>
        </div>

        {/* Messages area */}
        <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
          {messages.length === 0 && (
            <div className="flex flex-col items-center justify-center h-full text-center">
              <div className="text-4xl mb-4 opacity-30">
                {mode === "investigate" ? "🔍" : "💬"}
              </div>
              <h2 className="text-lg font-medium text-zinc-300 mb-2">
                {mode === "investigate"
                  ? "Investigation Mode"
                  : "Ask a question"}
              </h2>
              <p className="text-sm text-zinc-500 max-w-md mb-6">
                {mode === "investigate"
                  ? "The AI agent will autonomously investigate issues using multiple data sources and tool calls."
                  : "Translate natural language questions into structured queries against your event data."}
              </p>

              {/* Anomaly banner */}
              {anomalies?.anomalies?.length > 0 && (
                <Card className="mb-6 border-yellow-800/50 bg-yellow-950/20 max-w-lg w-full">
                  <div className="p-3">
                    <p className="text-xs text-yellow-400 font-medium mb-1">
                      Anomalies Detected
                    </p>
                    {anomalies.anomalies.slice(0, 3).map((a: { source: string; metric: string; z_score: number }, i: number) => (
                      <p key={i} className="text-xs text-zinc-400">
                        {a.source}: {a.metric} (Z-score: {a.z_score?.toFixed(1)})
                      </p>
                    ))}
                  </div>
                </Card>
              )}

              <div className="grid grid-cols-2 gap-2 max-w-lg w-full">
                {EXAMPLE_QUESTIONS.map((q) => (
                  <button
                    key={q}
                    onClick={() => {
                      setInput(q);
                    }}
                    className="text-left text-xs text-zinc-400 bg-zinc-900 hover:bg-zinc-800 px-3 py-2 rounded-md transition-colors border border-zinc-800"
                  >
                    {q}
                  </button>
                ))}
              </div>
            </div>
          )}

          {messages.map((msg) => (
            <MessageBubble key={msg.id} message={msg} />
          ))}

          {isLoading && (
            <div className="flex items-center gap-2 text-zinc-500 text-sm">
              <div className="flex gap-1">
                <span className="animate-bounce delay-0">.</span>
                <span className="animate-bounce" style={{ animationDelay: "0.1s" }}>.</span>
                <span className="animate-bounce" style={{ animationDelay: "0.2s" }}>.</span>
              </div>
              <span>{mode === "investigate" ? "Investigating" : "Thinking"}</span>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input */}
        <div className="border-t border-zinc-800 px-6 py-4 shrink-0">
          <form onSubmit={handleSubmit} className="flex gap-3">
            <input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder={
                mode === "investigate"
                  ? "Describe the incident to investigate..."
                  : "Ask a question about your events..."
              }
              className="flex-1 px-4 py-2.5 bg-zinc-900 border border-zinc-800 rounded-lg text-sm text-zinc-300 focus:outline-none focus:border-zinc-600 placeholder-zinc-600"
              disabled={isLoading}
            />
            <button
              type="submit"
              disabled={isLoading || !input.trim()}
              className={cn(
                "px-5 py-2.5 rounded-lg text-sm font-medium transition-colors disabled:opacity-50",
                mode === "investigate"
                  ? "bg-violet-600 text-white hover:bg-violet-500"
                  : "bg-blue-600 text-white hover:bg-blue-500"
              )}
            >
              {mode === "investigate" ? "Investigate" : "Ask"}
            </button>
          </form>
        </div>
      </div>
    </AuthGuard>
  );
}

function MessageBubble({ message }: { message: ChatMessage }) {
  const [showSteps, setShowSteps] = useState(false);

  if (message.role === "user") {
    return (
      <div className="flex justify-end">
        <div className="bg-blue-600/20 border border-blue-800/50 rounded-lg px-4 py-2.5 max-w-[70%]">
          <p className="text-sm text-zinc-200">{message.content}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3 max-w-[85%]">
        <div className="text-sm text-zinc-300 whitespace-pre-wrap">
          {message.content}
        </div>

        {message.confidence !== undefined && (
          <div className="mt-2 flex items-center gap-2">
            <Badge
              className={cn(
                "text-xs",
                message.confidence >= 0.8
                  ? "bg-emerald-900/50 text-emerald-400 border-emerald-800"
                  : message.confidence >= 0.5
                    ? "bg-yellow-900/50 text-yellow-400 border-yellow-800"
                    : "bg-red-900/50 text-red-400 border-red-800"
              )}
            >
              Confidence: {(message.confidence * 100).toFixed(0)}%
            </Badge>
          </div>
        )}

        {message.query && (
          <details className="mt-2">
            <summary className="text-xs text-zinc-600 cursor-pointer hover:text-zinc-400">
              View generated query
            </summary>
            <pre className="mt-1 text-xs text-zinc-500 bg-zinc-950 rounded p-2 overflow-x-auto">
              {JSON.stringify(message.query, null, 2)}
            </pre>
          </details>
        )}

        {message.steps && message.steps.length > 0 && (
          <div className="mt-2">
            <button
              onClick={() => setShowSteps(!showSteps)}
              className="text-xs text-zinc-600 hover:text-zinc-400 transition-colors"
            >
              {showSteps ? "Hide" : "Show"} investigation steps ({message.steps.length})
            </button>

            {showSteps && (
              <div className="mt-2 space-y-1.5">
                {message.steps.map((step, i) => (
                  <StepDisplay key={i} step={step} />
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

function StepDisplay({ step }: { step: InvestigationStep }) {
  const [expanded, setExpanded] = useState(false);

  const icon =
    step.type === "thinking"
      ? "💭"
      : step.type === "tool_call"
        ? "🔧"
        : step.type === "tool_result"
          ? "📊"
          : "✅";

  const label =
    step.type === "tool_call"
      ? `${step.tool_name}()`
      : step.type === "tool_result"
        ? `Result: ${step.tool_name}`
        : step.type;

  return (
    <div className="border-l-2 border-zinc-800 pl-3">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex items-center gap-1.5 text-xs text-zinc-500 hover:text-zinc-400"
      >
        <span>{icon}</span>
        <span className="font-mono">{label}</span>
      </button>

      {expanded && step.content && (
        <pre className="mt-1 text-xs text-zinc-600 bg-zinc-950 rounded p-2 overflow-x-auto max-h-48 overflow-y-auto">
          {step.content.length > 2000
            ? step.content.slice(0, 2000) + "\n... (truncated)"
            : step.content}
        </pre>
      )}

      {expanded && step.tool_args && (
        <pre className="mt-1 text-xs text-zinc-600 bg-zinc-950 rounded p-1.5 overflow-x-auto">
          {step.tool_args}
        </pre>
      )}
    </div>
  );
}
