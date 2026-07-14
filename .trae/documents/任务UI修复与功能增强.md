# 任务UI修复与功能增强

## Context

用户反馈多个问题：
1. 下载时后台流量在跑，但任务日志UI显示"等待中"，无进度更新，也没有数据补全任务
2. 自动数据补全应跟随自动同步运行，而非独立定时
3. SSL未上传中间证书时拒绝连接，重启后恢复
4. 系统登录后Login页错误显示网易云未登录UI，刷新后正常
5. 容器重启后应自动恢复未完成任务（断点续传）
6. 数据补全遇到API频率限制应重试，支持断点续传
7. 下载前扫盘跳过已下载文件

## 修改文件清单

| 文件 | 改动 |
|------|------|
| `internal/download/engine.go` | 任务状态更新、数据补全任务预创建、扫盘跳过、断点续传恢复、频率限制重试 |
| `cmd/server/main.go` | SSL证书加载重试机制 |
| `internal/api/router.go` | HTTPS就绪标记，重定向中间件检查 |
| `web/src/views/Login.vue` | onMounted中token同步修复 |
| `web/src/views/Settings.vue` | 移除数据补全间隔配置UI |

---

## 1. 任务UI不更新（P0）

**根因**: `executeTask` 只更新了 `DownloadTask.Status = "downloading"`，但未调用 `taskService.SetTaskStatus` 更新 TaskService 中的任务状态。TaskService 状态直到所有下载完成才通过 `UpdateTaskProgress` 更新为 running。

**修复** (`engine.go`):
- 在 `executeTask` 开头，`task.Status = "downloading"` 之后，添加 `e.taskService.SetTaskStatus(phase.DownloadTaskID, service.TaskStatusRunning)`
- 在 `downloadFile` 的 512KB 持久化处，`UpdateTaskProgress` 调用前确保状态为 running（`UpdateTaskProgress` 内部已自动设置，无需额外改动）

## 2. 数据补全任务不可见（P1）

**根因**: metadata task 在 `checkPlaylistPhaseComplete` 中（所有下载完成后）才创建，前端看不到等待信息。

**修复** (`engine.go`):
- 在 `AddPlaylistTask` 中创建 downloadTask 后，立即预创建 metadataTask（状态 pending），赋值 `phase.MetadataTaskID`
- 在 `checkPlaylistPhaseComplete` 下载阶段完成时，不再创建新 task，而是将预创建的 metadataTask 状态设为 running，更新 Total 为实际下载成功数
- metadata 阶段完成时正常 CompleteTask

## 3. 自动数据补全跟随同步运行（P2）

**修复** (`engine.go` + `Settings.vue`):
- `settingsWatcher` 中移除 `auto_data_complete` 的独立定时逻辑
- 在 `checkPlaylistPhaseComplete` 的下载阶段完成分支中，如果 `auto_data_complete` 开启，自动将预创建的 metadataTask 启动（此逻辑已在问题2中实现）
- `Settings.vue` 移除数据补全的间隔/单位配置UI（第101-106行），保留开关和子选项，添加说明文字"数据补全将在同步完成后自动执行"
- `saveSyncSettings` 中移除 `data_complete_interval` 和 `data_complete_unit` 的保存

## 4. SSL证书加载失败时拒绝连接（P2）

**根因**: `main.go` 中证书加载失败只打印警告，HTTPS不启动。但 `router.go` 的重定向中间件在每次请求时检查证书有效性，如果证书无效则不重定向。问题可能是证书路径存在但文件内容不完整导致 `tls.LoadX509KeyPair` 失败。

**修复** (`main.go`):
- 提取证书加载逻辑为独立函数 `loadTLSCert()`
- 首次加载失败时，启动后台 goroutine 每30秒重试
- 加载成功后启动 HTTPS 服务器

**修复** (`router.go`):
- 在 api 包新增 `HTTPSReady atomic.Bool` 变量
- `main.go` 中 HTTPS 启动成功后调用 `api.SetHTTPSReady()`
- 重定向中间件中增加 `api.HTTPSReady.Load()` 检查，HTTPS未就绪时不重定向

## 5. Login页状态同步（P1）

**根因**: Pinia store 初始化时从 localStorage 读取 token，但某些时序下可能未同步。

**修复** (`Login.vue`):
- `onMounted` 开头添加防御性恢复：从 localStorage 读取 token 同步到 systemAuth store
```ts
const savedToken = localStorage.getItem('system_token')
if (savedToken && !systemAuth.token) {
  systemAuth.token = savedToken
  systemAuth.username = localStorage.getItem('system_username')
}
```

## 6. 容器重启后自动恢复任务（P0）

**修复** (`engine.go`):
- 新增 `recoverIncompleteTasks(ctx)` 方法
- 在 `Start()` 中 worker 启动后调用
- 从数据库查询 `status IN ('pending','downloading')` 且 `phase='download'` 的记录，检查 .partial 文件，推入 downloadQueue
- 从数据库查询 `phase='metadata'` 且元数据未完成的记录，推入 metadataQueue
- 按 playlist_id 重建 PlaylistPhase 结构

## 7. 数据补全频率限制重试（P2）

**修复** (`engine.go`):
- 新增 `executeWithRetry` 辅助方法，检测频率限制错误（包含 "频率"、"rate limit"、"429" 等关键词），指数退避重试（最多5次，初始30秒，最大120秒）
- `executeMetadataTask` 中三个补全步骤使用 retry 包装
- 每个步骤执行前检查数据库中对应字段（`cover_downloaded`、`lyrics_downloaded`、`artist_completed`），已完成的跳过

## 8. 下载前扫盘跳过已下载（P1）

**修复** (`engine.go`):
- `AddPlaylistTask` 中获取 trackIDs 后，查询数据库中该歌单已完成下载的记录
- 同时扫描目标目录检查文件是否存在
- 构建已下载 songID 集合，从 trackIDs 中排除
- 如果全部已下载，直接完成 downloadTask 并进入 metadata 阶段

---

## 验证方法

1. 歌单下载 → Downloads页面立即显示"进行中"，进度条实时更新
2. 下载完成后自动显示"补全元数据"任务，进度可见
3. 同一歌单二次下载 → 跳过已下载文件
4. 系统登录后访问Login页 → 直接显示网易云登录区域
5. 配置SSL但不上传中间证书 → HTTP正常访问，不重定向
6. 下载中途kill进程 → 重启后自动续传
7. Settings页数据补全 → 只显示开关，无间隔配置
