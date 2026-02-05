#!/bin/bash
# ============================================
# indrive - Проверка статуса сервисов
# ============================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_DIR="/ssd/www/indrive"

# Порты
PORT_AUTH=9080
PORT_USER=9081
PORT_GEOLOCATION=9082
PORT_RIDE=9083
PORT_PAYMENT=9084
PORT_NOTIFICATION=9085
PORT_WEBADMIN=9010

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  indrive - Статус сервисов${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Docker контейнеры
echo -e "${YELLOW}Docker контейнеры:${NC}"
cd $PROJECT_DIR
docker compose -f infra/docker-compose.yml ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "Docker Compose не доступен"
echo ""

# Backend сервисы
echo -e "${YELLOW}Backend сервисы:${NC}"

check_service() {
    local name=$1
    local port=$2
    if curl -s --connect-timeout 2 http://localhost:$port/health > /dev/null 2>&1; then
        echo -e "  $name (порт $port): ${GREEN}OK${NC}"
    else
        # Проверим процесс
        if pgrep -f "$name" > /dev/null 2>&1; then
            echo -e "  $name (порт $port): ${YELLOW}ЗАПУЩЕН (health не отвечает)${NC}"
        else
            echo -e "  $name (порт $port): ${RED}НЕ ЗАПУЩЕН${NC}"
        fi
    fi
}

check_service "auth" $PORT_AUTH
check_service "user" $PORT_USER
check_service "geolocation" $PORT_GEOLOCATION
check_service "ride" $PORT_RIDE
check_service "payment" $PORT_PAYMENT

# Notification (проверяем по процессу)
if pgrep -f "node dist/index.js" > /dev/null 2>&1; then
    echo -e "  notification (порт $PORT_NOTIFICATION): ${GREEN}OK${NC}"
else
    echo -e "  notification (порт $PORT_NOTIFICATION): ${RED}НЕ ЗАПУЩЕН${NC}"
fi

# Web Admin
if pgrep -f "next start" > /dev/null 2>&1; then
    echo -e "  web-admin (порт $PORT_WEBADMIN): ${GREEN}OK${NC}"
else
    echo -e "  web-admin (порт $PORT_WEBADMIN): ${RED}НЕ ЗАПУЩЕН${NC}"
fi

echo ""
echo -e "${YELLOW}Порты:${NC}"
ss -tlnp | grep -E "(9080|9081|9082|9083|9084|9085|9010)" 2>/dev/null || netstat -tlnp | grep -E "(9080|9081|9082|9083|9084|9085|9010)" 2>/dev/null || echo "Не удалось получить информацию о портах"
echo ""
