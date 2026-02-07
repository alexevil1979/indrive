import { NextRequest, NextResponse } from "next/server";

/**
 * Server-side API proxy — adds ADMIN_JWT to requests.
 * Client calls /api/proxy/v1/admin/settings → proxy forwards to USER_API/api/v1/admin/settings with auth header.
 */

const SERVICE_MAP: Record<string, string> = {
  // User service routes
  "v1/users": process.env.NEXT_PUBLIC_USER_API_URL ?? "http://localhost:8081",
  "v1/verification": process.env.NEXT_PUBLIC_USER_API_URL ?? "http://localhost:8081",
  "v1/admin/verifications": process.env.NEXT_PUBLIC_USER_API_URL ?? "http://localhost:8081",
  "v1/admin/settings": process.env.NEXT_PUBLIC_USER_API_URL ?? "http://localhost:8081",
  // Ride service routes
  "v1/rides": process.env.NEXT_PUBLIC_RIDE_API_URL ?? "http://localhost:8083",
  "v1/admin/rides": process.env.NEXT_PUBLIC_RIDE_API_URL ?? "http://localhost:8083",
  "v1/admin/ratings": process.env.NEXT_PUBLIC_RIDE_API_URL ?? "http://localhost:8083",
  "v1/ratings": process.env.NEXT_PUBLIC_RIDE_API_URL ?? "http://localhost:8083",
  // Payment service routes
  "v1/payments": process.env.NEXT_PUBLIC_PAYMENT_API_URL ?? "http://localhost:8084",
  "v1/admin/promos": process.env.NEXT_PUBLIC_PAYMENT_API_URL ?? "http://localhost:8084",
  "v1/promos": process.env.NEXT_PUBLIC_PAYMENT_API_URL ?? "http://localhost:8084",
  // Auth service routes
  "v1/admin/users": process.env.NEXT_PUBLIC_AUTH_API_URL ?? "http://localhost:8080",
};

function resolveBackend(path: string): string {
  // Try longest prefix match
  const sorted = Object.keys(SERVICE_MAP).sort((a, b) => b.length - a.length);
  for (const prefix of sorted) {
    if (path.startsWith(prefix)) {
      return SERVICE_MAP[prefix];
    }
  }
  // Default to user service
  return process.env.NEXT_PUBLIC_USER_API_URL ?? "http://localhost:8081";
}

async function handler(
  req: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path } = await params;
  let apiPath = path.join("/");
  // Strip leading "api/" if present (client sends /api/proxy/api/v1/...)
  if (apiPath.startsWith("api/")) {
    apiPath = apiPath.slice(4);
  }
  const backend = resolveBackend(apiPath);
  const targetUrl = `${backend}/api/${apiPath}`;

  // Forward query params
  const searchParams = req.nextUrl.searchParams.toString();
  const fullUrl = searchParams ? `${targetUrl}?${searchParams}` : targetUrl;

  const token = process.env.ADMIN_JWT ?? "";

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  try {
    const fetchOptions: RequestInit = {
      method: req.method,
      headers,
    };

    if (req.method !== "GET" && req.method !== "HEAD") {
      const body = await req.text();
      if (body) {
        fetchOptions.body = body;
      }
    }

    const res = await fetch(fullUrl, fetchOptions);
    const data = await res.text();

    return new NextResponse(data, {
      status: res.status,
      headers: { "Content-Type": res.headers.get("Content-Type") ?? "application/json" },
    });
  } catch (error) {
    console.error("Proxy error:", error);
    return NextResponse.json({ error: "Backend unavailable" }, { status: 502 });
  }
}

export const GET = handler;
export const POST = handler;
export const PUT = handler;
export const DELETE = handler;
export const PATCH = handler;
