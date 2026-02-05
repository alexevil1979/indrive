# Установка indrive на VPS (только Apache)

Вариант развёртывания для VPS, где установлен **только Apache** (без Nginx). Остальная логика — как в [deploy-vps.md](deploy-vps.md).

- **Каталог на сервере:** `/ssd/www/indrive`
- **Домен:** `indrive.1tlt.ru`

---

## 1. Требования к серверу

- **ОС:** Ubuntu 22.04 LTS (или 24.04)
- **Память:** минимум 4 GB RAM (рекомендуется 8 GB для Kafka + всех сервисов)
- **Диск:** SSD, свободно ≥20 GB
- **Сеть:** открытые порты 80, 443

Установлены:

- **Apache 2.4** (с mod_proxy, mod_proxy_http, mod_proxy_wstunnel)
- **Node.js** 20+ (LTS)
- **pnpm** 9.x
- **Go** 1.23+
- **Docker** и **Docker Compose**
- **Git**

---

## 2. Установка зависимостей (Ubuntu/Debian)

```bash
# Обновление и базовые пакеты
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl git apache2 certbot python3-certbot-apache

# Модули Apache для прокси и WebSocket
sudo a2enmod proxy proxy_http proxy_wstunnel rewrite headers ssl
sudo systemctl restart apache2

# Node.js 20 LTS
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# pnpm
sudo npm install -g pnpm@9

# Go 1.23+
wget https://go.dev/dl/go1.23.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc

# Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
# Выйти и зайти в сессию или: newgrp docker
```

---

## 3. Клонирование и каталог

```bash
sudo mkdir -p /ssd/www
sudo chown $USER:$USER /ssd/www
cd /ssd/www
git clone https://github.com/alexevil1979/indrive.git
cd indrive
```

Дальнейшие команды — из `/ssd/www/indrive`, если не указано иное.

---

## 4. Инфраструктура (Docker Compose)

Поднимаем PostgreSQL (PostGIS), Redis, Kafka, Zookeeper, MinIO, Prometheus, Grafana, Jaeger:

```bash
cd /ssd/www/indrive
docker compose -f infra/docker-compose.yml up -d
```

Проверка:

```bash
docker compose -f infra/docker-compose.yml ps
```

Дождаться готовности Postgres и Redis (≈30 сек), затем запускать сервисы приложения.

Порты по умолчанию:

| Сервис     | Порт        | Описание                    |
|------------|-------------|-----------------------------|
| Postgres   | 5432        | PostgreSQL с PostGIS        |
| Redis      | 6379        | Кэш и GEO-индексы           |
| MinIO      | 9000, 9001  | S3-совместимое хранилище    |
| Zookeeper  | 2181        | Координация Kafka           |
| Kafka      | 9092, 29092 | Брокер сообщений            |
| Prometheus | 9090        | Метрики                     |
| Grafana    | 3001        | Дашборды (admin/admin)      |
| Jaeger     | 16686       | Трассировка                 |

---

## 5. Переменные окружения

Создайте общий файл с секретами и DSN (не коммитить в git):

```bash
# /ssd/www/indrive/.env.production

# === Общие ===
# Сгенерируйте надёжный JWT_SECRET: openssl rand -base64 32
export JWT_SECRET="your-production-jwt-secret-change-me"
export PG_DSN="postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export KAFKA_BROKERS="localhost:9092"

# === Observability (опционально) ===
export OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"

# === MinIO (для верификации документов) ===
export MINIO_ENDPOINT="localhost:9000"
export MINIO_ACCESS_KEY="ridehail_minio"
export MINIO_SECRET_KEY="ridehail_minio_secret"
export MINIO_BUCKET="ridehail-documents"
export MINIO_PUBLIC_URL="https://indrive.1tlt.ru/minio"

# === OAuth2 провайдеры (опционально) ===
export GOOGLE_CLIENT_ID=""
export GOOGLE_CLIENT_SECRET=""
export YANDEX_CLIENT_ID=""
export YANDEX_CLIENT_SECRET=""
export VK_CLIENT_ID=""
export VK_CLIENT_SECRET=""
export OAUTH_REDIRECT_BASE_URL="https://indrive.1tlt.ru"

# === Платёжные шлюзы (опционально) ===
# Tinkoff
export TINKOFF_TERMINAL_KEY=""
export TINKOFF_PASSWORD=""
export TINKOFF_TEST_MODE="true"
# YooMoney
export YOOMONEY_SHOP_ID=""
export YOOMONEY_SECRET_KEY=""
export YOOMONEY_WEBHOOK_SECRET=""
# Sberbank
export SBER_USERNAME=""
export SBER_PASSWORD=""
export SBER_TOKEN=""
export SBER_TEST_MODE="true"

# === Firebase Push (опционально) ===
export FIREBASE_CREDENTIALS_JSON=""
```

