/**
 * PostgreSQL — device_tokens, chat_messages
 */
import pg from "pg";

const { Pool } = pg;

const pool = new Pool({
  connectionString:
    process.env.PG_DSN ??
    "postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable",
  max: 10,
});

export async function initDb(): Promise<void> {
  const client = await pool.connect();
  try {
    await client.query(`
      CREATE TABLE IF NOT EXISTS notification_schema_migrations (version TEXT PRIMARY KEY)
    `);
    await client.query(`
      CREATE TABLE IF NOT EXISTS device_tokens (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        user_id TEXT NOT NULL,
        token TEXT NOT NULL UNIQUE,
        platform TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
      )
    `);
    await client.query(`
      CREATE TABLE IF NOT EXISTS chat_messages (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        ride_id TEXT NOT NULL,
        user_id TEXT NOT NULL,
        text TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
      )
    `);
    for (const v of ["006_device_tokens", "007_chat_messages"]) {
      await client.query(
        "INSERT INTO notification_schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING",
        [v]
      );
    }
  } finally {
    client.release();
  }
}

export { pool };
