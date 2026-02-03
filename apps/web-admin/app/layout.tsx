import type { Metadata } from "next";
import "./globals.css";
import { Sidebar } from "@/components/Sidebar";

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
        <Sidebar />
        <main className="flex-1 p-6 lg:p-8">{children}</main>
      </body>
    </html>
  );
}