Для продакшена замените пароли и секреты на реальные значения.

---

## 6. Запуск бэкенд-сервисов (Go)

Все Go-сервисы используют одну БД и один `JWT_SECRET`; Auth должен быть запущен первым (миграции).

Загрузите переменные и запустите в отдельных терминалах или через systemd/supervisor.

### 6.1 Auth (порт 8080)

Регистрация, логин, JWT, OAuth2 (Google/Yandex/VK).

```bash
cd /ssd/www/indrive
source .env.production 2>/dev/null || true
cd services/auth
go mod tidy
go build -o ../../bin/auth . && ../../bin/auth
```

### 6.2 User (порт 8081)

Профили пользователей, верификация водителей (MinIO).

```bash
cd /ssd/www/indrive
source .env.production 2>/dev/null || true
export PORT=8081
cd services/user
go mod tidy
go build -o ../../bin/user . && ../../bin/user
```

### 6.3 Geolocation (порт 8082)

Позиции водителей, поиск ближайших (Redis GEO).

```bash
cd /ssd/www/indrive
source .env.production 2>/dev/null || true
export PORT=8082
cd services/geolocation
go mod tidy
go build -o ../../bin/geolocation . && ../../bin/geolocation
```

### 6.4 Ride (порт 8083)

Поездки, ставки, рейтинги, Kafka-события.

```bash
cd /ssd/www/indrive
source .env.production 2>/dev/null || true
export PORT=8083
cd services/ride
go mod tidy
go build -o ../../bin/ride . && ../../bin/ride
```

### 6.5 Payment (порт 8084)

Платежи (Tinkoff, YooMoney, Sber, наличные), промокоды, возвраты.

```bash
cd /ssd/www/indrive
source .env.production 2>/dev/null || true
export PORT=8084
cd services/payment
go mod tidy
go build -o ../../bin/payment . && ../../bin/payment
```

---

## 7. Notification (Node.js, порт 8085)

Push-уведомления (Firebase), чат и трекинг через WebSocket.

```bash
cd /ssd/www/indrive
source .env.production 2>/dev/null || true
export PORT=8085
cd services/notification
pnpm install
pnpm run build
node dist/index.js  # или: pnpm start
```

При необходимости задайте `PG_DSN`, `KAFKA_BROKERS`, `FIREBASE_CREDENTIALS_JSON` в `.env.production`.

---

## 8. Web Admin (Next.js)

Сборка и запуск:

```bash
cd /ssd/www/indrive
pnpm install
pnpm build
cd apps/web-admin
NEXT_PUBLIC_RIDE_API_URL=https://indrive.1tlt.ru pnpm start
```

Админка по умолчанию слушает порт **3000**. Проксирование на домен — через Apache (см. ниже).

---

## 9. Apache и домен indrive.1tlt.ru

Убедитесь, что DNS для `indrive.1tlt.ru` указывает на IP вашего VPS.

Создайте конфиг:

```bash
sudo nano /etc/apache2/sites-available/indrive.1tlt.ru.conf
```

Содержимое (прокси на локальные порты):

