/**
 * Push notifications — Firebase stub (FCM)
 * If FIREBASE_CREDENTIALS_JSON not set, no-op
 */
import admin from "firebase-admin";
import { pool } from "./db.js";

let firebaseApp: admin.app.App | null = null;

export function initFirebase(): void {
  const creds = process.env.FIREBASE_CREDENTIALS_JSON;
  if (!creds) {
    console.warn("FIREBASE_CREDENTIALS_JSON not set — push disabled");
    return;
  }
  try {
    firebaseApp = admin.initializeApp({
      credential: admin.credential.cert(JSON.parse(creds) as admin.ServiceAccount),
    });
  } catch (e) {
    console.warn("Firebase init failed:", e);
  }
}

export async function registerToken(
  userId: string,
  token: string,
  platform?: string
): Promise<void> {
  await pool.query(
    `INSERT INTO device_tokens (user_id, token, platform)
     VALUES ($1, $2, $3)
     ON CONFLICT (token) DO UPDATE SET user_id = $1, platform = $3`,
    [userId, token, platform ?? "unknown"]
  );
}

export async function sendPushToUser(
  userId: string,
  title: string,
  body: string,
  data?: Record<string, string>
): Promise<number> {
  const r = await pool.query(
    "SELECT token FROM device_tokens WHERE user_id = $1",
    [userId]
  );
  const tokens = r.rows.map((row: { token: string }) => row.token);
  if (tokens.length === 0) return 0;
  return sendPushToTokens(tokens, title, body, data);
}

export async function sendPushToTokens(
  tokens: string[],
  title: string,
  body: string,
  data?: Record<string, string>
): Promise<number> {
  if (!firebaseApp) return 0;
  try {
    const message: admin.messaging.MulticastMessage = {
      tokens,
      notification: { title, body },
      data: data ?? {},
    };
    const res = await admin.messaging().sendEachForMulticast(message);
    return res.successCount;
  } catch (e) {
    console.error("FCM send error:", e);
    return 0;
  }
}
