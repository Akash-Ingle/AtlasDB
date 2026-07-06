import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatTimestamp(ts: string): string {
  const d = new Date(ts);
  return d.toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });
}

export function formatDate(ts: string): string {
  return new Date(ts).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function timeAgo(ts: string): string {
  const seconds = Math.floor((Date.now() - new Date(ts).getTime()) / 1000);
  if (seconds < 60) return `${seconds}s ago`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  return `${Math.floor(seconds / 86400)}d ago`;
}

export const severityColor: Record<string, string> = {
  debug: "text-zinc-400",
  info: "text-blue-400",
  warn: "text-amber-400",
  error: "text-red-400",
  fatal: "text-red-600",
};

export const severityBg: Record<string, string> = {
  debug: "bg-zinc-400/10 text-zinc-400 border-zinc-400/20",
  info: "bg-blue-400/10 text-blue-400 border-blue-400/20",
  warn: "bg-amber-400/10 text-amber-400 border-amber-400/20",
  error: "bg-red-400/10 text-red-400 border-red-400/20",
  fatal: "bg-red-600/10 text-red-500 border-red-500/20",
};