```apache
<VirtualHost *:80>
    ServerName indrive.1tlt.ru

    # Логи
    ErrorLog ${APACHE_LOG_DIR}/indrive-error.log
    CustomLog ${APACHE_LOG_DIR}/indrive-access.log combined

    # Запрет раздачи конфигов
    <FilesMatch "\.(env|git|md)$">
        Require all denied
    </FilesMatch>

    # Увеличенный лимит для загрузки файлов (верификация)
    LimitRequestBody 20971520

    # === Auth Service ===
    # /auth/register, /auth/login, /auth/refresh, /auth/oauth/*
    ProxyPass /auth/ http://127.0.0.1:8080/auth/
    ProxyPassReverse /auth/ http://127.0.0.1:8080/auth/
    ProxyPass /health http://127.0.0.1:8080/health
    ProxyPassReverse /health http://127.0.0.1:8080/health

    # === User Service ===
    # /api/v1/users/me, /api/v1/verification/*, /api/v1/admin/verifications/*
    ProxyPass /api/v1/users/ http://127.0.0.1:8081/api/v1/users/
    ProxyPassReverse /api/v1/users/ http://127.0.0.1:8081/api/v1/users/
    ProxyPass /api/v1/verification/ http://127.0.0.1:8081/api/v1/verification/
    ProxyPassReverse /api/v1/verification/ http://127.0.0.1:8081/api/v1/verification/
    ProxyPass /api/v1/admin/verifications http://127.0.0.1:8081/api/v1/admin/verifications
    ProxyPassReverse /api/v1/admin/verifications http://127.0.0.1:8081/api/v1/admin/verifications

    # === Ride Service ===
    # /api/v1/rides/*, /api/v1/bids/*, /api/v1/ratings/*, /api/v1/admin/rides/*
    ProxyPass /api/v1/rides http://127.0.0.1:8083/api/v1/rides
    ProxyPassReverse /api/v1/rides http://127.0.0.1:8083/api/v1/rides
    ProxyPass /api/v1/bids http://127.0.0.1:8083/api/v1/bids
    ProxyPassReverse /api/v1/bids http://127.0.0.1:8083/api/v1/bids
    ProxyPass /api/v1/ratings http://127.0.0.1:8083/api/v1/ratings
    ProxyPassReverse /api/v1/ratings http://127.0.0.1:8083/api/v1/ratings
    ProxyPass /api/v1/user/ratings http://127.0.0.1:8083/api/v1/user/ratings
    ProxyPassReverse /api/v1/user/ratings http://127.0.0.1:8083/api/v1/user/ratings
    ProxyPass /api/v1/admin/rides http://127.0.0.1:8083/api/v1/admin/rides
    ProxyPassReverse /api/v1/admin/rides http://127.0.0.1:8083/api/v1/admin/rides
    ProxyPass /api/v1/admin/ratings http://127.0.0.1:8083/api/v1/admin/ratings
    ProxyPassReverse /api/v1/admin/ratings http://127.0.0.1:8083/api/v1/admin/ratings

    # === Payment Service ===
    # /api/v1/payments/*, /api/v1/promos/*, /api/v1/admin/promos/*
    ProxyPass /api/v1/payments http://127.0.0.1:8084/api/v1/payments
    ProxyPassReverse /api/v1/payments http://127.0.0.1:8084/api/v1/payments
    ProxyPass /api/v1/promos http://127.0.0.1:8084/api/v1/promos
    ProxyPassReverse /api/v1/promos http://127.0.0.1:8084/api/v1/promos
    ProxyPass /api/v1/user/promos http://127.0.0.1:8084/api/v1/user/promos
    ProxyPassReverse /api/v1/user/promos http://127.0.0.1:8084/api/v1/user/promos
    ProxyPass /api/v1/admin/promos http://127.0.0.1:8084/api/v1/admin/promos
    ProxyPassReverse /api/v1/admin/promos http://127.0.0.1:8084/api/v1/admin/promos
    ProxyPass /webhooks/ http://127.0.0.1:8084/webhooks/
    ProxyPassReverse /webhooks/ http://127.0.0.1:8084/webhooks/

    # === Geolocation Service ===
    # /api/v1/drivers/*
    ProxyPass /api/v1/drivers http://127.0.0.1:8082/api/v1/drivers
    ProxyPassReverse /api/v1/drivers http://127.0.0.1:8082/api/v1/drivers

    # === Notification Service ===
    # /api/v1/device-tokens, /api/v1/notifications/*, /api/v1/chat/*
    ProxyPass /api/v1/device-tokens http://127.0.0.1:8085/api/v1/device-tokens
    ProxyPassReverse /api/v1/device-tokens http://127.0.0.1:8085/api/v1/device-tokens
    ProxyPass /api/v1/notifications/ http://127.0.0.1:8085/api/v1/notifications/
    ProxyPassReverse /api/v1/notifications/ http://127.0.0.1:8085/api/v1/notifications/
    ProxyPass /api/v1/chat/ http://127.0.0.1:8085/api/v1/chat/
    ProxyPassReverse /api/v1/chat/ http://127.0.0.1:8085/api/v1/chat/

    # === WebSocket endpoints (требуется mod_proxy_wstunnel) ===
    # /ws/chat — чат пассажир-водитель
    RewriteEngine On
    RewriteCond %{HTTP:Upgrade} websocket [NC]
    RewriteCond %{HTTP:Connection} upgrade [NC]
    RewriteRule ^/ws/chat$ ws://127.0.0.1:8085/ws/chat [P,L]
    ProxyPass /ws/chat ws://127.0.0.1:8085/ws/chat
    ProxyPassReverse /ws/chat ws://127.0.0.1:8085/ws/chat

    # /ws/tracking — real-time трекинг позиции водителя
    RewriteCond %{HTTP:Upgrade} websocket [NC]
    RewriteCond %{HTTP:Connection} upgrade [NC]
    RewriteRule ^/ws/tracking$ ws://127.0.0.1:8085/ws/tracking [P,L]
    ProxyPass /ws/tracking ws://127.0.0.1:8085/ws/tracking
    ProxyPassReverse /ws/tracking ws://127.0.0.1:8085/ws/tracking

    # === Админ-панель (Next.js) — всё остальное (должно быть последним) ===
    ProxyPass / http://127.0.0.1:3000/
    ProxyPassReverse / http://127.0.0.1:3000/

    # Заголовки для бэкендов
    RequestHeader set X-Forwarded-Proto "http"
    RequestHeader set X-Forwarded-Port "80"
</VirtualHost>
```

