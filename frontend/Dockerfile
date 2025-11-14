FROM node:20-alpine

WORKDIR /app

# Установка системных зависимостей
RUN apk add --no-cache libc6-compat

# Копирование файлов зависимостей
COPY package.json yarn.lock* ./

# Установка Yarn и зависимостей
RUN yarn install

# Копирование остального кода
COPY . .

# Переменные окружения
ENV NODE_ENV=development
ENV NEXT_TELEMETRY_DISABLED=1
ENV HOSTNAME=0.0.0.0

EXPOSE 3000

# Запуск
CMD ["yarn", "dev"]