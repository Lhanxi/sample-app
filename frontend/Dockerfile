FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN npm run build

FROM nginxinc/nginx-unprivileged:1.30.4-alpine3.24

COPY --chown=101:101 nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder --chown=101:101 /app/dist /usr/share/nginx/html

USER 101

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
