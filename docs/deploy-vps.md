# Установка indrive на VPS

Документ описывает развёртывание проекта **indrive** на VPS для тестового стенда.

- **Каталог на сервере:** `/ssd/www/indrive`
- **Домен:** `indrive.1tlt.ru`

---

## 1. Требования к серверу

- **ОС:** Ubuntu 22.04 LTS (или 24.04)
- **Память:** минимум 4 GB RAM (рекомендуется 8 GB для Kafka + всех сервисов)
- **Диск:** SSD, свободно ≥20 GB
- **Сеть:** открытые порты 80, 443 (для веб и SSL)

Установлены:

- **Node.js** 20+ (LTS)
- **pnpm** 9.x
- **Go** 1.23+
- **Docker** и **Docker Compose**
- **Nginx**
- **Git**

---

## 2. Установка зависимостей (Ubuntu/Debian)

```bash
# Обновление и базовые пакеты
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl git nginx certbot python3-certbot-nginx

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

Сборка и запуск из корня монорепозитория:

```bash
cd /ssd/www/indrive
pnpm install
pnpm build
cd apps/web-admin
NEXT_PUBLIC_RIDE_API_URL=https://indrive.1tlt.ru pnpm start
```

Админка по умолчанию слушает порт **3000**. Проксирование на домен — через Nginx (см. ниже).

---

## 9. Nginx и домен indrive.1tlt.ru

Убедитесь, что DNS для `indrive.1tlt.ru` указывает на IP вашего VPS.

Создайте конфиг Nginx:

```bash
sudo nano /etc/nginx/sites-available/indrive.1tlt.ru
```

Содержимое (прокси на локальные порты):

```nginx
upstream auth         { server 127.0.0.1:8080; }
upstream user         { server 127.0.0.1:8081; }
upstream geolocation  { server 127.0.0.1:8082; }
upstream ride         { server 127.0.0.1:8083; }
upstream payment      { server 127.0.0.1:8084; }
upstream notification { server 127.0.0.1:8085; }
upstream webadmin     { server 127.0.0.1:3000; }

server {
    listen 80;
    server_name indrive.1tlt.ru;
    # Редирект на HTTPS после certbot:
    # return 301 https://$server_name$request_uri;

    # === Auth Service ===
    # /auth/register, /auth/login, /auth/refresh, /auth/oauth/*
    location /auth/ {
        proxy_pass http://auth/auth/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /health {
        proxy_pass http://auth/health;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
    }

    # === User Service ===
    # /api/v1/users/me, /api/v1/verification/*, /api/v1/admin/verifications/*
    location /api/v1/users/ {
        proxy_pass http://user/api/v1/users/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/verification/ {
        proxy_pass http://user/api/v1/verification/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        client_max_body_size 20M;
    }
    location /api/v1/admin/verifications {
        proxy_pass http://user/api/v1/admin/verifications;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # === Ride Service ===
    # /api/v1/rides/*, /api/v1/bids/*, /api/v1/ratings/*, /api/v1/user/ratings, /api/v1/admin/rides/*
    location /api/v1/rides {
        proxy_pass http://ride/api/v1/rides;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/bids {
        proxy_pass http://ride/api/v1/bids;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/ratings {
        proxy_pass http://ride/api/v1/ratings;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/user/ratings {
        proxy_pass http://ride/api/v1/user/ratings;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/admin/rides {
        proxy_pass http://ride/api/v1/admin/rides;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/admin/ratings {
        proxy_pass http://ride/api/v1/admin/ratings;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # === Payment Service ===
    # /api/v1/payments/*, /api/v1/promos/*, /api/v1/user/promos, /api/v1/admin/promos/*
    location /api/v1/payments {
        proxy_pass http://payment/api/v1/payments;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/promos {
        proxy_pass http://payment/api/v1/promos;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/user/promos {
        proxy_pass http://payment/api/v1/user/promos;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/admin/promos {
        proxy_pass http://payment/api/v1/admin/promos;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /webhooks/ {
        proxy_pass http://payment/webhooks/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # === Geolocation Service ===
    # /api/v1/drivers/*
    location /api/v1/drivers {
        proxy_pass http://geolocation/api/v1/drivers;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # === Notification Service ===
    # /api/v1/device-tokens, /api/v1/notifications/*, /api/v1/chat/*
    location /api/v1/device-tokens {
        proxy_pass http://notification/api/v1/device-tokens;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/notifications/ {
        proxy_pass http://notification/api/v1/notifications/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/chat/ {
        proxy_pass http://notification/api/v1/chat/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # === WebSocket endpoints (Notification Service) ===
    # /ws/chat — чат пассажир-водитель
    location /ws/chat {
        proxy_pass http://notification/ws/chat;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400;
    }
    # /ws/tracking — real-time трекинг позиции водителя
    location /ws/tracking {
        proxy_pass http://notification/ws/tracking;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400;
    }

    # === Админ-панель (Next.js) ===
    location / {
        proxy_pass http://webadmin;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Включите сайт и перезагрузите Nginx:

```bash
sudo ln -sf /etc/nginx/sites-available/indrive.1tlt.ru /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

---

## 10. SSL (Let's Encrypt)

```bash
sudo certbot --nginx -d indrive.1tlt.ru
```

После этого в конфиге Nginx раскомментируйте редирект с HTTP на HTTPS и снова выполните:

```bash
sudo nginx -t && sudo systemctl reload nginx
```

---

## 11. Запуск через systemd (опционально)

Чтобы сервисы поднимались после перезагрузки, создайте unit-файлы.

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

Документ актуален для тестового стенда. Для продакшена дополнительно настройте:
- Резервное копирование БД
- Ротацию логов
- Мониторинг и алерты
- Ограничение доступа к observability-сервисам
