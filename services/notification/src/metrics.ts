/**
 * Prometheus metrics for notification service.
 */
import promClient from "prom-client";

const register = new promClient.Registry();

// Default metrics (GC, memory, event loop, etc.)
promClient.collectDefaultMetrics({ register });

// Custom metrics
export const httpRequestsTotal = new promClient.Counter({
  name: "ridehail_notification_http_requests_total",
  help: "Total HTTP requests",
  labelNames: ["method", "path", "status"],
  registers: [register],
});

export const httpRequestDuration = new promClient.Histogram({
  name: "ridehail_notification_http_request_duration_seconds",
  help: "HTTP request duration in seconds",
  labelNames: ["method", "path", "status"],
  buckets: [0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10],
  registers: [register],
});

export const wsConnectionsActive = new promClient.Gauge({
  name: "ridehail_notification_ws_connections_active",
  help: "Active WebSocket connections",
  registers: [register],
});

export const errorsTotal = new promClient.Counter({
  name: "ridehail_notification_errors_total",
  help: "Total errors",
  labelNames: ["type"],
  registers: [register],
});

export { register };
