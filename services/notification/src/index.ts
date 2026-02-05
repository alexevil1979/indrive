/**
 * Notification service — push (Firebase stub) + in-app chat WebSocket
 * With structured logging (pino) and Prometheus metrics.
 */
import express, { Request, Response, NextFunction } from "express";
import { createServer } from "http";
import { WebSocketServer } from "ws";
import { initDb, pool } from "./db.js";
import { initFirebase, registerToken, sendPushToUser } from "./push.js";
import * as chat from "./chat.js";
import * as tracking from "./tracking.js";
import logger from "./logger.js";
import {
  register,
  httpRequestsTotal,
  httpRequestDuration,
  wsConnectionsActive,
  errorsTotal,
} from "./metrics.js";

const PORT = parseInt(process.env.PORT ?? "8085", 10);

await initDb();
initFirebase();

const app = express();
app.use(express.json());

// Observability middleware
app.use((req: Request, res: Response, next: NextFunction) => {
  const start = process.hrtime.bigint();
  const { method, path } = req;

  res.on("finish", () => {
    const duration = Number(process.hrtime.bigint() - start) / 1e9;
    const status = res.statusCode.toString();
    const statusClass = `${Math.floor(res.statusCode / 100)}xx`;

    httpRequestsTotal.inc({ method, path, status: statusClass });
    httpRequestDuration.observe({ method, path, status: statusClass }, duration);

    // Log non-health endpoints
    if (path !== "/health" && path !== "/ready" && path !== "/metrics") {
      logger.info({
        msg: "http_request",
        method,
        path,
        status: res.statusCode,
        duration_ms: Math.round(duration * 1000),
      });
    }

    if (res.statusCode >= 500) {
      errorsTotal.inc({ type: "http_5xx" });
    }
  });

  next();
});

app.get("/health", (_req, res) => {
  res.json({ status: "ok", service: "notification" });
});

app.get("/ready", async (_req, res) => {
  try {
    await pool.query("SELECT 1");
    res.json({ status: "ready" });
  } catch {
    res.status(503).json({ status: "not ready" });
  }
});

// Prometheus metrics endpoint
app.get("/metrics", async (_req, res) => {
  try {
    res.set("Content-Type", register.contentType);
    res.end(await register.metrics());
  } catch (err) {
    res.status(500).end();
  }
});

// Register device token for push (JWT optional for stub)
app.post("/api/v1/device-tokens", async (req, res) => {
  const userId = (req.headers["x-user-id"] as string) ?? req.body?.user_id ?? "anonymous";
  const token = req.body?.token;
  if (!token) {
    return res.status(400).json({ error: "token required" });
  }
  await registerToken(userId, token, req.body?.platform);
  logger.info({ msg: "device_token_registered", userId, platform: req.body?.platform });
  res.json({ status: "ok" });
});

// Generic: send push to user (internal or test)
app.post("/api/v1/notifications/send", async (req, res) => {
  const { user_id, title, body, data } = req.body ?? {};
  if (!user_id || !title) {
    return res.status(400).json({ error: "user_id and title required" });
  }
  const count = await sendPushToUser(user_id, title, body ?? "", data);
  logger.info({ msg: "push_sent", user_id, count });
  res.json({ sent: count });
});

// Notify passenger about new bid
app.post("/api/v1/notifications/new-bid", async (req, res) => {
  const { passenger_id, ride_id, bid_id, price, driver_name } = req.body ?? {};
  if (!passenger_id || !ride_id) {
    return res.status(400).json({ error: "passenger_id and ride_id required" });
  }
  const title = "Новая ставка! 🚗";
  const body = driver_name 
    ? `${driver_name} предложил ${price} ₽`
    : `Водитель предложил ${price} ₽`;
  const count = await sendPushToUser(passenger_id, title, body, {
    type: "new_bid",
    ride_id,
    bid_id: bid_id ?? "",
    price: String(price ?? ""),
  });
  logger.info({ msg: "new_bid_notification", passenger_id, ride_id, count });
  res.json({ sent: count });
});

