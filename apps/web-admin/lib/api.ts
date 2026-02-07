/**
 * API client — BFF calls via server-side proxy /api/proxy/...
 * The proxy adds ADMIN_JWT on the server side.
 * For client components, all requests go through /api/proxy/ to avoid exposing tokens.
 */
const PROXY = "/api/proxy";

// Internal URLs for server-side direct access (no loops through external proxy)
const INTERNAL_RIDE = process.env.INTERNAL_RIDE_URL ?? "http://localhost:9083";
const INTERNAL_USER = process.env.INTERNAL_USER_URL ?? "http://localhost:9081";
const INTERNAL_PAYMENT = process.env.INTERNAL_PAYMENT_URL ?? "http://localhost:9084";
const INTERNAL_AUTH = process.env.INTERNAL_AUTH_URL ?? "http://localhost:9080";

// Check if running in browser
const isBrowser = typeof window !== "undefined";

// Get admin token from env (server-side only)
function getAdminToken(): string {
  if (isBrowser) return "";
  return process.env.ADMIN_JWT ?? "";
}

function authHeaders(): HeadersInit {
  const token = getAdminToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

// Resolve base URL: use proxy in browser, internal URL on server
function rideApi(): string {
  return isBrowser ? PROXY : INTERNAL_RIDE;
}
function userApi(): string {
  return isBrowser ? PROXY : INTERNAL_USER;
}
function paymentApi(): string {
  return isBrowser ? PROXY : INTERNAL_PAYMENT;
}
function authApi(): string {
  return isBrowser ? PROXY : INTERNAL_AUTH;
}

// ============ RIDES ============

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

export async function fetchRides(): Promise<Ride[]> {
  const base = rideApi();
  const url = `${base}/api/v1/admin/rides?limit=100`;
  const res = await fetch(url, { headers: authHeaders(), cache: "no-store" });
  if (!res.ok) {
    // Fallback to public endpoint
    const res2 = await fetch(`${base}/api/v1/rides?limit=100`, { cache: "no-store" });
    if (!res2.ok) return [];
    const data = await res2.json();
    return (data as { rides?: Ride[] }).rides ?? [];
  }
  const data = await res.json();
  return (data as { rides?: Ride[] }).rides ?? [];
}

// ============ USERS ============

export type User = {
  id: string;
  email: string;
  phone?: string;
  name?: string;
  role: string;
  verified?: boolean;
  created_at: string;
  updated_at?: string;
};

export async function fetchUsers(): Promise<User[]> {
  try {
    const res = await fetch(`${authApi()}/api/v1/admin/users?limit=100`, {
      headers: authHeaders(),
      cache: "no-store",
    });
    if (!res.ok) return [];
    const data = await res.json();
    return data.users ?? [];
  } catch {
    return [];
  }
}

export async function fetchUser(id: string): Promise<User | null> {
  try {
    const res = await fetch(`${authApi()}/api/v1/admin/users/${id}`, {
      headers: authHeaders(),
      cache: "no-store",
    });
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

// ============ DRIVER VERIFICATIONS ============

export type DriverDocument = {
  id: string;
  user_id: string;
  doc_type: string;
  file_name: string;
  file_size: number;
  content_type: string;
  storage_url: string;
  status: string;
  reject_reason?: string;
  uploaded_at: string;
  reviewed_at?: string;
  reviewed_by?: string;
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
  reviewed_at?: string;
  reviewed_by?: string;
  created_at: string;
  updated_at: string;
};

export async function fetchVerifications(): Promise<DriverVerification[]> {
  try {
    const res = await fetch(`${userApi()}/api/v1/admin/verifications?limit=100`, {
      headers: authHeaders(),
      cache: "no-store",
    });
    if (!res.ok) return [];
    const data = await res.json();
    return data.verifications ?? data ?? [];
  } catch {
    return [];
  }
}

export async function fetchVerification(id: string): Promise<DriverVerification | null> {
  try {
    const res = await fetch(`${userApi()}/api/v1/admin/verifications/${id}`, {
      headers: authHeaders(),
      cache: "no-store",
    });
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

export async function reviewVerification(
  id: string,
  approved: boolean,
  rejectReason?: string
): Promise<boolean> {
  try {
    const res = await fetch(`${userApi()}/api/v1/admin/verifications/${id}/review`, {
      method: "POST",
      headers: { ...authHeaders(), "Content-Type": "application/json" },
      body: JSON.stringify({ approved, reject_reason: rejectReason }),
    });
    return res.ok;
  } catch {
    return false;
  }
}

export async function reviewDocument(
  id: string,
  approved: boolean,
  rejectReason?: string
): Promise<boolean> {
  try {
    const res = await fetch(`${userApi()}/api/v1/admin/documents/${id}/review`, {
      method: "POST",
      headers: { ...authHeaders(), "Content-Type": "application/json" },
      body: JSON.stringify({ approved, reject_reason: rejectReason }),
    });
    return res.ok;
  } catch {
    return false;
  }
}

// ============ PAYMENTS ============

export type Payment = {
  id: string;
  ride_id: string;
  user_id: string;
  amount: number;
  currency: string;
  method: string;
  provider: string;
  status: string;
  external_id?: string;
  confirm_url?: string;
  description?: string;
  fail_reason?: string;
  refunded_at?: string;
  paid_at?: string;
  created_at: string;
  updated_at: string;
};

export type PaymentStats = {
  total_count: number;
  total_amount: number;
  completed_count: number;
  completed_amount: number;
  pending_count: number;
  failed_count: number;
  refunded_count: number;
};

export async function fetchPayments(): Promise<Payment[]> {
  try {
    const res = await fetch(`${paymentApi()}/api/v1/admin/payments?limit=100`, {
      headers: authHeaders(),
      cache: "no-store",
    });
    if (!res.ok) return [];
    const data = await res.json();
    return data.payments ?? data ?? [];
  } catch {
    return [];
  }
}

export async function fetchPayment(id: string): Promise<Payment | null> {
  try {
    const res = await fetch(`${paymentApi()}/api/v1/payments/${id}`, {
      headers: authHeaders(),
      cache: "no-store",
    });
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

export async function refundPayment(
  id: string,
  amount?: number,
  reason?: string
): Promise<{ success: boolean; error?: string }> {
  try {
    const res = await fetch(`${paymentApi()}/api/v1/payments/${id}/refund`, {
      method: "POST",
      headers: { ...authHeaders(), "Content-Type": "application/json" },
      body: JSON.stringify({ amount, reason }),
    });
    if (!res.ok) {
      const data = await res.json();
      return { success: false, error: data.error ?? "Refund failed" };
    }
    return { success: true };
  } catch {
    return { success: false, error: "Network error" };
  }
}

// ============ RATINGS ============

export type Rating = {
  id: string;
  ride_id: string;
  from_user_id: string;
  to_user_id: string;
  role: "passenger" | "driver";
  score: number;
  comment?: string;
  tags?: string[];
  created_at: string;
};

export type RatingsResponse = {
  ratings: Rating[];
  total: number;
  limit: number;
  offset: number;
};

export type UserRating = {
  user_id: string;
  role: string;
  average_score: number;
  total_ratings: number;
  score_5_count: number;
  score_4_count: number;
  score_3_count: number;
  score_2_count: number;
  score_1_count: number;
};

export async function fetchRatings(
  limit = 20,
  offset = 0
): Promise<RatingsResponse> {
  const res = await fetch(
    `${rideApi()}/api/v1/admin/ratings?limit=${limit}&offset=${offset}`,
    { headers: authHeaders() }
  );
  if (!res.ok) {
    throw new Error("Failed to fetch ratings");
  }
  return res.json();
}

export async function fetchUserRating(
  userId: string,
  role: "passenger" | "driver" = "driver"
): Promise<UserRating> {
  const res = await fetch(
    `${rideApi()}/api/v1/users/${userId}/rating?role=${role}`,
    { headers: authHeaders() }
  );
  if (!res.ok) {
    throw new Error("Failed to fetch user rating");
  }
  return res.json();
}

export const TAG_LABELS: Record<string, string> = {
  polite: "Вежливый",
  clean_car: "Чистая машина",
  safe_driving: "Безопасное вождение",
  fast: "Быстрая поездка",
  good_music: "Хорошая музыка",
  comfortable: "Комфортно",
  on_time: "Вовремя",
  professional: "Профессиональный",
  friendly: "Дружелюбный",
  respectful: "Уважительный",
  clean: "Аккуратный",
};

// ============ PROMO CODES ============

export type Promo = {
  id: string;
  code: string;
  description: string;
  type: "percent" | "fixed";
  value: number;
  min_order_value: number;
  max_discount: number;
  usage_limit: number;
  usage_count: number;
  per_user_limit: number;
  is_active: boolean;
  starts_at: string;
  expires_at: string;
  created_at: string;
  updated_at: string;
};

export type PromosResponse = {
  promos: Promo[];
  total: number;
  limit: number;
  offset: number;
};

export type CreatePromoInput = {
  code: string;
  description?: string;
  type: "percent" | "fixed";
  value: number;
  min_order_value?: number;
  max_discount?: number;
  usage_limit?: number;
  per_user_limit?: number;
  starts_at?: string;
  expires_at?: string;
};

export type UpdatePromoInput = {
  description?: string;
  type?: "percent" | "fixed";
  value?: number;
  min_order_value?: number;
  max_discount?: number;
  usage_limit?: number;
  per_user_limit?: number;
  is_active?: boolean;
  starts_at?: string;
  expires_at?: string;
};

export async function fetchPromos(
  limit = 20,
  offset = 0,
  activeOnly = false
): Promise<PromosResponse> {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });
  if (activeOnly) params.append("active_only", "true");

  const res = await fetch(`${paymentApi()}/api/v1/admin/promos?${params}`, {
    headers: authHeaders(),
  });
  if (!res.ok) throw new Error("Failed to fetch promos");
  return res.json();
}

export async function fetchPromo(id: string): Promise<Promo> {
  const res = await fetch(`${paymentApi()}/api/v1/admin/promos/${id}`, {
    headers: authHeaders(),
  });
  if (!res.ok) throw new Error("Failed to fetch promo");
  return res.json();
}

export async function createPromo(data: CreatePromoInput): Promise<Promo> {
  const res = await fetch(`${paymentApi()}/api/v1/admin/promos`, {
    method: "POST",
    headers: { ...authHeaders(), "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to create promo");
  return res.json();
}

export async function updatePromo(
  id: string,
  data: UpdatePromoInput
): Promise<Promo> {
  const res = await fetch(`${paymentApi()}/api/v1/admin/promos/${id}`, {
    method: "PUT",
    headers: { ...authHeaders(), "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to update promo");
  return res.json();
}

export async function deletePromo(id: string): Promise<void> {
  const res = await fetch(`${paymentApi()}/api/v1/admin/promos/${id}`, {
    method: "DELETE",
    headers: authHeaders(),
  });
  if (!res.ok) throw new Error("Failed to delete promo");
}

// ============ APP SETTINGS ============

export type MapProvider = "google" | "yandex";

export type AppSettings = {
  id: string;
  map_provider: MapProvider;
  google_maps_api_key?: string;
  yandex_maps_api_key?: string;
  default_language: string;
  default_currency: string;
  updated_at: string;
  updated_by?: string;
};

export type UpdateSettingsInput = {
  map_provider: MapProvider;
  google_maps_api_key?: string;
  yandex_maps_api_key?: string;
  default_language?: string;
  default_currency?: string;
};

export async function fetchSettings(): Promise<AppSettings> {
  const res = await fetch(`${userApi()}/api/v1/admin/settings`, {
    headers: authHeaders(),
  });
  if (!res.ok) throw new Error("Failed to fetch settings");
  return res.json();
}

export async function updateSettings(data: UpdateSettingsInput): Promise<void> {
  const res = await fetch(`${userApi()}/api/v1/admin/settings`, {
    method: "PUT",
    headers: { ...authHeaders(), "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to update settings");
}

// ============ HELPERS ============

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatCurrency(amount: number, currency = "RUB"): string {
  return new Intl.NumberFormat("ru-RU", {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(amount);
}

export function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    pending: "bg-yellow-100 text-yellow-800",
    processing: "bg-blue-100 text-blue-800",
    completed: "bg-green-100 text-green-800",
    approved: "bg-green-100 text-green-800",
    failed: "bg-red-100 text-red-800",
    rejected: "bg-red-100 text-red-800",
    cancelled: "bg-gray-100 text-gray-800",
    refunded: "bg-purple-100 text-purple-800",
    in_progress: "bg-blue-100 text-blue-800",
    matched: "bg-indigo-100 text-indigo-800",
  };
  return colors[status] ?? "bg-gray-100 text-gray-800";
}

export function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: "Ожидает",
    processing: "В обработке",
    completed: "Завершён",
    approved: "Одобрено",
    failed: "Ошибка",
    rejected: "Отклонено",
    cancelled: "Отменён",
    refunded: "Возврат",
    in_progress: "В процессе",
    matched: "Назначен",
  };
  return labels[status] ?? status;
}
