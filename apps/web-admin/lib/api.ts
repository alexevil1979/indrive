/**
 * API client — BFF calls to Ride/Auth/User services
 * Set RIDE_API_URL; for admin list set ADMIN_JWT (JWT with role=admin from Auth).
 */
const RIDE_API = process.env.NEXT_PUBLIC_RIDE_API_URL ?? "http://localhost:8083";

export type Ride = {
  id: string;
  passenger_id: string;
  driver_id?: string;
  status: string;
  from: { lat: number; lng: number; address?: string };
  to: { lat: number; lng: number; address?: string };
  price?: number;
  created_at: string;
  updated_at: string;
};

/** Fetch rides: use admin endpoint if ADMIN_JWT set, else public (may return 401/empty) */
export async function fetchRides(): Promise<Ride[]> {
  const adminToken = process.env.ADMIN_JWT;
  const url = adminToken
    ? `${RIDE_API}/api/v1/admin/rides?limit=100`
    : `${RIDE_API}/api/v1/rides?limit=100`;
  const headers: HeadersInit = {};
  if (adminToken) headers["Authorization"] = `Bearer ${adminToken}`;
  const res = await fetch(url, { headers, cache: "no-store" });
  if (!res.ok) return [];
  const data = await res.json();
  return (data as { rides?: Ride[] }).rides ?? [];
}

/** Stub: list users (Auth/User service would expose admin endpoint) */
export type UserStub = { id: string; email: string; role: string; created_at?: string };

export async function fetchUsersStub(_token?: string): Promise<UserStub[]> {
  return [];
}