Порядок директив важен: более специфичные пути (`/auth/`, `/api/v1/...`, `/ws/...`) — перед общим `ProxyPass /`.

Включите сайт и перезагрузите Apache:

```bash
sudo a2ensite indrive.1tlt.ru.conf
sudo apache2ctl configtest
sudo systemctl reload apache2
```

---

## 10. SSL (Let's Encrypt) с Apache

```bash
sudo certbot --apache -d indrive.1tlt.ru
```

Certbot создаст или изменит VirtualHost для HTTPS (порт 443) и при необходимости настроит редирект с HTTP на HTTPS. После этого перезагрузите Apache:

```bash
sudo systemctl reload apache2
```

Если редирект с HTTP на HTTPS не появился, добавьте в конфиг для `*:80`:

```apache
Redirect permanent / https://indrive.1tlt.ru/
```

В HTTPS-блоке (`<VirtualHost *:443>`) продублируйте те же `ProxyPass`/`ProxyPassReverse`, что и для порта 80, и установите:

```apache
RequestHeader set X-Forwarded-Proto "https"
RequestHeader set X-Forwarded-Port "443"
```

---

## 11. Запуск через systemd (опционально)

Чтобы сервисы поднимались после перезагрузки, заведите unit-файлы (аналогично [deploy-vps.md](deploy-vps.md)).

Пример для Auth:

