# ---------- Build stage ----------
FROM node:20-alpine AS builder

WORKDIR /app

# Install ALL deps (dev deps needed for CRA build)
COPY package*.json ./
RUN npm ci

COPY . .

ARG REACT_APP_API_URL
ARG REACT_APP_FRONTEND_URL

ENV REACT_APP_API_URL=$REACT_APP_API_URL
ENV REACT_APP_FRONTEND_URL=$REACT_APP_FRONTEND_URL

RUN npm run build

# ---------- Production stage ----------
FROM nginx:alpine

COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]