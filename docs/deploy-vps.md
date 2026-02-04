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

Поднимаем PostgreSQL (PostGIS), Redis, Kafka, Zookeeper, MinIO:

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

| Сервис    | Порт  |
|----------|-------|
| Postgres | 5432  |
| Redis    | 6379  |
| MinIO    | 9000, 9001 |
| Zookeeper| 2181  |
| Kafka    | 9092, 29092 |

---

## 5. Переменные окружения

Создайте общий файл с секретами и DSN (не коммитить в git):

```bash
# /ssd/www/indrive/.env.production
# Сгенерируйте надёжный JWT_SECRET: openssl rand -base64 32
export JWT_SECRET="your-production-jwt-secret-change-me"
export PG_DSN="postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export KAFKA_BROKERS="localhost:9092"
```

Для продакшена замените пароль БД и `JWT_SECRET` на свои значения.

---

## 6. Запуск бэкенд-сервисов (Go)

Все Go-сервисы используют одну БД и один `JWT_SECRET`; Auth должен быть запущен первым (миграции пользователей и профилей).

Загрузите переменные и запустите в отдельных терминалах или через systemd/supervisor.

### 6.1 Auth (порт 8080)

```bash
cd /ssd/www/indrive
source .env.production 2>/dev/null || true
cd services/auth
go mod tidy
go build -o ../../bin/auth . && ../../bin/auth
# Или: go run .
```

### 6.2 User (порт 8081)

```bash
cd /ssd/www/indrive
source .env.production 2>/dev/null || true
export PORT=8081
cd services/user
go mod tidy
go build -o ../../bin/user . && ../../bin/user
```

### 6.3 Geolocation (порт 8082)

```bash
cd /ssd/www/indrive
source .env.production 2>/dev/null || true
export PORT=8082
cd services/geolocation
go mod tidy
go build -o ../../bin/geolocation . && ../../bin/geolocation
```

### 6.4 Ride (порт 8083)

```bash
cd /ssd/www/indrive
source .env.production 2>/dev/null || true
export PORT=8083
cd services/ride
go mod tidy
go build -o ../../bin/ride . && ../../bin/ride
```

### 6.5 Payment (порт 8084)

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

```bash
cd /ssd/www/indrive
source .env.production 2>/dev/null || true
export PORT=8085
cd services/notification
pnpm install
pnpm run build  # если есть скрипт build
node dist/index.js  # или: pnpm start / npx tsx src/index.ts
```

При необходимости задайте `PG_DSN`, `KAFKA_BROKERS`, `FIREBASE_CREDENTIALS_JSON` в `.env.production` или в unit-файле.

---

## 8. Web Admin (Next.js)

Сборка и запуск из корня монорепозитория:

```bash
cd /ssd/www/indrive
pnpm install
pnpm build
# Для админки:
cd apps/web-admin
NEXT_PUBLIC_RIDE_API_URL=https://indrive.1tlt.ru pnpm start
```

Или из корня после сборки:

```bash
cd /ssd/www/indrive/apps/web-admin
NEXT_PUBLIC_RIDE_API_URL=https://indrive.1tlt.ru pnpm start
```

Админка по умолчанию слушает порт **3000**. Проксирование на домен — через Nginx (см. ниже).

---

## 9. Nginx и домен indrive.1tlt.ru

Убедитесь, что DNS для `indrive.1tlt.ru` указывает на IP вашего VPS.

Создайте конфиг Nginx (например):

```bash
sudo nano /etc/nginx/sites-available/indrive.1tlt.ru
```

Содержимое (прокси на локальные порты; пути совпадают с API сервисов):

