"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Radio,
  Search,
  BarChart3,
  Bell,
  BotMessageSquare,
  HeartPulse,
  Settings,
  LogOut,
  Activity,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

const navigation = [
  { name: "Overview", href: "/", icon: LayoutDashboard },
  { name: "Events", href: "/events", icon: Radio },
  { name: "Search", href: "/search", icon: Search },
  { name: "Analytics", href: "/analytics", icon: BarChart3 },
  { name: "Alerts", href: "/alerts", icon: Bell },
  { name: "AI Copilot", href: "/copilot", icon: BotMessageSquare },
  { name: "System", href: "/system", icon: HeartPulse },
  { name: "Settings", href: "/settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const { email, logout } = useAuthStore();

  return (
    <aside className="flex flex-col w-60 bg-zinc-950 border-r border-zinc-800 h-screen sticky top-0">
      {/* Logo */}
      <div className="flex items-center gap-2 px-4 py-5 border-b border-zinc-800">
        <Activity className="h-6 w-6 text-blue-500" />
        <span className="text-lg font-bold text-white tracking-tight">
          AtlasDB
        </span>
      </div>

      {/* Nav */}
      <nav className="flex-1 px-3 py-4 space-y-1">
        {navigation.map((item) => {
          const active =
            item.href === "/"
              ? pathname === "/"
              : pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors",
                active
                  ? "bg-zinc-800 text-white"
                  : "text-zinc-400 hover:bg-zinc-900 hover:text-white"
              )}
            >
              <item.icon className="h-4 w-4" />
              {item.name}
            </Link>
          );
        })}
      </nav>

      {/* User */}
      <div className="border-t border-zinc-800 px-3 py-4">
        <div className="flex items-center justify-between">
          <span className="text-xs text-zinc-500 truncate max-w-[140px]">
            {email}
          </span>
          <button
            onClick={logout}
            className="text-zinc-500 hover:text-white transition-colors"
            title="Logout"
          >
            <LogOut className="h-4 w-4" />
          </button>
        </div>
      </div>
    </aside>
  );
}
