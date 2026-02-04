# AI Context — indrive (ride-hailing monorepo)

Контекст для AI: текущее состояние проекта и следующий шаг.

---

## Текущее состояние (зафиксировано)

### Инфраструктура
- **Monorepo:** Turborepo + pnpm workspaces, Node ≥20, Go 1.23+
- **infra/docker-compose.yml:** PostgreSQL 16 (PostGIS), Redis 7, MinIO, Zookeeper, Kafka
- **infra/k8s:** namespace.yaml (базовый манифест)

### Backend-сервисы (Go)
- **auth** (порт 8080): регистрация, логин, refresh, JWT, миграции users/profiles, Redis опционально
- **user** (8081): профили, `/api/v1/users/me`, заглушка верификации водителя
- **geolocation** (8082): обновление позиции водителя, nearest drivers (Redis GEO), WebSocket `/ws/tracking`
- **ride** (8083): создание поездки, ставки, принятие ставки, статусы, Kafka-события, admin `/api/v1/admin/rides`
- **payment** (8084): checkout (cash/card stub), JWT

### Backend-сервисы (Node.js)
- **notification** (8085): device tokens, push (Firebase stub), чат (WebSocket `/ws/chat`), история сообщений, Kafka consumer stub для ride events → push

### Приложения
- **web-admin** (Next.js 15, порт 3000): дашборд, список поездок, пользователи (stub), Tailwind, shadcn/ui
- **mobile-passenger** (Expo): регистрация, табы, поездки, экран поездки, AuthContext, api/config
- **mobile-driver** (Expo): то же, профиль, поездки водителя

### Пакеты (packages)
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
- Последний коммит: документация по установке на VPS (Nginx и Apache)

---

## Следующий шаг

**Observability (наблюдаемость):** внедрить из заявленного стека (instructions.md) и довести до рабочего состояния.

1. **Структурированные JSON-логи** во всех Go-сервисах и при необходимости в notification (Node) — единый формат (timestamp, level, service, msg, trace_id при наличии).
2. **OpenTelemetry** — трассировка: инициализация в каждом Go-сервисе и в notification, проброс trace_id в логах и между сервисами (HTTP-заголовки).
3. **Prometheus** — метрики: `/metrics` в auth, user, geolocation, ride, payment, notification (счётчики запросов, латентности, ошибок).
4. **Grafana** — дашборды и алерты по логам/метрикам (опционально в docker-compose или отдельно).

После этого можно переходить к: OAuth2 (Google/Yandex/VK), верификация водителя (загрузка документов), платёжные интеграции (Tinkoff/Sber/YooMoney) или E2E/интеграционным тестам.

---

*Файл обновляется при смене фазы разработки или по запросу.*
