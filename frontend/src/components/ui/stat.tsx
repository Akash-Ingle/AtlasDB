import { cn } from "@/lib/utils";

interface StatProps {
  label: string;
  value: string | number;
  change?: string;
  changeType?: "positive" | "negative" | "neutral";
}

export function Stat({ label, value, change, changeType = "neutral" }: StatProps) {
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-950 p-4">
      <p className="text-xs text-zinc-500 uppercase tracking-wider">{label}</p>
      <p className="mt-1 text-2xl font-semibold text-white">{value}</p>
      {change && (
        <p
          className={cn(
            "mt-1 text-xs",
            changeType === "positive" && "text-emerald-400",
            changeType === "negative" && "text-red-400",
            changeType === "neutral" && "text-zinc-500"
          )}
        >
          {change}
        </p>
      )}
    </div>
  );
}
