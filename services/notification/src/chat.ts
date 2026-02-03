/**
 * In-app chat — WebSocket rooms by ride_id, message store
 */
import type { WebSocket } from "ws";
import { pool } from "./db.js";

type Client = { ws: WebSocket; userId: string; rideId: string };

const rooms = new Map<string, Set<Client>>();

function getRoom(rideId: string): Set<Client> {
  let room = rooms.get(rideId);
  if (!room) {
    room = new Set();
    rooms.set(rideId, room);
  }
  return room;
}

export function joinRoom(ws: WebSocket, rideId: string, userId: string): void {
  const client: Client = { ws, userId, rideId };
  getRoom(rideId).add(client);
  ws.on("close", () => {
    getRoom(rideId).delete(client);
    if (getRoom(rideId).size === 0) rooms.delete(rideId);
  });
}

export function broadcastToRide(
  rideId: string,
  payload: { type: string; userId: string; text?: string; messageId?: string; createdAt?: string }
): void {
  const room = rooms.get(rideId);
  if (!room) return;
  const msg = JSON.stringify(payload);
  for (const c of room) {
    if (c.ws.readyState === 1) c.ws.send(msg); // 1 = OPEN
  }
}

export async function saveMessage(
  rideId: string,
  userId: string,
  text: string
): Promise<{ id: string; ride_id: string; user_id: string; text: string; created_at: string }> {
  const r = await pool.query(
    `INSERT INTO chat_messages (ride_id, user_id, text) VALUES ($1, $2, $3)
     RETURNING id, ride_id, user_id, text, created_at`,
    [rideId, userId, text]
  );
  const row = r.rows[0] as { id: string; ride_id: string; user_id: string; text: string; created_at: string };
  return {
    id: row.id,
    ride_id: row.ride_id,
    user_id: row.user_id,
    text: row.text,
    created_at: row.created_at,
  };
}

export async function getMessages(
  rideId: string,
  limit: number
): Promise<{ id: string; ride_id: string; user_id: string; text: string; created_at: string }[]> {
  const r = await pool.query(
    `SELECT id, ride_id, user_id, text, created_at FROM chat_messages
     WHERE ride_id = $1 ORDER BY created_at DESC LIMIT $2`,
    [rideId, limit]
  );
  return r.rows as { id: string; ride_id: string; user_id: string; text: string; created_at: string }[];
}
