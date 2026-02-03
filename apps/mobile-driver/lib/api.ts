/**
 * API client — Auth (8080), Ride (8083), User (8081) — driver app
 */
import { config } from "./config";

export type TokenResponse = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user_id?: string;
  role?: string;
};

export async function login(email: string, password: string): Promise<TokenResponse> {
  const res = await fetch(`${config.apiBaseUrl}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Login failed");
  }
  return res.json();
}

export async function register(
  email: string,
  password: string,
  role: "passenger" | "driver" = "driver"
): Promise<TokenResponse> {
  const res = await fetch(`${config.apiBaseUrl}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password, role }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Registration failed");
  }
  return res.json();
}

function authHeaders(token: string): HeadersInit {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  };
}

export type Point = { lat: number; lng: number; address?: string };
export type Ride = {
  id: string;
  passenger_id: string;
  driver_id?: string;
  status: string;
  from: Point;
  to: Point;
  price?: number;
  created_at: string;
  updated_at: string;
};
export type Bid = {
  id: string;
  ride_id: string;
  driver_id: string;
  price: number;
  status: string;
  created_at: string;
};

export async function listAvailableRides(token: string, limit = 50): Promise<Ride[]> {
  const res = await fetch(
    `${config.rideApiUrl}/api/v1/rides/available?limit=${limit}`,
    { headers: authHeaders(token) }
  );
  if (!res.ok) throw new Error("List available rides failed");
  const data = await res.json();
  return (data as { rides: Ride[] }).rides ?? [];
}

export async function getRide(token: string, rideId: string): Promise<Ride> {
  const res = await fetch(`${config.rideApiUrl}/api/v1/rides/${rideId}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Ride not found");
  return res.json();
}

export async function placeBid(
  token: string,
  rideId: string,
  price: number
): Promise<Bid> {
  const res = await fetch(
    `${config.rideApiUrl}/api/v1/rides/${rideId}/bids`,
    {
      method: "POST",
      headers: authHeaders(token),
      body: JSON.stringify({ price }),
    }
  );
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Place bid failed");
  }
  return res.json();
}

export async function listMyRides(token: string, limit = 20): Promise<Ride[]> {
  const res = await fetch(
    `${config.rideApiUrl}/api/v1/rides?limit=${limit}`,
    { headers: authHeaders(token) }
  );
  if (!res.ok) throw new Error("List rides failed");
  const data = await res.json();
  return (data as { rides: Ride[] }).rides ?? [];
}

export async function updateRideStatus(
  token: string,
  rideId: string,
  status: string
): Promise<Ride> {
  const res = await fetch(
    `${config.rideApiUrl}/api/v1/rides/${rideId}/status`,
    {
      method: "PATCH",
      headers: authHeaders(token),
      body: JSON.stringify({ status }),
    }
  );
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Update status failed");
  }
  return res.json();
}

export async function createDriverProfile(
  token: string,
  licenseNumber: string
): Promise<unknown> {
  const res = await fetch(`${config.userApiUrl}/api/v1/users/me/driver`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ license_number: licenseNumber }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Create driver profile failed");
  }
  return res.json();
}

export async function getProfile(token: string): Promise<unknown> {
  const res = await fetch(`${config.userApiUrl}/api/v1/users/me`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Get profile failed");
  return res.json();
}
