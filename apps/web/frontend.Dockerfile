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

# envsubst is included in gettext; needed to template nginx.conf at startup
RUN apk add --no-cache gettext

COPY --from=builder /app/build /usr/share/nginx/html
# Store the template — entrypoint will substitute ${API_PROXY_PASS} at runtime
COPY nginx.conf /etc/nginx/nginx.conf.template

EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:80 || exit 1

# Substitute env vars into nginx config, then start nginx
CMD ["/bin/sh", "-c", "if [ -n \"${API_BACKEND_HOSTPORT}\" ]; then export API_PROXY_PASS=\"http://${API_BACKEND_HOSTPORT}\"; elif [ -n \"${API_BACKEND_URL}\" ]; then export API_PROXY_PASS=\"${API_BACKEND_URL%/}\"; else echo \"API_BACKEND_HOSTPORT and API_BACKEND_URL are not set; falling back to https://todos-api.onrender.com\"; export API_PROXY_PASS=\"https://todos-api.onrender.com\"; fi; envsubst '${API_PROXY_PASS}' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf && nginx -g 'daemon off;'"]
