# Notification Service (Node.js)

Push (Firebase stub) + in-app chat WebSocket.

## Run locally

1. Start Postgres: `docker compose -f ../../infra/docker-compose.yml up -d postgres`
2. `pnpm install && pnpm dev` (or `npx tsx src/index.ts`)
3. Health: `curl http://localhost:8085/health`

## API

- **POST /api/v1/device-tokens** — register FCM token: `{"token":"...","platform":"android"}`. Header `X-User-Id` or body `user_id`.
- **POST /api/v1/notifications/send** — stub send: `{"user_id":"...","title":"...","body":"..."}`.
- **GET /api/v1/chat/:rideId/messages?limit=50** — chat history.
- **WebSocket /ws/chat?rideId= & userId=** — join room, send `{"type":"message","text":"..."}` to broadcast.

## Push (Firebase)

Set `FIREBASE_CREDENTIALS_JSON` (JSON string of service account key) to enable FCM. Otherwise push is no-op.

## Kafka

Set `KAFKA_BROKERS` and implement consumer in `kafka-consumer.ts` to send push on ride.matched / ride.status.changed.

## Env

- `PORT` (default 8085)
- `PG_DSN` (same as Auth)
- `FIREBASE_CREDENTIALS_JSON` (optional)
- `KAFKA_BROKERS` (optional)
