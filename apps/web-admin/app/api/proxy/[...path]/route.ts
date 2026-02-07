import { NextRequest, NextResponse } from "next/server";

/**
 * Server-side API proxy — adds ADMIN_JWT to requests.
 * Client calls /api/proxy/v1/admin/settings → proxy forwards to USER_API/api/v1/admin/settings with auth header.
 */

// Internal URLs for server-side proxy (NOT public URLs to avoid loops)
const INTERNAL_USER = process.env.INTERNAL_USER_URL ?? "http://localhost:9081";
const INTERNAL_RIDE = process.env.INTERNAL_RIDE_URL ?? "http://localhost:9083";
const INTERNAL_PAYMENT = process.env.INTERNAL_PAYMENT_URL ?? "http://localhost:9084";
const INTERNAL_AUTH = process.env.INTERNAL_AUTH_URL ?? "http://localhost:9080";

const SERVICE_MAP: Record<string, string> = {
  // User service routes
  "v1/users": INTERNAL_USER,
  "v1/verification": INTERNAL_USER,
  "v1/admin/verifications": INTERNAL_USER,
  "v1/admin/settings": INTERNAL_USER,
  "v1/admin/documents": INTERNAL_USER,
  // Ride service routes
  "v1/rides": INTERNAL_RIDE,
  "v1/admin/rides": INTERNAL_RIDE,
  "v1/admin/ratings": INTERNAL_RIDE,
  "v1/ratings": INTERNAL_RIDE,
  "v1/bids": INTERNAL_RIDE,
  "v1/user/ratings": INTERNAL_RIDE,
  // Payment service routes
  "v1/payments": INTERNAL_PAYMENT,
  "v1/admin/payments": INTERNAL_PAYMENT,
  "v1/admin/promos": INTERNAL_PAYMENT,
  "v1/promos": INTERNAL_PAYMENT,
  "v1/user/promos": INTERNAL_PAYMENT,
  // Auth service routes
  "v1/admin/users": INTERNAL_AUTH,
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
