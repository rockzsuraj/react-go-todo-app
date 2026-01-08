FROM node:20-alpine

WORKDIR /app

# Install deps first (cache-friendly)
COPY apps/web/package.json apps/web/package-lock.json* ./
RUN npm install

# Copy frontend source
COPY apps/web .

RUN npm run build

EXPOSE 3000
CMD ["npm", "start"]