```ini
# /etc/systemd/system/indrive-auth.service
[Unit]
Description=indrive Auth Service
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/ssd/www/indrive
EnvironmentFile=/ssd/www/indrive/.env.production
ExecStart=/ssd/www/indrive/bin/auth
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Аналогично создаются unit-файлы для остальных сервисов:
- `indrive-user` (PORT=8081)
- `indrive-geolocation` (PORT=8082)
- `indrive-ride` (PORT=8083)
- `indrive-payment` (PORT=8084)
- `indrive-notification` (PORT=8085, ExecStart=node)
- `indrive-webadmin` (порт 3000, ExecStart=pnpm start)

После создания unit-файлов:

```bash
sudo systemctl daemon-reload
sudo systemctl enable indrive-auth indrive-user indrive-ride indrive-geolocation indrive-payment indrive-notification indrive-webadmin
sudo systemctl start indrive-auth indrive-user indrive-ride indrive-geolocation indrive-payment indrive-notification indrive-webadmin
```

---

## 12. Порты сервисов (сводка)

| Сервис       | Порт  | Описание                                      |
|--------------|-------|-----------------------------------------------|
| Auth         | 8080  | Регистрация, JWT, OAuth2                      |
| User         | 8081  | Профили, верификация водителей                |
| Geolocation  | 8082  | Позиции, поиск ближайших                      |
| Ride         | 8083  | Поездки, ставки, рейтинги                     |
| Payment      | 8084  | Платежи, промокоды, возвраты                  |
| Notification | 8085  | Push, чат (WS), трекинг (WS)                  |
| Web Admin    | 3000  | Админ-панель (Next.js)                        |

---

## 13. API endpoints (сводка)

| Путь                           | Сервис       | Описание                        |
|--------------------------------|--------------|---------------------------------|
| `/auth/*`                      | Auth         | Регистрация, логин, OAuth2      |
| `/api/v1/users/*`              | User         | Профили пользователей           |
| `/api/v1/verification/*`       | User         | Верификация водителей           |
| `/api/v1/admin/verifications/*`| User         | Админ: верификация              |
| `/api/v1/drivers/*`            | Geolocation  | Позиции и поиск водителей       |
| `/api/v1/rides/*`              | Ride         | Поездки                         |
| `/api/v1/bids/*`               | Ride         | Ставки                          |
| `/api/v1/ratings/*`            | Ride         | Рейтинги и отзывы               |
| `/api/v1/admin/rides/*`        | Ride         | Админ: поездки                  |
| `/api/v1/payments/*`           | Payment      | Платежи                         |
| `/api/v1/promos/*`             | Payment      | Промокоды                       |
| `/api/v1/admin/promos/*`       | Payment      | Админ: промокоды                |
| `/webhooks/*`                  | Payment      | Вебхуки платёжных систем        |
| `/api/v1/device-tokens`        | Notification | Push-токены                     |
| `/api/v1/notifications/*`      | Notification | Push-уведомления                |
| `/api/v1/chat/*`               | Notification | История чата                    |
| `/ws/chat`                     | Notification | WebSocket: чат                  |
| `/ws/tracking`                 | Notification | WebSocket: трекинг водителя     |

---

## 14. Observability (опционально)

После запуска инфраструктуры доступны:

- **Prometheus:** http://localhost:9090 — метрики сервисов
- **Grafana:** http://localhost:3001 (admin / admin) — дашборды
- **Jaeger:** http://localhost:16686 — трассировка запросов

Для трассировки установите `OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318` в `.env.production`.

Каждый сервис экспортирует метрики на `/metrics`.

---

## 15. Проверка после установки

1. Инфра: `docker compose -f infra/docker-compose.yml ps`
2. Health Auth: `curl http://localhost:8080/health`
3. Health Ride: `curl http://localhost:8083/health`
4. Health Payment: `curl http://localhost:8084/health`
5. В браузере: `https://indrive.1tlt.ru` — должна открыться админ-панель
6. Регистрация тестового пользователя:
   ```bash
   curl -X POST https://indrive.1tlt.ru/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"password123","role":"passenger"}'
   ```

При необходимости откройте порты в файрволе:

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

---

## 16. Мобильные приложения

Для мобильных приложений (Expo) укажите базовый URL API:

```typescript
// apps/mobile-passenger/lib/config.ts
// apps/mobile-driver/lib/config.ts
export const config = {
  apiUrl: "https://indrive.1tlt.ru",
  authApiUrl: "https://indrive.1tlt.ru",
  rideApiUrl: "https://indrive.1tlt.ru",
  paymentApiUrl: "https://indrive.1tlt.ru",
  geolocationApiUrl: "https://indrive.1tlt.ru",
  notificationApiUrl: "https://indrive.1tlt.ru",
  wsUrl: "wss://indrive.1tlt.ru",
};
```

---

Документ актуален для тестового стенда с **только Apache**. Общая последовательность (инфра, сервисы, systemd) совпадает с [deploy-vps.md](deploy-vps.md); отличия — установка Apache вместо Nginx и конфиг VirtualHost с `ProxyPass`/`ProxyPassReverse` и `mod_proxy_wstunnel` для WebSocket.

Для продакшена дополнительно настройте:
- Резервное копирование БД
- Ротацию логов
- Мониторинг и алерты
- Ограничение доступа к observability-сервисам
