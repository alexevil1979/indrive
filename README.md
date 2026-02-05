# indrive — ride-hailing monorepo

inDrive-подобный сервис заказа поездок с торгами по цене. Monorepo на Go + TypeScript/Node + React Native + Next.js.

## Стек

| Компонент       | Технология                                      |
|-----------------|------------------------------------------------|
| Backend         | Go 1.23+ (auth с OAuth2, user с верификацией, geolocation, ride, payment с Tinkoff/YooMoney/Sber) |
| Realtime        | Node.js (notification: push, chat WebSocket)    |
| Mobile          | React Native / Expo (карты, геолокация, платежи, верификация, push, чат, **рейтинги**) |
| Web Admin       | Next.js 15, Tailwind, shadcn/ui (верификация, платежи, пользователи) |
| DB              | PostgreSQL 16 + PostGIS                        |
| Cache / GEO     | Redis 7                                         |
| Broker          | Kafka (Confluent)                               |
| Storage         | MinIO (S3-compatible)                           |
| Observability   | Prometheus, Grafana, Jaeger (OpenTelemetry)     |
| Monorepo        | Turborepo + pnpm workspaces                     |

## Структура

```
indrive/
├── apps/
│   ├── mobile-driver/       # Expo — водитель
│   ├── mobile-passenger/    # Expo — пассажир
│   └── web-admin/           # Next.js — админка
├── services/
│   ├── auth/                # Go — регистрация, JWT, OAuth2
│   ├── user/                # Go — профили, верификация водителя (MinIO)
│   ├── geolocation/         # Go — трекинг, Redis GEO
│   ├── ride/                # Go — поездки, ставки, Kafka
│   ├── payment/             # Go — платежи (Tinkoff, YooMoney, Sber)
│   └── notification/        # Node — push, chat
├── packages/
│   ├── otel-go/             # Go — observability (logger, tracing, metrics)
│   ├── types/               # TS — общие типы
│   ├── ui/                  # TS — UI-компоненты
│   ├── utils/               # TS — утилиты
│   └── config/              # TS — конфиг
├── infra/
│   ├── docker-compose.yml   # Postgres, Redis, MinIO, Kafka, Prometheus, Grafana, Jaeger
│   ├── prometheus/          # prometheus.yml
│   ├── grafana/             # provisioning (datasources, dashboards)
│   └── k8s/                 # namespace.yaml
├── docs/
│   ├── deploy-vps.md        # Установка на VPS (Nginx)
│   └── deploy-vps-apache.md # Установка на VPS (Apache)
├── go.work                  # Go workspaces
├── turbo.json
├── pnpm-workspace.yaml
└── package.json
```

## Быстрый старт (локально)

### 1. Зависимости

- Node.js ≥20, pnpm 9+
- Go 1.23+
- Docker, Docker Compose

### 2. Инфраструктура

```bash
docker compose -f infra/docker-compose.yml up -d
```

Запускает: Postgres (5432), Redis (6379), MinIO (9000/9001), Zookeeper (2181), Kafka (9092), Prometheus (9090), Grafana (3001), Jaeger (16686).

### 3. Backend-сервисы

Каждый сервис можно запустить из своей папки:

```bash
# Auth (порт 8080)
cd services/auth && go run .

# User (порт 8081)
cd services/user && PORT=8081 go run .

# Geolocation (порт 8082)
cd services/geolocation && PORT=8082 go run .

# Ride (порт 8083)
cd services/ride && PORT=8083 go run .

# Payment (порт 8084)
cd services/payment && PORT=8084 go run .

# Notification (порт 8085)
cd services/notification && pnpm install && pnpm dev
```

### 4. Web Admin

```bash
pnpm install
pnpm build
cd apps/web-admin && pnpm start
```

Открыть http://localhost:3000

### 5. Observability

- **Prometheus:** http://localhost:9090
- **Grafana:** http://localhost:3001 (admin / admin)
- **Jaeger UI:** http://localhost:16686

Для трассировки установите `OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318` в переменных окружения сервисов.

## Переменные окружения

| Переменная                   | По умолчанию                                            | Описание                        |
|-----------------------------|---------------------------------------------------------|---------------------------------|
| `PORT`                      | 8080–8085                                               | Порт сервиса                    |
| `PG_DSN`                    | postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable | PostgreSQL DSN |
| `REDIS_ADDR`                | localhost:6379                                          | Redis адрес                     |
| `JWT_SECRET`                | dev-secret-change-in-production                         | Секрет JWT                      |
| `KAFKA_BROKERS`             | (пусто — noop)                                          | Kafka брокеры                   |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | (пусто — noop)                                        | OTLP endpoint (Jaeger)          |
| `FIREBASE_CREDENTIALS_JSON` | (пусто — noop)                                          | Firebase credentials            |
| `MINIO_ENDPOINT`            | localhost:9000                                          | MinIO endpoint                  |
| `MINIO_ACCESS_KEY`          | ridehail_minio                                          | MinIO access key                |
| `MINIO_SECRET_KEY`          | ridehail_minio_secret                                   | MinIO secret key                |
| `MINIO_BUCKET`              | ridehail-documents                                      | MinIO bucket name               |
| `MINIO_PUBLIC_URL`          | (пусто)                                                 | Public URL для MinIO            |
| `TINKOFF_TERMINAL_KEY`      | (пусто)                                                 | Tinkoff terminal ID             |
| `TINKOFF_PASSWORD`          | (пусто)                                                 | Tinkoff terminal password       |
| `TINKOFF_TEST_MODE`         | true                                                    | Tinkoff test mode               |
| `YOOMONEY_SHOP_ID`          | (пусто)                                                 | YooMoney shop ID                |
| `YOOMONEY_SECRET_KEY`       | (пусто)                                                 | YooMoney API secret             |
| `YOOMONEY_WEBHOOK_SECRET`   | (пусто)                                                 | YooMoney webhook secret         |
| `SBER_USERNAME`             | (пусто)                                                 | Sberbank API username           |
| `SBER_PASSWORD`             | (пусто)                                                 | Sberbank API password           |
| `SBER_TOKEN`                | (пусто)                                                 | Sberbank API token (alt)        |
| `SBER_TEST_MODE`            | true                                                    | Sberbank test mode              |

## Документация

- [Установка на VPS (Nginx)](docs/deploy-vps.md)
- [Установка на VPS (Apache)](docs/deploy-vps-apache.md)
- [instructions.md](instructions.md) — правила разработки

## Лицензия

MIT
