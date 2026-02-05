# Конфигурация портов indrive

## Сервер: 192.168.1.121

Порты изменены чтобы не конфликтовать с другим сервером (192.168.1.152).

## Порты сервисов

| Сервис       | Порт  | Переменная окружения |
|--------------|-------|----------------------|
| Auth         | 9080  | `PORT=9080`          |
| User         | 9081  | `PORT=9081`          |
| Geolocation  | 9082  | `PORT=9082`          |
| Ride         | 9083  | `PORT=9083`          |
| Payment      | 9084  | `PORT=9084`          |
| Notification | 9085  | `PORT=9085`          |
| Web Admin    | 9000  | порт Next.js         |

## Запуск сервисов на VPS

```bash
# Auth (порт 9080)
cd /ssd/www/indrive/services/auth
PORT=9080 ./auth

# User (порт 9081)  
cd /ssd/www/indrive/services/user
PORT=9081 ./user

# Geolocation (порт 9082)
cd /ssd/www/indrive/services/geolocation
PORT=9082 ./geolocation

# Ride (порт 9083)
cd /ssd/www/indrive/services/ride
PORT=9083 ./ride

# Payment (порт 9084)
cd /ssd/www/indrive/services/payment
PORT=9084 ./payment

# Notification (порт 9085)
cd /ssd/www/indrive/services/notification
PORT=9085 node dist/index.js

# Web Admin (порт 9000)
cd /ssd/www/indrive/apps/web-admin
PORT=9000 pnpm start
```

## Проброс портов на роутере (Keenetic)

Добавить правила для **192.168.1.121**:

| Внешний порт | Внутренний IP   | Внутренний порт | Протокол |
|--------------|-----------------|-----------------|----------|
| 9080         | 192.168.1.121   | 9080            | TCP      |
| 9081         | 192.168.1.121   | 9081            | TCP      |
| 9082         | 192.168.1.121   | 9082            | TCP      |
| 9083         | 192.168.1.121   | 9083            | TCP      |
| 9084         | 192.168.1.121   | 9084            | TCP      |
| 9085         | 192.168.1.121   | 9085            | TCP      |
| 9000         | 192.168.1.121   | 9000            | TCP      |

## Systemd unit файлы

Пример для Auth:

```ini
# /etc/systemd/system/indrive-auth.service
[Unit]
Description=indrive Auth Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/ssd/www/indrive
Environment=PORT=9080
Environment=JWT_SECRET=your-secret
Environment=PG_DSN=postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable
ExecStart=/ssd/www/indrive/bin/auth
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Аналогично для остальных сервисов с соответствующими портами.

## Доступ к сервисам

- **Auth API:** http://192.168.1.121:9080
- **User API:** http://192.168.1.121:9081
- **Geolocation API:** http://192.168.1.121:9082
- **Ride API:** http://192.168.1.121:9083
- **Payment API:** http://192.168.1.121:9084
- **Notification API:** http://192.168.1.121:9085
- **Web Admin:** http://192.168.1.121:9000

## Внешний доступ (после проброса портов)

Если внешний IP роутера например `YOUR_EXTERNAL_IP`:

- **Auth API:** http://YOUR_EXTERNAL_IP:9080
- и т.д.

## Примечание

80/443 порты будут настроены позже через второй сервер (192.168.1.152) как reverse proxy на 192.168.1.121.
