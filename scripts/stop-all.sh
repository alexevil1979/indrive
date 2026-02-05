#!/bin/bash
# ============================================
# indrive - Скрипт остановки всех сервисов
# ============================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_DIR="/ssd/www/indrive"

echo -e "${YELLOW}Остановка всех сервисов indrive...${NC}"

# Остановить backend
pkill -f "./bin/auth" 2>/dev/null || true
pkill -f "./bin/user" 2>/dev/null || true
pkill -f "./bin/geolocation" 2>/dev/null || true
pkill -f "./bin/ride" 2>/dev/null || true
pkill -f "./bin/payment" 2>/dev/null || true
pkill -f "node dist/index.js" 2>/dev/null || true
pkill -f "next start" 2>/dev/null || true

echo -e "${GREEN}Backend сервисы остановлены${NC}"

# Опционально: остановить Docker
read -p "Остановить Docker контейнеры? (y/N): " answer
if [ "$answer" = "y" ] || [ "$answer" = "Y" ]; then
    cd $PROJECT_DIR
    docker compose -f infra/docker-compose.yml down
    echo -e "${GREEN}Docker контейнеры остановлены${NC}"
fi

echo -e "${GREEN}Готово!${NC}"
