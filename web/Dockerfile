FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json ./
RUN npm install

COPY . .
RUN npm run build

FROM nginx:stable-alpine

COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

# ponytail: nginx:stable-alpine 1.30 falla con `pwrite EPERM` sobre el pid file
# en CentOS 7 (kernel 3.10 + overlay2 + SELinux spc_t). Desactivamos la
# escritura del pid: nginx funciona igual sin él, reinicios via `docker compose`.
RUN sed -i 's|^pid .*|pid off;|' /etc/nginx/nginx.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
