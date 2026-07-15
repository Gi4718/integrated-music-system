# 集成音乐系统

基于 Docker 的网易云音乐管理与下载系统，提供 Web 界面操作。支持扫码/手机/邮箱/QQ 多种登录方式、多音质下载、歌单同步、SSL 加密、ACME 自动证书续期等功能。

UI 风格致敬《明日方舟：终末地》。

## 功能特性

### 账号与认证
- 多种网易云登录方式：扫码、手机号验证码、手机号密码、邮箱、QQ
- 系统级用户注册/登录（JWT 认证），保护所有功能接口
- 二次验证支持

### 音乐搜索与下载
- 关键词搜索歌曲
- 多音质下载：标准 128k / 高品质 320k / 无损 FLAC
- 单曲下载与歌单批量下载
- 下载进度实时跟踪
- 断点续传支持
- 下载前自动扫描目录，跳过已下载歌曲
- 自适应 API 限流，避免请求被拒绝

### 歌单管理
- 查看用户歌单列表
- 歌单详情浏览与收藏
- 歌单同步到本地（间隔模式 / 定时模式）
- 自动数据补全：专辑封面、歌词、艺人信息
- 同步完成后自动触发数据补全

### Web 播放器
- 在线流式播放
- 播放模式切换：顺序 / 随机 / 单曲循环
- 播放列表管理
- 添加到收藏歌单

### 下载任务管理
- 实时任务日志面板
- 任务终止按钮
- 自动恢复未完成任务（服务重启后）
- 已下载歌曲自动跳过并计入完成数

### SSL 与 HTTPS
- 三种 SSL 配置方式：
  - 本地证书（路径映射或上传）
  - DNS-ACME 自动申请（支持 Cloudflare、阿里云、腾讯云等主流 DNS 服务商）
- 证书验证详情展示（颁发者、有效期、域名）
- 证书临期自动续期（30 天阈值，每日检查，续期后热加载）
- HTTP 到 HTTPS 自动重定向
- 中间证书链自动合并

### 界面与体验
- 日间 / 夜间主题自动跟随系统
- 手动主题切换
- 移动端响应式适配（<768px）
- 页面过渡动画（可关闭）

## 快速开始

### 使用 Docker Compose

```bash
# GitHub
git clone https://github.com/Gi4718/integrated-music-system.git
# 或 Gitee（国内加速）
git clone https://gitee.com/luofeng0108/integrated-music-system.git

cd integrated-music-system
docker compose up -d --build
```

访问 `http://localhost:33550` 即可使用。

### 端口说明

- `33550`: HTTP 端口
- `33551`: HTTPS 端口（需开启 SSL）

### 数据目录

| 挂载路径 | 说明 |
|---------|------|
| `./data/config` | 配置文件 |
| `./data/ssl` | SSL 证书 |
| `./data/downloads` | 下载文件 |
| `./data/db` | 数据库 |
| `/vol1/1000/music` → `/music` | 音乐文件存储 |

### 首次使用

1. 打开 Web 界面，注册系统账号
2. 使用网易云账号登录（推荐扫码方式）
3. 在设置页面配置下载路径和同步规则
4. 开始搜索和下载音乐

## 页面说明

| 页面 | 路径 | 说明 |
|------|------|------|
| 首页 | `/` | 推荐歌曲与歌单 |
| 登录 | `/login` | 系统账号登录 |
| 注册 | `/register` | 系统账号注册 |
| 搜索 | `/search` | 歌曲搜索与下载 |
| 歌单 | `/playlist` | 歌单管理与同步 |
| 下载 | `/downloads` | 下载历史与任务管理 |
| 设置 | `/settings` | 系统配置与 SSL 管理 |

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

### 构建 Docker 镜像

```bash
docker compose up -d --build
```

## 技术栈

- **后端**: Go 1.21 + Gin
- **前端**: Vue 3 + TypeScript + Pinia
- **UI 组件**: Element Plus
- **数据库**: SQLite
- **网易云 API**: NeteaseCloudMusicApi (Node.js)
- **ACME**: lego v4 (支持多 DNS 服务商)
- **容器化**: Docker + Docker Compose + Supervisor

## 许可证

MIT
