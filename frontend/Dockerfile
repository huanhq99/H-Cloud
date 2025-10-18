FROM node:18-alpine AS builder

WORKDIR /app

# 复制package.json和package-lock.json
COPY package*.json ./
RUN npm install

# 复制源代码
COPY . .

# 构建前端应用
RUN npm run build

# 生产环境镜像
FROM nginx:alpine

# 复制构建产物到Nginx服务目录
COPY --from=builder /app/build /usr/share/nginx/html

# 复制Nginx配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]