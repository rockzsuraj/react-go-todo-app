# Multi-stage build for production
FROM node:20-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev --ignore-scripts

COPY . .
ARG REACT_APP_API_URL
ARG REACT_APP_SUPABASE_URL  
ARG REACT_APP_SUPABASE_ANON_KEY
RUN npm run build

# Production stage
FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]