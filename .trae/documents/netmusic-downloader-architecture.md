# 网易云音乐下载器 - 架构规划

## 背景

构建一个 x86_64 Linux Docker 可用的网易云音乐下载器，提供 Web 界面和命令行两种操作方式。核心需求：扫码登录获取长期 Cookie、多音质下载、歌单批量下载、SSL 三种模式（上传/路径/DNS-ACME）、Web 端音乐播放。

---

## 技术栈

| 组件 | 选型 | 理由 |
|------|------|------|
| 后端 | Go 1.21 + Gin | 单二进制部署，性能好 |
| 前端 | Vue 3 + Element Plus | 生态成熟，中文文档丰富 |
| 网易云 API | NeteaseCloudMusicApi (Node.js) | 开源社区维护，接口全 |
| 数据库 | SQLite | 轻量，无需额外服务 |
| CLI | Cobra | Go 标准 CLI 框架 |
| ACME | go-acme/lego | 支持多种 DNS Provider |
| 音频标签 | go-audio/taglib 或 dhowden/tag | 元数据嵌入 |

---

## 项目结构

```
netmusic-downloader/
├── cmd/
│   ├── server/main.go          # Web 服务入口
│   └── cli/main.go             # CLI 入口
├── internal/
│   ├── api/                    # HTTP 路由与处理器
│   │   ├── router.go
│   │   ├── auth.go             # 扫码登录、Cookie 管理
│   │   ├── search.go           # 搜索歌曲
│   │   ├── download.go         # 下载任务
│   │   ├── playlist.go         # 歌单（账户/分享）
│   │   ├── player.go           # 在线播放流
│   │   └── settings.go         # 配置管理
│   ├── service/
│   │   ├── netease.go          # 网易云 API 封装（调用 Node.js API）
│   │   ├── downloader.go       # 下载核心（限速、重试、优先级）
│   │   ├── metadata.go         # 元数据补全（后台任务）
│   │   └── queue.go            # 下载队列管理
│   ├── model/                  # 数据模型
│   ├── config/                 # 配置加载
│   ├── ssl/
│   │   ├── manager.go          # SSL 模式管理
│   │   ├── upload.go           # 上传证书
│   │   ├── path.go             # 路径映射
│   │   └── acme.go             # DNS-ACME 自动申请
│   ├── acme/                   # DNS Provider 适配
│   │   ├── provider.go         # 统一接口
│   │   ├── aliyun.go
│   │   ├── tencent.go
│   │   ├── cloudflare.go
│   │   ├── huawei.go
│   │   ├── volcengine.go       # 火山引擎
│   │   ├── baidu.go            # 百度智能云
│   │   ├── westcn.go           # 西部数码
│   │   ├── sanwu.go            # 三五互联
│   │   ├── dns.com             # 帝恩思
│   │   ├── shidai.go           # 时代互联
│   │   ├── jdcloud.go          # 京东云
│   │   └── custom.go           # 自定义接口
│   └── db/                     # SQLite 操作
├── web/                        # Vue 前端源码
│   ├── src/
│   │   ├── views/
│   │   │   ├── Search.vue      # 搜歌页
│   │   │   ├── Playlist.vue    # 歌单页
│   │   │   ├── Downloads.vue   # 下载历史
│   │   │   ├── Player.vue      # 播放器
│   │   │   ├── Login.vue       # 扫码登录
│   │   │   └── Settings.vue    # 设置页
│   │   └── components/
│   └── package.json
├── netease-api/                # NeteaseCloudMusicApi 子模块
├── Dockerfile                  # 多阶段构建
├── docker-compose.yml
└── README.md
```

---

## 核心模块设计

### 1. Docker 容器架构

单容器方案，内部运行两个进程（由 supervisord 管理）：
- Go 主服务（Web + CLI 共用）
- Node.js 网易云 API 服务（本地 127.0.0.1:3000）

