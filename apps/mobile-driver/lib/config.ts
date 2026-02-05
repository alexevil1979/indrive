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
} as const;
