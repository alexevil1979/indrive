/**
 * App config — EXPO_PUBLIC_* in Expo; fallbacks for dev
 */
export const config = {
  apiBaseUrl:
    (typeof process !== "undefined" && process.env?.EXPO_PUBLIC_API_URL) ||
    "http://localhost:8080",
  rideApiUrl:
    (typeof process !== "undefined" && process.env?.EXPO_PUBLIC_RIDE_API_URL) ||
    "http://localhost:8083",
  userApiUrl:
    (typeof process !== "undefined" && process.env?.EXPO_PUBLIC_USER_API_URL) ||
    "http://localhost:8081",
  paymentApiUrl:
    (typeof process !== "undefined" && process.env?.EXPO_PUBLIC_PAYMENT_API_URL) ||
    "http://localhost:8084",
  geolocationApiUrl:
    (typeof process !== "undefined" && process.env?.EXPO_PUBLIC_GEOLOCATION_API_URL) ||
    "http://localhost:8082",
} as const;
