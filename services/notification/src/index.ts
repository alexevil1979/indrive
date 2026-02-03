/**
 * Notification service — push (Firebase stub) + in-app chat WebSocket
 */
import express from "express";
import { createServer } from "http";
import { WebSocketServer } from "ws";
import { initDb, pool } from "./db.js";
import { initFirebase, registerToken, sendPushToUser } from "./push.js";
import * as chat from "./chat.js";

const PORT = parseInt(process.env.PORT ?? "8085", 10);

await initDb();
initFirebase();

const app = express();
app.use(express.json());

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

// Register device token for push (JWT optional for stub)
app.post("/api/v1/device-tokens", async (req, res) => {
  const userId = (req.headers["x-user-id"] as string) ?? req.body?.user_id ?? "anonymous";
  const token = req.body?.token;
  if (!token) {
    return res.status(400).json({ error: "token required" });
  }
  await registerToken(userId, token, req.body?.platform);
  res.json({ status: "ok" });
});

// Stub: send push to user (internal or test)
app.post("/api/v1/notifications/send", async (req, res) => {
  const { user_id, title, body, data } = req.body ?? {};
  if (!user_id || !title) {
    return res.status(400).json({ error: "user_id and title required" });
  }
  const count = await sendPushToUser(user_id, title, body ?? "", data);
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
const wss = new WebSocketServer({ server, path: "/ws/chat" });

wss.on("connection", (ws, req) => {
  const url = new URL(req.url ?? "", `http://localhost`);
  const rideId = url.searchParams.get("rideId") ?? "";
  const userId = url.searchParams.get("userId") ?? "anonymous";
  if (!rideId) {
    ws.close(4000, "rideId required");
    return;
  }
  chat.joinRoom(ws, rideId, userId);

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
    } catch {
      // ignore
    }
  });
});

server.listen(PORT, () => {
  console.log(`Notification service listening on :${PORT}`);
});
