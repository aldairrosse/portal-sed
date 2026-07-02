FROM node:22-alpine AS builder

ARG VITE_SHOW_DEV_TOOLBAR=false
ARG VITE_SHOW_DEMO_USERS=false
ENV VITE_SHOW_DEV_TOOLBAR=$VITE_SHOW_DEV_TOOLBAR
ENV VITE_SHOW_DEMO_USERS=$VITE_SHOW_DEMO_USERS

WORKDIR /app

COPY package.json ./
RUN npm install

COPY . .
RUN npm run build

FROM nginx:stable-alpine

COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
