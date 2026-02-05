// Backend server IP with custom ports (to avoid conflicts)
const BASE_URL = "http://192.168.1.121";

export const config = {
  apiBaseUrl:
    (typeof process !== "undefined" && process.env?.EXPO_PUBLIC_API_URL) ||
    `${BASE_URL}:9080`,
  rideApiUrl:
    (typeof process !== "undefined" && process.env?.EXPO_PUBLIC_RIDE_API_URL) ||
    `${BASE_URL}:9083`,
  userApiUrl:
    (typeof process !== "undefined" && process.env?.EXPO_PUBLIC_USER_API_URL) ||
    `${BASE_URL}:9081`,
  paymentApiUrl:
    (typeof process !== "undefined" && process.env?.EXPO_PUBLIC_PAYMENT_API_URL) ||
    `${BASE_URL}:9084`,
  geolocationApiUrl:
    (typeof process !== "undefined" && process.env?.EXPO_PUBLIC_GEOLOCATION_API_URL) ||
    `${BASE_URL}:9082`,
  notificationApiUrl:
    (typeof process !== "undefined" && process.env?.EXPO_PUBLIC_NOTIFICATION_API_URL) ||
    `${BASE_URL}:9085`,
} as const;
