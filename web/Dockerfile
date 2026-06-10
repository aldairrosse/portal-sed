FROM node:22-alpine AS builder

ARG VITE_SHOW_DEV_TOOLBAR=false
ENV VITE_SHOW_DEV_TOOLBAR=$VITE_SHOW_DEV_TOOLBAR

WORKDIR /app
RUN corepack enable && corepack prepare pnpm@latest --activate

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile --store-dir /tmp/.pnpm-store

COPY . .
RUN pnpm run build

FROM nginx:stable-alpine

COPY --from=builder /app/.svelte-kit/output/client /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