**端口映射**:
- 33550: HTTP
- 33551: HTTPS（开启 SSL 时）

**卷映射**:
- `/data/config`: 配置文件
- `/data/ssl`: SSL 证书存储
- `/data/downloads`: 下载目录（用户自定义映射路径）
- `/data/db`: SQLite 数据库

### 2. SSL 三种模式

**模式 A - 上传证书**:
- Web 界面上传 `.crt` + `.key` 文件
- 存储到 `/data/ssl/uploaded/`

**模式 B - 路径映射**:
- 用户指定容器内绝对路径（如 `/data/ssl/acme/fullchain.pem`）
- 适用于用户自行通过 acme.sh 等工具管理的证书

**模式 C - DNS-ACME 自动申请**:
- 内置 lego ACME 客户端
- 支持 12 家 DNS 服务商 + 自定义接口
- 统一配置格式：账号（必填）、密钥（必填）、区域 ID（可选）
- 网页提供各厂商配置指南

**DNS Provider 适配清单**:
| 厂商 | lego 内置支持 | 备注 |
|------|--------------|------|
| 阿里云 | ✅ (alidns) | 直接适配 |
| 腾讯云/DNSPod | ✅ (tencentcloud) | 直接适配 |
| Cloudflare | ✅ (cloudflare) | 直接适配 |
| 华为云 | ✅ (huaweicloud) | 直接适配 |
| 火山引擎 | ❌ | 需自定义实现 |
| 百度智能云 | ❌ | 需自定义实现 |
| 西部数码 | ❌ | 需自定义实现 |
| 三五互联 | ❌ | 需自定义实现 |
| 帝恩思 | ❌ | 需自定义实现 |
| 时代互联 | ❌ | 需自定义实现 |
| 京东云 | ❌ | 需自定义实现 |
| 自定义接口 | - | 预留 webhook 接口 |

### 3. 扫码登录流程

1. Web 端请求 `/api/auth/qr-key` 获取登录 key
2. 生成二维码图片展示在页面
3. 轮询 `/api/auth/qr-check` 检查登录状态
4. 登录成功后获取 Cookie 存储到 SQLite
5. CLI 端读取同一 Cookie 文件，实现共享

**模拟电脑端**: 请求头携带 PC 端 User-Agent，获取长期有效 Cookie（约 30 天）

### 4. 下载核心逻辑

**优先级策略**:
1. **P0**: 歌曲音频文件下载（保证成功）
2. **P1**: 后台补全元数据（作曲者、专辑、封面、歌词）

**限速控制**:
- 音频下载：间隔 1-2 秒
- 元数据请求：间隔 3-5 秒（网易云限制严格）
- 失败重试：指数退避，最多 3 次

**音质选择**:
- 标准 (128kbps MP3)
- 高品质 (320kbps MP3)
- 无损 (FLAC)
- 用户可在设置页配置默认音质

**文件名安全**:
- 过滤非法字符：`/\:*?"<>|`
- 长度截断（文件名最大 200 字符）
- Unicode 规范化
- 重名处理：自动追加序号

### 5. 下载类型

**A. 搜歌下载**:
- 输入歌名关键词
- 调用网易云搜索 API
- 按相似度排序展示
- 选择音质后下载单曲

**B. 歌单下载**:
- **账户歌单**: 读取登录用户收藏/创建的歌单
- **分享歌单**: 输入歌单 ID 或分享链接
- 批量下载整个歌单，自动创建子文件夹

### 6. Web 播放器

- 使用 Howler.js 实现音频播放
- 后端提供 `/api/player/stream/:id` 流式接口
- 支持播放/暂停、进度条、音量控制
- 播放时可查看歌曲信息

### 7. Web 界面页面

| 页面 | 功能 |
|------|------|
| 登录页 | 扫码登录、Cookie 状态显示 |
| 搜索页 | 关键词搜索、结果列表、在线播放、下载 |
| 歌单页 | 账户歌单列表、分享歌单导入、批量下载 |
| 下载页 | 下载进度、历史记录、失败重试 |
| 设置页 | 下载路径、音质、SSL 配置、DNS-ACME 配置 |

