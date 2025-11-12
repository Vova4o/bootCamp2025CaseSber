FROM node:20-alpine

WORKDIR /app

# Копирование package files
COPY package*.json ./

# Установка зависимостей
RUN npm ci

# Копирование остальных файлов
COPY . .

EXPOSE 3000

# Запуск dev сервера
CMD ["npm", "run", "dev"]