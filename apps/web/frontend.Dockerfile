# Development stage with hot reload
FROM node:22-alpine AS development

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy source code
COPY . .

ARG REACT_APP_API_URL
ARG REACT_APP_FRONTEND_URL

ENV REACT_APP_API_URL=$REACT_APP_API_URL
ENV REACT_APP_FRONTEND_URL=$REACT_APP_FRONTEND_URL

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Start development server with hot reload
CMD ["npm", "start"]

# Build stage
FROM node:22-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy frontend source
COPY . .

ARG REACT_APP_API_URL
ARG REACT_APP_FRONTEND_URL

ENV REACT_APP_API_URL=$REACT_APP_API_URL
ENV REACT_APP_FRONTEND_URL=$REACT_APP_FRONTEND_URL

RUN npm run build

# Production stage
FROM nginx:alpine AS production

ARG REACT_APP_API_URL
ARG BACKEND_HOST
ENV REACT_APP_API_URL=$REACT_APP_API_URL
ENV BACKEND_HOST=$BACKEND_HOST

COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf.template

EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:80 || exit 1

CMD ["sh", "-c", "envsubst '${REACT_APP_API_URL} ${BACKEND_HOST}' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf && nginx -g 'daemon off;'"]