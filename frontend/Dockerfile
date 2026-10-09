# ---- build stage ----
FROM node:22-alpine AS build
RUN corepack enable
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

# ---- run stage ----
FROM nginx:alpine
# Image nginx menjalankan envsubst otomatis pada /etc/nginx/templates/*.template.
# Default upstream = backend di host (pemakaian standalone).
# Root compose monorepo meng-override jadi http://api:8080 (service name, 1 network).
ENV API_UPSTREAM=http://host.docker.internal:8080
COPY nginx/default.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
  CMD wget -qO- http://localhost/ > /dev/null || exit 1
