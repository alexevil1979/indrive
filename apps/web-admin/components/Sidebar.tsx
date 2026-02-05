"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const nav = [
  { href: "/", label: "Дашборд", icon: "📊" },
  { href: "/rides", label: "Поездки", icon: "🚗" },
  { href: "/users", label: "Пользователи", icon: "👥" },
  { href: "/verifications", label: "Верификация", icon: "✅" },
  { href: "/payments", label: "Платежи", icon: "💳" },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-56 border-r border-border bg-muted/30 flex flex-col">
      <div className="p-4 border-b border-border">
        <h1 className="font-semibold text-lg">RideHail Admin</h1>
        <p className="text-xs text-muted-foreground">v0.2</p>
      </div>
      <nav className="p-2 flex-1 space-y-1">
        {nav.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              "flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors",
              pathname === item.href
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            )}
          >
            <span>{item.icon}</span>
            {item.label}
          </Link>
        ))}
      </nav>
      <div className="p-4 border-t border-border">
        <p className="text-xs text-muted-foreground">
          API: {process.env.NEXT_PUBLIC_RIDE_API_URL ? "connected" : "localhost"}
        </p>
      </div>
    </aside>
  );
}