```nginx
upstream auth        { server 127.0.0.1:8080; }
upstream user        { server 127.0.0.1:8081; }
upstream geolocation { server 127.0.0.1:8082; }
upstream ride        { server 127.0.0.1:8083; }
upstream payment     { server 127.0.0.1:8084; }
upstream notification { server 127.0.0.1:8085; }
upstream webadmin    { server 127.0.0.1:3000; }

server {
    listen 80;
    server_name indrive.1tlt.ru;
    # Редирект на HTTPS после certbot:
    # return 301 https://$server_name$request_uri;

    # Auth: /auth/register, /auth/login, /health
    location /auth/ {
        proxy_pass http://auth/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /health { proxy_pass http://auth/health; proxy_http_version 1.1; proxy_set_header Host $host; }

    # User: /api/v1/users/me
    location /api/v1/users/ {
        proxy_pass http://user/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Ride: /api/v1/rides, /api/v1/admin/rides
    location /api/v1/rides { proxy_pass http://ride/; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /api/v1/admin/ {
        proxy_pass http://ride/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Payment: /api/v1/payments
    location /api/v1/payments {
        proxy_pass http://payment/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Geolocation: /api/v1/drivers, /ws/tracking
    location /api/v1/drivers {
        proxy_pass http://geolocation/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /ws/tracking {
        proxy_pass http://geolocation/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }

    # Notification: /api/v1/device-tokens, /api/v1/notifications, /api/v1/chat, /ws/chat
    location /api/v1/device-tokens { proxy_pass http://notification/; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /api/v1/notifications/ { proxy_pass http://notification/; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /api/v1/chat/ { proxy_pass http://notification/; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /ws/chat {
        proxy_pass http://notification/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }

    # Админ-панель (Next.js)
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

Для админки и мобильных приложений укажите базовый URL API: `https://indrive.1tlt.ru` — тогда запросы идут на `/auth/...`, `/api/v1/rides`, `/api/v1/users` и т.д. по тому же домену.

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

После этого в конфиге Nginx раскомментируйте редирект с HTTP на HTTPS (строка `return 301 https://...`) и снова выполните:

```bash
sudo nginx -t && sudo systemctl reload nginx
```

---

## 11. Запуск через systemd (опционально)

Чтобы сервисы поднимались после перезагрузки, можно завести unit-файлы.

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

Аналогично создаются `indrive-user`, `indrive-ride`, `indrive-geolocation`, `indrive-payment`, `indrive-notification`, `indrive-webadmin` (для webadmin `ExecStart` — вызов `node`/`pnpm start` из `apps/web-admin` с нужным `NEXT_PUBLIC_RIDE_API_URL`).

После создания unit-файлов:

```bash
sudo systemctl daemon-reload
sudo systemctl enable indrive-auth indrive-user indrive-ride indrive-geolocation indrive-payment indrive-notification indrive-webadmin
sudo systemctl start indrive-auth indrive-user indrive-ride indrive-geolocation indrive-payment indrive-notification indrive-webadmin
```

---

## 12. Порты сервисов (сводка)

| Сервис       | Порт  | Описание          |
|-------------|-------|-------------------|
| Auth        | 8080  | Регистрация, JWT  |
| User        | 8081  | Профили           |
| Geolocation | 8082  | Геолокация, WS    |
| Ride        | 8083  | Поездки, ставки   |
| Payment     | 8084  | Платежи           |
| Notification| 8085  | Push, чат          |
| Web Admin   | 3000  | Админ-панель      |

---

## 13. Проверка после установки

1. Инфра: `docker compose -f infra/docker-compose.yml ps`
2. Health Auth: `curl http://localhost:8080/health`
3. Health Ride: `curl http://localhost:8083/health`
4. В браузере: `https://indrive.1tlt.ru` — должна открыться админ-панель.
5. Регистрация тестового пользователя:  
   `POST https://indrive.1tlt.ru/auth/register` с телом  
   `{"email":"test@example.com","password":"password123","role":"passenger"}`.

При необходимости откройте порты в файрволе:

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

---

Документ актуален для тестового стенда. Для продакшена дополнительно настройте резервное копирование БД, ротацию логов и мониторинг.
