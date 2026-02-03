"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const nav = [
  { href: "/", label: "Дашборд" },
  { href: "/rides", label: "Поездки" },
  { href: "/users", label: "Пользователи" },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-56 border-r border-border bg-muted/30 flex flex-col">
      <div className="p-4 border-b border-border">
        <h1 className="font-semibold text-lg">RideHail Admin</h1>
        <p className="text-xs text-muted-foreground">v0.1</p>
      </div>
      <nav className="p-2 flex-1">
        {nav.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              "block px-3 py-2 rounded-md text-sm font-medium transition-colors",
              pathname === item.href
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            )}
          >
            {item.label}
          </Link>
        ))}
      </nav>
    </aside>
  );
}