// Notify passenger about ride status change
app.post("/api/v1/notifications/ride-status", async (req, res) => {
  const { passenger_id, ride_id, status, driver_name } = req.body ?? {};
  if (!passenger_id || !ride_id || !status) {
    return res.status(400).json({ error: "passenger_id, ride_id and status required" });
  }
  
  const statusMessages: Record<string, { title: string; body: string }> = {
    matched: { title: "Водитель найден! 🎉", body: driver_name ? `${driver_name} принял вашу заявку` : "Водитель принял вашу заявку" },
    driver_arrived: { title: "Водитель прибыл! 📍", body: "Водитель ожидает вас" },
    in_progress: { title: "Поездка началась 🚗", body: "Приятной поездки!" },
    completed: { title: "Поездка завершена ✅", body: "Спасибо за поездку!" },
    cancelled: { title: "Поездка отменена", body: "Водитель отменил поездку" },
  };
  
  const msg = statusMessages[status] ?? { title: "Статус поездки", body: `Статус изменён на: ${status}` };
  const count = await sendPushToUser(passenger_id, msg.title, msg.body, {
    type: "ride_status",
    ride_id,
    status,
  });
  logger.info({ msg: "ride_status_notification", passenger_id, ride_id, status, count });
  res.json({ sent: count });
});

// Notify driver about new ride nearby
app.post("/api/v1/notifications/new-ride", async (req, res) => {
  const { driver_id, ride_id, from_address, to_address } = req.body ?? {};
  if (!driver_id || !ride_id) {
    return res.status(400).json({ error: "driver_id and ride_id required" });
  }
  const title = "Новая заявка поблизости! 📍";
  const body = from_address ? `Откуда: ${from_address}` : "Нажмите, чтобы посмотреть";
  const count = await sendPushToUser(driver_id, title, body, {
    type: "new_ride",
    ride_id,
    from_address: from_address ?? "",
    to_address: to_address ?? "",
  });
  logger.info({ msg: "new_ride_notification", driver_id, ride_id, count });
  res.json({ sent: count });
});

// Notify driver about bid acceptance
app.post("/api/v1/notifications/bid-accepted", async (req, res) => {
  const { driver_id, ride_id, passenger_name, from_address } = req.body ?? {};
  if (!driver_id || !ride_id) {
    return res.status(400).json({ error: "driver_id and ride_id required" });
  }
  const title = "Ваша ставка принята! 🎉";
  const body = from_address 
    ? `Заберите пассажира: ${from_address}`
    : "Откройте приложение для деталей";
  const count = await sendPushToUser(driver_id, title, body, {
    type: "bid_accepted",
    ride_id,
    from_address: from_address ?? "",
  });
  logger.info({ msg: "bid_accepted_notification", driver_id, ride_id, count });
  res.json({ sent: count });
});

// Notify driver about ride cancellation
app.post("/api/v1/notifications/ride-cancelled", async (req, res) => {
  const { driver_id, ride_id, reason } = req.body ?? {};
  if (!driver_id || !ride_id) {
    return res.status(400).json({ error: "driver_id and ride_id required" });
  }
  const title = "Поездка отменена";
  const body = reason ?? "Пассажир отменил поездку";
  const count = await sendPushToUser(driver_id, title, body, {
    type: "ride_cancelled",
    ride_id,
  });
  logger.info({ msg: "ride_cancelled_notification", driver_id, ride_id, count });
  res.json({ sent: count });
});

// Chat history
app.get("/api/v1/chat/:rideId/messages", async (req, res) => {
  const rideId = req.params.rideId;
  const limit = Math.min(parseInt(req.query.limit as string ?? "50", 10), 100);
  const messages = await chat.getMessages(rideId, limit);
  res.json({ messages: messages.reverse() });
});

const server = createServer(app);

// WebSocket chat: ?rideId= & userId= (JWT in production)
const wssChat = new WebSocketServer({ noServer: true });

// WebSocket tracking: driver location streaming
const wssTracking = new WebSocketServer({ noServer: true });

// Handle upgrade requests
server.on("upgrade", (request, socket, head) => {
  const url = new URL(request.url ?? "", `http://localhost`);
  
  if (url.pathname === "/ws/chat") {
    wssChat.handleUpgrade(request, socket, head, (ws) => {
      wssChat.emit("connection", ws, request);
    });
  } else if (url.pathname === "/ws/tracking") {
    wssTracking.handleUpgrade(request, socket, head, (ws) => {
      wssTracking.emit("connection", ws, request);
    });
  } else {
    socket.destroy();
  }
});

