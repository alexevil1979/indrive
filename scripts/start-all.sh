#!/bin/bash
# ============================================
# indrive - Скрипт запуска всех сервисов
# Сервер: 192.168.1.121
# ============================================

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PROJECT_DIR="/ssd/www/indrive"
LOG_DIR="/var/log/indrive"

# Порты сервисов
PORT_AUTH=9080
PORT_USER=9081
PORT_GEOLOCATION=9082
PORT_RIDE=9083
PORT_PAYMENT=9084
PORT_NOTIFICATION=9085
PORT_WEBADMIN=9010

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  indrive - Запуск всех сервисов${NC}"
echo -e "${GREEN}========================================${NC}"

# Создать директорию для логов
mkdir -p $LOG_DIR

# ============================================
# 1. Остановить старые процессы
# ============================================
echo -e "${YELLOW}[1/5] Остановка старых процессов...${NC}"
pkill -f "./bin/auth" 2>/dev/null || true
pkill -f "./bin/user" 2>/dev/null || true
pkill -f "./bin/geolocation" 2>/dev/null || true
pkill -f "./bin/ride" 2>/dev/null || true
pkill -f "./bin/payment" 2>/dev/null || true
pkill -f "node dist/index.js" 2>/dev/null || true
pkill -f "next start" 2>/dev/null || true
sleep 2
echo -e "${GREEN}   Готово${NC}"

# ============================================
# 2. Запуск Docker инфраструктуры
# ============================================
echo -e "${YELLOW}[2/5] Запуск Docker инфраструктуры...${NC}"
cd $PROJECT_DIR
docker compose -f infra/docker-compose.yml up -d
echo -e "${GREEN}   Контейнеры запущены${NC}"

# ============================================
# 3. Ожидание готовности PostgreSQL
# ============================================
echo -e "${YELLOW}[3/5] Ожидание готовности PostgreSQL...${NC}"
MAX_ATTEMPTS=30
ATTEMPT=0
until docker compose -f infra/docker-compose.yml exec -T postgres pg_isready -U ridehail > /dev/null 2>&1; do
    ATTEMPT=$((ATTEMPT + 1))
    if [ $ATTEMPT -ge $MAX_ATTEMPTS ]; then
        echo -e "${RED}   PostgreSQL не готов после $MAX_ATTEMPTS попыток${NC}"
        exit 1
    fi
    echo -n "."
    sleep 1
done
echo -e "\n${GREEN}   PostgreSQL готов${NC}"

# Ожидание Redis
echo -e "${YELLOW}   Проверка Redis...${NC}"
ATTEMPT=0
until docker compose -f infra/docker-compose.yml exec -T redis redis-cli ping > /dev/null 2>&1; do
    ATTEMPT=$((ATTEMPT + 1))
    if [ $ATTEMPT -ge $MAX_ATTEMPTS ]; then
        echo -e "${RED}   Redis не готов${NC}"
        exit 1
    fi
    sleep 1
done
echo -e "${GREEN}   Redis готов${NC}"

# ============================================
# 4. Запуск backend сервисов
# ============================================
echo -e "${YELLOW}[4/5] Запуск backend сервисов...${NC}"
cd $PROJECT_DIR

# Auth
echo -e "   Запуск Auth (порт $PORT_AUTH)..."
PORT=$PORT_AUTH nohup ./bin/auth > $LOG_DIR/auth.log 2>&1 &
sleep 1

# User
echo -e "   Запуск User (порт $PORT_USER)..."
PORT=$PORT_USER nohup ./bin/user > $LOG_DIR/user.log 2>&1 &
sleep 1

# Geolocation
echo -e "   Запуск Geolocation (порт $PORT_GEOLOCATION)..."
PORT=$PORT_GEOLOCATION nohup ./bin/geolocation > $LOG_DIR/geolocation.log 2>&1 &
sleep 1

# Ride
echo -e "   Запуск Ride (порт $PORT_RIDE)..."
PORT=$PORT_RIDE nohup ./bin/ride > $LOG_DIR/ride.log 2>&1 &
sleep 1

# Payment
echo -e "   Запуск Payment (порт $PORT_PAYMENT)..."
PORT=$PORT_PAYMENT nohup ./bin/payment > $LOG_DIR/payment.log 2>&1 &
sleep 1

# Notification
echo -e "   Запуск Notification (порт $PORT_NOTIFICATION)..."
cd $PROJECT_DIR/services/notification
PORT=$PORT_NOTIFICATION nohup node dist/index.js > $LOG_DIR/notification.log 2>&1 &
sleep 1

echo -e "${GREEN}   Backend сервисы запущены${NC}"

# ============================================
# 5. Запуск Web Admin
# ============================================
echo -e "${YELLOW}[5/5] Запуск Web Admin (порт $PORT_WEBADMIN)...${NC}"
cd $PROJECT_DIR/apps/web-admin
PORT=$PORT_WEBADMIN nohup pnpm start > $LOG_DIR/webadmin.log 2>&1 &
sleep 2
echo -e "${GREEN}   Web Admin запущен${NC}"

# ============================================
# Проверка статуса
# ============================================
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Все сервисы запущены!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Порты сервисов:"
echo -e "  Auth:         http://192.168.1.121:$PORT_AUTH"
echo -e "  User:         http://192.168.1.121:$PORT_USER"
echo -e "  Geolocation:  http://192.168.1.121:$PORT_GEOLOCATION"
echo -e "  Ride:         http://192.168.1.121:$PORT_RIDE"
echo -e "  Payment:      http://192.168.1.121:$PORT_PAYMENT"
echo -e "  Notification: http://192.168.1.121:$PORT_NOTIFICATION"
echo -e "  Web Admin:    http://192.168.1.121:$PORT_WEBADMIN"
echo ""
echo -e "Логи: $LOG_DIR/"
echo ""
echo -e "Проверка health:"
sleep 3
curl -s http://localhost:$PORT_AUTH/health && echo " - Auth OK" || echo " - Auth FAIL"
echo ""
echo -e "${YELLOW}Для просмотра логов: tail -f $LOG_DIR/*.log${NC}"