### 8. CLI 命令设计

```bash
# 启动 Web 服务
netmusic server --port 33550

# 搜索歌曲
netmusic search "周杰伦 晴天"

# 下载歌曲
netmusic download --id 123456 --quality flac

# 下载歌单
netmusic playlist --id 789 --quality 320

# 登录状态
netmusic login --status

# 扫码登录（终端显示二维码）
netmusic login --qr
```

---

## 数据流

```
用户操作 (Web/CLI)
    ↓
Go API Server (Gin)
    ↓
NeteaseCloudMusicApi (Node.js, 本地 3000 端口)
    ↓
网易云服务器
    ↓
返回音频 URL → Go 下载器 → 写入磁盘
    ↓
后台元数据任务 → 嵌入标签 → 完成
```

---

## Docker 构建

**多阶段构建**:
1. **Stage 1**: 编译 Go 二进制
2. **Stage 2**: 构建 Vue 前端
3. **Stage 3**: 运行镜像（Alpine + Go binary + Node.js + supervisord）

**镜像大小预估**: ~200MB（含 Node.js 运行时）

---

## 开发阶段

### Phase 1: 基础框架 (3-4 天)
- [ ] 项目初始化、目录结构
- [ ] Docker 多阶段构建
- [ ] Go + Vue 基础骨架
- [ ] SQLite 数据库初始化
- [ ] NeteaseCloudMusicApi 集成

### Phase 2: 登录与认证 (2 天)
- [ ] 扫码登录 API
- [ ] Cookie 存储与共享
- [ ] Web 登录页面

### Phase 3: 搜索与下载 (4-5 天)
- [ ] 搜索 API 封装
- [ ] 下载器核心（限速、重试）
- [ ] 元数据后台补全
- [ ] 文件名安全处理
- [ ] 音质选择

### Phase 4: 歌单功能 (2 天)
- [ ] 账户歌单读取
- [ ] 分享歌单解析
- [ ] 批量下载队列

### Phase 5: Web 播放器 (2 天)
- [ ] 流式播放接口
- [ ] Howler.js 播放器组件
- [ ] 播放控制 UI

### Phase 6: SSL 管理 (3 天)
- [ ] 上传证书模式
- [ ] 路径映射模式
- [ ] DNS-ACME 自动申请
- [ ] 12 家 DNS Provider 适配
- [ ] 配置指南页面

### Phase 7: CLI 工具 (2 天)
- [ ] Cobra 命令框架
- [ ] 搜索、下载、歌单命令
- [ ] 终端二维码显示

### Phase 8: 测试与优化 (2-3 天)
- [ ] 端到端测试
- [ ] 性能优化
- [ ] 文档编写

**总预估**: 20-23 天

---

## 关键依赖

```
Go:
- github.com/gin-gonic/gin
- github.com/spf13/cobra
- github.com/go-sqlite3/sqlite3
- github.com/go-acme/lego/v4
- github.com/dhowden/tag
- github.com/skip2/go-qrcode

Vue:
- vue@3
- element-plus
- pinia
- axios
- howler
```

---

## 待确认事项

1. **网易云 API 部署方式**: 是否直接使用官方 NeteaseCloudMusicApi 项目，还是自行实现？
   - 建议：直接使用，稳定性好，社区维护

2. **元数据嵌入格式**: 是否支持所有音质格式？
   - MP3: ID3 标签
   - FLAC: Vorbis Comment

3. **下载并发数**: 默认并发下载数量？
   - 建议：默认 2，可配置

4. **日志级别**: 是否需要详细的 API 请求日志？
   - 建议：可配置，默认 INFO

---

## 下一步

等待用户确认架构方案后，开始 Phase 1 实现。
