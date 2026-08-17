FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json ./
RUN npm install

COPY . .
RUN npm run build

FROM nginx:stable-alpine

COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
# ponytail: pid en /tmp — /dev/shm pwrite falla en CentOS 7 (kernel 3.10 + SELinux)
CMD ["nginx", "-g", "daemon off; pid /tmp/nginx.pid;"]
