# AI Context — indrive (ride-hailing monorepo)

Контекст для AI: текущее состояние проекта и следующий шаг.

---

## Текущее состояние (зафиксировано)

### Инфраструктура
- **Monorepo:** Turborepo + pnpm workspaces, Node ≥20, Go 1.23+
- **go.work:** объединяет все Go-модули (services + packages/otel-go)
- **infra/docker-compose.yml:** PostgreSQL 16 (PostGIS), Redis 7, MinIO, Zookeeper, Kafka, **Prometheus**, **Grafana**, **Jaeger**
- **infra/k8s:** namespace.yaml (базовый манифест)
- **infra/prometheus:** prometheus.yml с job'ами для всех сервисов
- **infra/grafana:** provisioning (datasources: Prometheus + Jaeger, дашборд RideHail Overview)

### Observability (внедрено)
- **Структурированные JSON-логи:** slog (Go), pino (Node)
- **OpenTelemetry трассировка:** packages/otel-go/tracing, инициализация во всех Go-сервисах; OTEL_EXPORTER_OTLP_ENDPOINT → Jaeger
- **Prometheus метрики:** `/metrics` во всех сервисах (Go: packages/otel-go/metrics, Node: prom-client)
- **Grafana:** datasources provisioned, базовый дашборд ridehail-overview

### Backend-сервисы (Go)
- **auth** (порт 8080): регистрация, логин, refresh, JWT, миграции, /metrics
- **user** (8081): профили, /api/v1/users/me, /metrics
- **geolocation** (8082): позиция водителя, nearest drivers (Redis GEO), WebSocket /ws/tracking, /metrics
- **ride** (8083): поездки, ставки, Kafka-события, admin /api/v1/admin/rides, /metrics
- **payment** (8084): checkout (cash/card stub), /metrics

### Backend-сервисы (Node.js)
- **notification** (8085): device tokens, push (Firebase stub), чат WebSocket /ws/chat, /metrics (prom-client), pino логи

### Приложения
- **web-admin** (Next.js 15, порт 3000): дашборд, список поездок, пользователи (stub), Tailwind, shadcn/ui
- **mobile-passenger** (Expo): регистрация, табы, поездки, экран поездки, AuthContext, api/config
- **mobile-driver** (Expo): профиль, поездки водителя

### Пакеты (packages)
- **otel-go:** logger, tracing, metrics, middleware — общий observability для Go
- **types:** ride, user
- **ui:** Button, Card, Input, Badge, MapPlaceholder
- **utils:** format, validation
- **config:** общий конфиг

### Документация
- **instructions.md** — главные правила и порядок доменных модулей
- **docs/deploy-vps.md** — установка на VPS (Nginx, домен indrive.1tlt.ru, каталог /ssd/www/indrive)
- **docs/deploy-vps-apache.md** — вариант только с Apache

### Репозиторий
- **origin:** https://github.com/alexevil1979/indrive.git, ветка **main**

---

## Следующий шаг

**Выбор одного из направлений:**

1. **OAuth2 (Google/Yandex/VK)** — реальная авторизация через провайдеров (сейчас stubs 501).
2. **Верификация водителя** — загрузка документов в MinIO, статус верификации в БД.
3. **Платёжные интеграции** — Tinkoff/Sber/YooMoney SDK вместо stub.
4. **E2E / интеграционные тесты** — Docker Compose + тесты на Go и Node.
5. **CI/CD** — GitHub Actions: lint, test, build, push images.

При следующем запросе уточнить, какое направление приоритетно, или продолжить по порядку из instructions.md.

---

*Файл обновляется при смене фазы разработки или по запросу.*
