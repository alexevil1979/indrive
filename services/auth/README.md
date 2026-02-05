# Auth Service (Go)

RideHail auth — Register, Login, Refresh, JWT, **OAuth2 (Google/Yandex/VK)**, PostgreSQL (PostGIS + migrations), Redis (optional).

## Run locally

1. Start infra: `docker compose -f ../../infra/docker-compose.yml up -d postgres redis`
2. `go mod tidy && go run .`
3. Health: `curl http://localhost:8080/health`

## API

### Email/Password Auth

- **Register:** `POST /auth/register`
  ```json
  {"email":"u@example.com","password":"password123","role":"passenger"}
  ```

- **Login:** `POST /auth/login`
  ```json
  {"email":"u@example.com","password":"password123"}
  ```

- **Refresh:** `POST /auth/refresh`
  ```json
  {"refresh_token":"..."}
  ```

### OAuth2

- **List providers:** `GET /auth/oauth/providers` — returns configured providers
- **Redirect to provider:** `GET /auth/oauth/:provider` — redirects to Google/Yandex/VK
- **Callback:** `GET /auth/oauth/:provider/callback` — handles OAuth callback, returns JWT

OAuth flow:
1. Frontend redirects user to `GET /auth/oauth/google` (or yandex/vk)
2. User authenticates with provider
3. Provider redirects to `/auth/oauth/google/callback?code=...&state=...`
4. Service exchanges code for tokens, gets user info, creates/links account
5. Returns JWT (as JSON or redirects to `FRONTEND_URL` with tokens in URL fragment)

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Service port |
| `PG_DSN` | postgres://ridehail:...@localhost:5432/ridehail | PostgreSQL connection |
| `JWT_SECRET` | dev-secret-... | JWT signing secret (change in production!) |
| `REDIS_ADDR` | (empty) | Redis address (optional) |
| `BASE_URL` | http://localhost:8080 | Base URL for OAuth callbacks |
| `FRONTEND_URL` | (empty) | Frontend URL for OAuth redirect |
| `GOOGLE_CLIENT_ID` | (empty) | Google OAuth Client ID |
| `GOOGLE_CLIENT_SECRET` | (empty) | Google OAuth Client Secret |
| `YANDEX_CLIENT_ID` | (empty) | Yandex OAuth Client ID |
| `YANDEX_CLIENT_SECRET` | (empty) | Yandex OAuth Client Secret |
| `VK_CLIENT_ID` | (empty) | VK OAuth Client ID |
| `VK_CLIENT_SECRET` | (empty) | VK OAuth Client Secret |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | (empty) | OpenTelemetry OTLP endpoint |

## OAuth Provider Setup

### Google

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create project → APIs & Services → Credentials → Create OAuth 2.0 Client
3. Add redirect URI: `https://your-domain.com/auth/oauth/google/callback`
4. Set `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`

### Yandex

1. Go to [Yandex OAuth](https://oauth.yandex.ru/)
2. Create app → Web services
3. Add redirect URI: `https://your-domain.com/auth/oauth/yandex/callback`
4. Set `YANDEX_CLIENT_ID` and `YANDEX_CLIENT_SECRET`

### VK

1. Go to [VK Developers](https://vk.com/apps?act=manage)
2. Create app → Website
3. Add redirect URI: `https://your-domain.com/auth/oauth/vk/callback`
4. Set `VK_CLIENT_ID` and `VK_CLIENT_SECRET`

## Database

OAuth accounts are stored in `oauth_accounts` table:
- Links external OAuth providers to internal users
- Supports multiple providers per user
- Stores access/refresh tokens (encrypted in production)

Migration: `006_oauth_accounts.up.sql`
