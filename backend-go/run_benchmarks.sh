#!/bin/bash

set -e

echo "🧪 Starting Research Pro Mode Benchmarks"
echo "========================================"

# Цвета для вывода
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Создаем директорию для результатов
mkdir -p benchmark_results

# Запускаем сервисы
echo -e "${YELLOW}📦 Starting services...${NC}"
docker-compose up -d postgres redis backend

# Ждем готовности backend
echo -e "${YELLOW}⏳ Waiting for backend to be ready...${NC}"
sleep 10

until curl -f http://localhost:8000/api/health > /dev/null 2>&1; do
  echo "Waiting for backend..."
  sleep 3
done

echo -e "${GREEN}✅ Backend is ready${NC}"

# Функция для запуска бенчмарка
run_benchmark() {
  local mode=$1
  local type=$2
  local limit=$3
  local output=$4
  
  echo -e "\n${YELLOW}🔬 Running $type benchmark in $mode mode (limit: $limit)...${NC}"
  
  if [ "$type" == "simple" ]; then
    docker-compose run --rm benchmark ./simple_bench \
      -mode="$mode" \
      -limit=$limit \
      -output="/benchmark/results/$output" \
      -api="http://backend:8000"
  else
    docker-compose run --rm benchmark ./frames_bench \
      -mode="$mode" \
      -limit=$limit \
      -output="/benchmark/results/$output" \
      -api="http://backend:8000"
  fi
  
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Completed: $output${NC}"
  else
    echo -e "${RED}❌ Failed: $output${NC}"
  fi
}

# SimpleQA Benchmarks
echo -e "\n${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  SimpleQA Benchmarks${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_benchmark "simple" "simple" 30 "simple_mode.json"
run_benchmark "pro" "simple" 30 "pro_mode.json"
run_benchmark "auto" "simple" 30 "auto_mode.json"

# FRAMES Benchmarks
echo -e "\n${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  FRAMES Multi-hop Benchmarks${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_benchmark "pro" "frames" 15 "frames_pro.json"

# Анализ результатов
echo -e "\n${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  Analyzing Results${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

python3 analyze_results.py

echo -e "\n${GREEN}🎉 All benchmarks completed!${NC}"
echo -e "${GREEN}📊 Results saved in ./benchmark_results/${NC}"
echo -e "${GREEN}📈 Check summary.md for presentation data${NC}"