// Chat WebSocket handler
wssChat.on("connection", (ws, req) => {
  const url = new URL(req.url ?? "", `http://localhost`);
  const rideId = url.searchParams.get("rideId") ?? "";
  const userId = url.searchParams.get("userId") ?? "anonymous";

  if (!rideId) {
    ws.close(4000, "rideId required");
    return;
  }

  wsConnectionsActive.inc();
  chat.joinRoom(ws, rideId, userId);
  logger.info({ msg: "ws_chat_connected", rideId, userId });

  ws.on("message", async (raw) => {
    try {
      const body = JSON.parse(raw.toString()) as { type: string; text?: string };
      if (body.type === "message" && typeof body.text === "string" && body.text.trim()) {
        const msg = await chat.saveMessage(rideId, userId, body.text.trim());
        chat.broadcastToRide(rideId, {
          type: "message",
          userId,
          text: msg.text,
          messageId: msg.id,
          createdAt: msg.created_at,
        });
      }
    } catch (err) {
      logger.error({ msg: "ws_message_error", error: (err as Error).message });
      errorsTotal.inc({ type: "ws_message" });
    }
  });

  ws.on("close", () => {
    wsConnectionsActive.dec();
    logger.info({ msg: "ws_chat_disconnected", rideId, userId });
  });

  ws.on("error", (err) => {
    logger.error({ msg: "ws_error", error: err.message });
    errorsTotal.inc({ type: "ws_error" });
  });
});

// Tracking WebSocket handler
// Driver: ?role=driver&rideId=&driverId=
// Passenger: ?role=passenger&rideId=&passengerId=
wssTracking.on("connection", (ws, req) => {
  const url = new URL(req.url ?? "", `http://localhost`);
  const role = url.searchParams.get("role") ?? "";
  const rideId = url.searchParams.get("rideId") ?? "";

  if (!rideId) {
    ws.close(4000, "rideId required");
    return;
  }

  wsConnectionsActive.inc();

  if (role === "driver") {
    const driverId = url.searchParams.get("driverId") ?? "";
    if (!driverId) {
      ws.close(4000, "driverId required for driver role");
      return;
    }

    tracking.registerDriver(ws, rideId, driverId);
    logger.info({ msg: "ws_tracking_driver_connected", rideId, driverId });

    ws.on("message", (raw) => {
      try {
        const body = JSON.parse(raw.toString()) as {
          type: string;
          lat?: number;
          lng?: number;
          heading?: number;
          speed?: number;
        };

        if (body.type === "location" && typeof body.lat === "number" && typeof body.lng === "number") {
          tracking.updateDriverLocation(rideId, driverId, {
            lat: body.lat,
            lng: body.lng,
            heading: body.heading,
            speed: body.speed,
            timestamp: Date.now(),
          });
        }
      } catch (err) {
        logger.error({ msg: "ws_tracking_message_error", error: (err as Error).message });
        errorsTotal.inc({ type: "ws_tracking_message" });
      }
    });

  } else if (role === "passenger") {
    const passengerId = url.searchParams.get("passengerId") ?? "";
    if (!passengerId) {
      ws.close(4000, "passengerId required for passenger role");
      return;
    }

    tracking.subscribePassenger(ws, rideId, passengerId);
    logger.info({ msg: "ws_tracking_passenger_connected", rideId, passengerId });

  } else {
    ws.close(4000, "role must be driver or passenger");
    return;
  }

  ws.on("close", () => {
    wsConnectionsActive.dec();
    logger.info({ msg: "ws_tracking_disconnected", rideId, role });
  });

  ws.on("error", (err) => {
    logger.error({ msg: "ws_tracking_error", error: err.message });
    errorsTotal.inc({ type: "ws_tracking_error" });
  });
});

// API to get tracking stats
app.get("/api/v1/tracking/stats", (_req, res) => {
  const stats = tracking.getStats();
  res.json(stats);
});

// API to cleanup tracking for a ride (when completed/cancelled)
app.post("/api/v1/tracking/cleanup", (req, res) => {
  const { ride_id } = req.body ?? {};
  if (!ride_id) {
    return res.status(400).json({ error: "ride_id required" });
  }
  tracking.cleanupRide(ride_id);
  res.json({ status: "ok" });
});

server.listen(PORT, () => {
  logger.info({ msg: "server_started", port: PORT });
});
