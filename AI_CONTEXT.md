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
- **web-admin** (Next.js 15, порт 3000): дашборд, поездки, **пользователи**, **верификация водителей (approve/reject)**, **платежи (refund)**, Tailwind, shadcn/ui
- **mobile-passenger** (Expo): интерактивная карта (react-native-maps), экран оплаты, push-уведомления, **чат с водителем (WebSocket)** в экране поездки
- **mobile-driver** (Expo): карта с заявками, переключатель «на линии», экран верификации, push-уведомления, **чат с пассажиром (WebSocket)** в экране поездки

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

1. ~~**OAuth2 (Google/Yandex/VK)**~~ — ✅ реализовано (auth service)
2. ~~**Верификация водителя**~~ — ✅ реализовано (user service: MinIO upload, domain, repo, usecase, HTTP handlers, admin review)
3. ~~**Платёжные интеграции**~~ — ✅ реализовано (payment service: Tinkoff, YooMoney, Sber gateways, webhooks, refunds, saved cards)
4. ~~**UI для web-admin**~~ — ✅ реализовано (панели: верификация водителей c approve/reject, платежи c refund, пользователи)
5. ~~**Mobile apps (платежи/верификация)**~~ — ✅ реализовано
6. ~~**Карты и геолокация**~~ — ✅ реализовано (react-native-maps + expo-location)
7. ~~**Push-уведомления**~~ — ✅ реализовано (expo-notifications + Firebase Admin SDK)
8. ~~**Чат пассажир-водитель**~~ — ✅ реализовано:
   - WebSocket хук useChat для real-time сообщений
   - Компонент Chat с UI (bubbles, connection status, input)
   - Экран чата /chat/[rideId] в обоих приложениях
   - Кнопка "Чат" на экране поездки (matched/in_progress)
   - История сообщений из notification service
9. **E2E / интеграционные тесты** — Docker Compose + тесты на Go и Node.
10. **CI/CD** — GitHub Actions: lint, test, build, push images.
11. **Рейтинги и отзывы** — система оценки после поездки.

При следующем запросе уточнить, какое направление приоритетно.

---

*Файл обновляется при смене фазы разработки или по запросу.*
