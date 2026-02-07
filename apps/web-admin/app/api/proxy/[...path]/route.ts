import { NextRequest, NextResponse } from "next/server";

/**
 * Server-side API proxy — adds auth token from cookie (or ADMIN_JWT env fallback).
 * Client calls /api/proxy/api/v1/admin/settings → proxy forwards to internal service with auth header.
 * On 401, attempts token refresh automatically.
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
  const sorted = Object.keys(SERVICE_MAP).sort((a, b) => b.length - a.length);
  for (const prefix of sorted) {
    if (path.startsWith(prefix)) {
      return SERVICE_MAP[prefix];
    }
  }
  return INTERNAL_USER;
}

function getToken(req: NextRequest): string {
  // 1. Cookie token (from login flow)
  const cookieToken = req.cookies.get("access_token")?.value;
  if (cookieToken) return cookieToken;
  // 2. Fallback to env (legacy)
  return process.env.ADMIN_JWT ?? "";
}

async function tryRefresh(req: NextRequest): Promise<string | null> {
  const refreshToken = req.cookies.get("refresh_token")?.value;
  if (!refreshToken) return null;

  try {
    const res = await fetch(`${INTERNAL_AUTH}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    if (!res.ok) return null;
    const data = await res.json();
    return data.access_token ?? null;
  } catch {
    return null;
  }
}

async function doFetch(
  fullUrl: string,
  method: string,
  token: string,
  body?: string
): Promise<Response> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const fetchOptions: RequestInit = { method, headers };
  if (body && method !== "GET" && method !== "HEAD") {
    fetchOptions.body = body;
  }

  return fetch(fullUrl, fetchOptions);
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

  const searchParams = req.nextUrl.searchParams.toString();
  const fullUrl = searchParams ? `${targetUrl}?${searchParams}` : targetUrl;

  const token = getToken(req);

  let body: string | undefined;
  if (req.method !== "GET" && req.method !== "HEAD") {
    body = await req.text();
    if (!body) body = undefined;
  }

  try {
    let res = await doFetch(fullUrl, req.method, token, body);

    // Auto-refresh on 401
    if (res.status === 401 && req.cookies.get("refresh_token")?.value) {
      const newToken = await tryRefresh(req);
      if (newToken) {
        res = await doFetch(fullUrl, req.method, newToken, body);

        if (res.ok || res.status !== 401) {
          const data = await res.text();
          const response = new NextResponse(data, {
            status: res.status,
            headers: { "Content-Type": res.headers.get("Content-Type") ?? "application/json" },
          });
          // Update cookie with new token
          response.cookies.set("access_token", newToken, {
            httpOnly: true,
            secure: process.env.NODE_ENV === "production",
            sameSite: "lax",
            path: "/",
            maxAge: 86400,
          });
          return response;
        }
      }
    }

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
