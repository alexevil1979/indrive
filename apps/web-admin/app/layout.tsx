import type { Metadata } from "next";
import "./globals.css";
import { LayoutShell } from "@/components/LayoutShell";

export const metadata: Metadata = {
  title: "RideHail Admin",
  description: "Admin panel — ride monitoring, users, analytics",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ru">
      <body className="min-h-screen flex">
        <LayoutShell>{children}</LayoutShell>
      </body>
    </html>
  );
}
