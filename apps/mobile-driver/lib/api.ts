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

// ============ DRIVER VERIFICATION ============

export type DriverDocument = {
  id: string;
  user_id: string;
  doc_type: string;
  file_name: string;
  file_size: number;
  content_type: string;
  storage_url?: string;
  status: string;
  reject_reason?: string;
  uploaded_at: string;
};

export type DriverVerification = {
  id: string;
  user_id: string;
  status: string;
  license_number?: string;
  vehicle_model?: string;
  vehicle_plate?: string;
  vehicle_year?: number;
  documents?: DriverDocument[];
  reject_reason?: string;
  submitted_at: string;
  created_at: string;
};

export type StartVerificationInput = {
  license_number: string;
  vehicle_model?: string;
  vehicle_plate?: string;
  vehicle_year?: number;
};

export async function startVerification(
  token: string,
  input: StartVerificationInput
): Promise<DriverVerification> {
  const res = await fetch(`${config.userApiUrl}/api/v1/verification`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Start verification failed");
  }
  return res.json();
}

export async function getVerificationStatus(token: string): Promise<DriverVerification | null> {
  const res = await fetch(`${config.userApiUrl}/api/v1/verification`, {
    headers: authHeaders(token),
  });
  if (!res.ok) {
    if (res.status === 404) return null;
    throw new Error("Get verification status failed");
  }
  return res.json();
}

export async function uploadDocument(
  token: string,
  docType: string,
  file: { uri: string; name: string; type: string }
): Promise<DriverDocument> {
  const formData = new FormData();
  formData.append("doc_type", docType);
  formData.append("file", {
    uri: file.uri,
    name: file.name,
    type: file.type,
  } as unknown as Blob);

  const res = await fetch(`${config.userApiUrl}/api/v1/verification/documents`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      // Don't set Content-Type for FormData
    },
    body: formData,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Upload document failed");
  }
  return res.json();
}

export async function listDocuments(token: string): Promise<DriverDocument[]> {
  const res = await fetch(`${config.userApiUrl}/api/v1/verification/documents`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("List documents failed");
  const data = await res.json();
  return Array.isArray(data) ? data : data.documents ?? [];
}

export async function deleteDocument(token: string, docId: string): Promise<void> {
  const res = await fetch(`${config.userApiUrl}/api/v1/verification/documents/${docId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Delete document failed");
}

export const DOC_TYPES = {
  license: "Водительское удостоверение",
  passport: "Паспорт",
  vehicle_reg: "СТС",
  insurance: "ОСАГО",
  photo: "Фото водителя",
  vehicle_photo: "Фото автомобиля",
} as const;

export const VERIFICATION_STATUSES = {
  pending: "На проверке",
  approved: "Одобрено",
  rejected: "Отклонено",
} as const;

// ============ GEOLOCATION ============

export async function updateDriverLocation(
  token: string,
  lat: number,
  lng: number
): Promise<void> {
  const res = await fetch(`${config.geolocationApiUrl}/api/v1/drivers/location`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ lat, lng }),
  });
  if (!res.ok) {
    // Silently fail for location updates
    console.warn("Failed to update location");
  }
}

export async function setDriverOnline(token: string, online: boolean): Promise<void> {
  const res = await fetch(`${config.geolocationApiUrl}/api/v1/drivers/status`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ online }),
  });
  if (!res.ok) {
    console.warn("Failed to set online status");
  }
}
