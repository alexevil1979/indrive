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
- **auth** (порт 8080): регистрация, логин, refresh, JWT, **OAuth2 (Google/Yandex/VK)**, миграции (включая oauth_accounts, driver_verifications, driver_documents), /metrics
- **user** (8081): профили, **верификация водителя (MinIO upload)**, /api/v1/users/me, /api/v1/verification/*, /api/v1/admin/verifications/*, /metrics
- **geolocation** (8082): позиция водителя, nearest drivers (Redis GEO), WebSocket /ws/tracking, /metrics
- **ride** (8083): поездки, ставки, Kafka-события, admin /api/v1/admin/rides, /metrics
- **payment** (8084): checkout (**Tinkoff, YooMoney, Sber** + cash), refund, saved cards, webhooks, /metrics

### Backend-сервисы (Node.js)
- **notification** (8085): device tokens, **push (Firebase Admin SDK)**, чат WebSocket /ws/chat, **/api/v1/notifications/{new-bid,ride-status,new-ride,bid-accepted,ride-cancelled}**, /metrics (prom-client), pino логи

### Приложения
- **web-admin** (Next.js 15, порт 3000): дашборд, поездки, пользователи, верификация, платежи, **отзывы и рейтинги**
- **mobile-passenger** (Expo): интерактивная карта, оплата, push, чат, **экран оценки водителя после поездки**
- **mobile-driver** (Expo): карта с заявками, верификация, push, чат, **таб рейтинга**, **экран оценки пассажира**

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

1. ~~**OAuth2 (Google/Yandex/VK)**~~ — ✅ реализовано
2. ~~**Верификация водителя**~~ — ✅ реализовано
3. ~~**Платёжные интеграции**~~ — ✅ реализовано
4. ~~**UI для web-admin**~~ — ✅ реализовано
5. ~~**Mobile apps (платежи/верификация)**~~ — ✅ реализовано
6. ~~**Карты и геолокация**~~ — ✅ реализовано
7. ~~**Push-уведомления**~~ — ✅ реализовано
8. ~~**Чат пассажир-водитель**~~ — ✅ реализовано
9. ~~**Рейтинги и отзывы**~~ — ✅ реализовано:
   - Domain: Rating entity, UserRating aggregate, tags
   - PostgreSQL: migrations с триггером для агрегации
   - Repository + UseCase для рейтингов
   - HTTP handlers: submit, get user/ride ratings, tags
   - mobile-passenger: экран оценки водителя после поездки
   - mobile-driver: таб "Рейтинг" + экран оценки пассажира
   - web-admin: панель отзывов с пагинацией и статистикой
10. **E2E / интеграционные тесты** — Docker Compose + тесты на Go и Node.
11. **CI/CD** — GitHub Actions: lint, test, build, push images.
12. **Real-time трекинг** — отслеживание позиции водителя на карте пассажира.

При следующем запросе уточнить, какое направление приоритетно.

---

*Файл обновляется при смене фазы разработки или по запросу.*
