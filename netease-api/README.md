# NeteaseCloudMusicApi

此目录用于存放网易云音乐 API Node.js 服务。

## 安装

```bash
# 克隆 NeteaseCloudMusicApi 项目（国内用户请使用 Gitee 镜像）
git clone https://gitee.com/mirrors/NeteaseCloudMusicApi.git .

# 安装依赖（国内换源）
npm install --registry=https://registry.npmmirror.com
```

## 运行

```bash
node app.js
```

服务将在 `http://localhost:3000` 启动。

## 说明

此服务由 Docker 容器内的 supervisord 自动管理，无需手动启动。
