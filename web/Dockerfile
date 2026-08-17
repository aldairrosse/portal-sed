FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json ./
RUN npm install

COPY . .
RUN npm run build

FROM nginx:stable-alpine

COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

# ponytail: pid en /dev/shm — kernel 3.10/overlayfs no permite pwrite en /run
RUN sed -i 's|/run/nginx.pid|/dev/shm/nginx.pid|' /etc/nginx/nginx.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
