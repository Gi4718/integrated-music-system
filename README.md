# 集成音乐系统

基于 Docker 的网易云音乐收集与下载系统，提供 Web 界面和命令行两种操作方式。

UI 风格致敬《明日方舟：终末地》。

## 功能特性

- 扫码登录（模拟电脑端获取长期 Cookie）
- 多音质下载（标准 128k / 高品质 320k / 无损 FLAC）
- 歌单批量下载（账户歌单 / 分享歌单）
- Web 端音乐播放
- SSL 支持（上传证书 / 路径映射 / DNS-ACME 自动申请）
- 自动元数据补全（封面、歌词、艺术家信息）
- 下载进度跟踪
- 文件名安全处理
- 日间 / 夜间主题自动切换

## 快速开始

### 使用 Docker Compose

```bash
git clone <repo-url>
cd netmusic-downloader

# 初始化 NeteaseCloudMusicApi（国内用户请使用 Gitee 镜像）
cd netease-api
git clone https://gitee.com/mirrors/NeteaseCloudMusicApi.git .
npm install --registry=https://registry.npmmirror.com
cd ..

# 启动服务
docker-compose up -d
```

访问 `http://localhost:33550` 即可使用。

### 端口说明

- `33550`: HTTP 端口
- `33551`: HTTPS 端口（需开启 SSL）

### 数据目录

- `./data/config`: 配置文件
- `./data/ssl`: SSL 证书
- `./data/downloads`: 下载文件
- `./data/db`: 数据库

## 命令行使用

```bash
# 搜索歌曲
docker exec -it endfield-music /app/endfield-music-cli search "周杰伦 晴天"

# 下载歌曲
docker exec -it endfield-music /app/endfield-music-cli download --id 123456 --quality flac

# 下载歌单
docker exec -it endfield-music /app/endfield-music-cli playlist --id 789 --quality 320

# 查看登录状态
docker exec -it endfield-music /app/endfield-music-cli login --status
```

## 开发

### 后端

```bash
go mod download
go run cmd/server/main.go --port 33550
```

### 前端

```bash
cd web
npm install
npm run dev    # 开发模式
npm run build  # 构建
```

## 技术栈

- **后端**: Go 1.21 + Gin
- **前端**: Vue 3 + Element Plus + TypeScript
- **数据库**: SQLite
- **网易云 API**: NeteaseCloudMusicApi (Node.js)
- **容器化**: Docker + Docker Compose

## 许可证

MIT
