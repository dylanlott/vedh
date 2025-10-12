# Multi-stage Dockerfile for building and serving the Vite + Vue 3 app

# --- Builder ---------------------------------------------------------------
FROM node:20-bullseye-slim AS builder
WORKDIR /app

# Install dependencies using package-lock for reproducible builds
COPY package.json package-lock.json ./
RUN npm ci

# Copy source and build
COPY . .
RUN npm run build


# --- Runner (nginx) ------------------------------------------------------
FROM nginx:stable-alpine AS runner

# Copy build output
COPY --from=builder /app/dist /usr/share/nginx/html

# Replace default nginx config with our SPA-aware config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
STOPSIGNAL SIGTERM

CMD ["nginx", "-g", "daemon off;"]
