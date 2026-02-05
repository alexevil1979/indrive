/**
 * API client — Auth (8080), Ride (8083), User (8081)
 * All ride/user endpoints require Authorization: Bearer <token>
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
  role: "passenger" | "driver" = "passenger"
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

export async function createRide(
  token: string,
  from: Point,
  to: Point
): Promise<Ride> {
  const res = await fetch(`${config.rideApiUrl}/api/v1/rides`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ from, to }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Create ride failed");
  }
  return res.json();
}

export async function getRide(token: string, rideId: string): Promise<Ride> {
  const res = await fetch(`${config.rideApiUrl}/api/v1/rides/${rideId}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Ride not found");
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

export async function listBids(token: string, rideId: string): Promise<Bid[]> {
  const res = await fetch(
    `${config.rideApiUrl}/api/v1/rides/${rideId}/bids`,
    { headers: authHeaders(token) }
  );
  if (!res.ok) throw new Error("List bids failed");
  const data = await res.json();
  return (data as { bids: Bid[] }).bids ?? [];
}

export async function acceptBid(
  token: string,
  rideId: string,
  bidId: string
): Promise<Ride> {
  const res = await fetch(
    `${config.rideApiUrl}/api/v1/rides/${rideId}/accept`,
    {
      method: "POST",
      headers: authHeaders(token),
      body: JSON.stringify({ bid_id: bidId }),
    }
  );
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Accept bid failed");
  }
  return res.json();
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

// ============ PAYMENTS ============

export const PAYMENT_PROVIDERS = {
  cash: "Наличные",
  tinkoff: "Тинькофф",
  yoomoney: "ЮMoney",
  sber: "Сбербанк",
} as const;

export const PAYMENT_METHODS = {
  cash: "Наличные",
  card: "Карта",
} as const;

export const PAYMENT_STATUSES = {
  pending: "Ожидание",
  processing: "Обработка",
  completed: "Оплачено",
  failed: "Ошибка",
  cancelled: "Отменён",
  refunded: "Возврат",
} as const;

export type PaymentMethod = {
  id: string;
  user_id: string;
  provider: string;
  last4: string;
  brand: string;
  exp_month: number;
  exp_year: number;
  is_default: boolean;
  created_at: string;
};

export type Payment = {
  id: string;
  ride_id: string;
  user_id: string;
  amount: number;
  currency: string;
  method: string;
  provider: string;
  status: string;
  confirm_url?: string;
  fail_reason?: string;
  created_at: string;
  paid_at?: string;
};

export type PaymentIntent = {
  payment_id: string;
  status: string;
  confirm_url?: string;
  requires_action: boolean;
};

export type CreatePaymentInput = {
  ride_id: string;
  amount: number;
  method: "cash" | "card";
  provider?: string;
  payment_method_id?: string;
  save_card?: boolean;
};

export async function getAvailableProviders(token: string): Promise<string[]> {
  const res = await fetch(`${config.paymentApiUrl}/api/v1/payments/providers`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Get providers failed");
  const data = await res.json();
  return data.providers ?? [];
}

export async function createPayment(
  token: string,
  input: CreatePaymentInput
): Promise<PaymentIntent> {
  const res = await fetch(`${config.paymentApiUrl}/api/v1/payments`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Create payment failed");
  }
  return res.json();
}

export async function getPayment(token: string, paymentId: string): Promise<Payment> {
  const res = await fetch(`${config.paymentApiUrl}/api/v1/payments/${paymentId}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Payment not found");
  return res.json();
}

export async function listPayments(token: string, limit = 20): Promise<Payment[]> {
  const res = await fetch(`${config.paymentApiUrl}/api/v1/payments?limit=${limit}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("List payments failed");
  const data = await res.json();
  return data.payments ?? [];
}

export async function listPaymentMethods(token: string): Promise<PaymentMethod[]> {
  const res = await fetch(`${config.paymentApiUrl}/api/v1/payments/methods`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("List payment methods failed");
  const data = await res.json();
  return data.methods ?? [];
}

export async function deletePaymentMethod(token: string, methodId: string): Promise<void> {
  const res = await fetch(`${config.paymentApiUrl}/api/v1/payments/methods/${methodId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Delete payment method failed");
}

export async function setDefaultPaymentMethod(token: string, methodId: string): Promise<void> {
  const res = await fetch(
    `${config.paymentApiUrl}/api/v1/payments/methods/${methodId}/default`,
    {
      method: "POST",
      headers: authHeaders(token),
    }
  );
  if (!res.ok) throw new Error("Set default payment method failed");
}

export async function confirmCashPayment(token: string, rideId: string): Promise<Payment> {
  const res = await fetch(`${config.paymentApiUrl}/api/v1/payments/confirm-cash`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ ride_id: rideId }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Confirm cash payment failed");
  }
  return res.json();
}

// ============ GEOLOCATION ============

export type NearbyDriver = {
  driver_id: string;
  lat: number;
  lng: number;
  distance: number; // meters
};

export async function getNearbyDrivers(
  token: string,
  lat: number,
  lng: number,
  radius: number = 5000 // meters
): Promise<NearbyDriver[]> {
  const res = await fetch(
    `${config.geolocationApiUrl}/api/v1/drivers/nearby?lat=${lat}&lng=${lng}&radius=${radius}`,
    { headers: authHeaders(token) }
  );
  if (!res.ok) return [];
  const data = await res.json();
  return data.drivers ?? [];
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

export type RatingTag = {
  id: string;
  label: string;
};

export async function submitRating(
  token: string,
  rideId: string,
  score: number,
  comment?: string,
  tags?: string[]
): Promise<Rating> {
  const res = await fetch(`${config.rideApiUrl}/api/v1/rides/${rideId}/rating`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ score, comment, tags }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error ?? "Submit rating failed");
  }
  return res.json();
}

export async function getRideRatings(token: string, rideId: string): Promise<Rating[]> {
  const res = await fetch(`${config.rideApiUrl}/api/v1/rides/${rideId}/ratings`, {
    headers: authHeaders(token),
  });
  if (!res.ok) return [];
  const data = await res.json();
  return data.ratings ?? [];
}

export async function getUserRating(
  token: string,
  userId: string,
  role: "passenger" | "driver" = "driver"
): Promise<UserRating> {
  const res = await fetch(`${config.rideApiUrl}/api/v1/users/${userId}/rating?role=${role}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Get user rating failed");
  return res.json();
}

export async function getMyRating(
  token: string,
  role: "passenger" | "driver" = "passenger"
): Promise<UserRating> {
  const res = await fetch(`${config.rideApiUrl}/api/v1/me/rating?role=${role}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Get my rating failed");
  return res.json();
}

export async function getRatingTags(
  token: string,
  role: "passenger" | "driver" = "driver"
): Promise<RatingTag[]> {
  const res = await fetch(`${config.rideApiUrl}/api/v1/ratings/tags?role=${role}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) return [];
  const data = await res.json();
  return data.tags ?? [];
}

// ============ PROMO CODES ============

export type Promo = {
  code: string;
  description: string;
  type: "percent" | "fixed";
  value: number;
  min_order_value: number;
  max_discount: number;
};

export type PromoResult = {
  valid: boolean;
  promo?: Promo;
  discount: number;
  final_price: number;
  error?: string;
};

export async function validatePromo(
  token: string,
  code: string,
  orderAmount: number
): Promise<PromoResult> {
  const res = await fetch(`${config.paymentApiUrl}/api/v1/promos/validate`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ code, order_amount: orderAmount }),
  });
  if (!res.ok) throw new Error("Failed to validate promo");
  return res.json();
}

export async function applyPromo(
  token: string,
  code: string,
  rideId: string,
  orderAmount: number
): Promise<PromoResult> {
  const res = await fetch(`${config.paymentApiUrl}/api/v1/promos/apply`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ code, ride_id: rideId, order_amount: orderAmount }),
  });
  if (!res.ok) throw new Error("Failed to apply promo");
  return res.json();
}

export async function getActivePromos(token: string): Promise<Promo[]> {
  const res = await fetch(`${config.paymentApiUrl}/api/v1/promos`, {
    headers: authHeaders(token),
  });
  if (!res.ok) return [];
  const data = await res.json();
  return data.promos ?? [];
}
