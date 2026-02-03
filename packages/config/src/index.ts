/**
 * @ridehail/config — env-based config (feature flags, URLs)
 */
export const config = {
  apiBaseUrl: process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080",
  wsUrl: process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8081",
  featureBidding: process.env.FEATURE_BIDDING !== "false",
} as const;
