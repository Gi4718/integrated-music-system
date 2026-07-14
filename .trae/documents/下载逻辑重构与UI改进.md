# 下载逻辑重构与 UI 改进方案

## Context

当前下载系统存在多个问题：

1. 没有断点续传能力（断电后需重新下载）
2. 下载和数据补全混合进行，不是分阶段
3. 封面 bug：所有歌曲封面都保存为同一个 `cover.jpg`
4. 没有将封面和艺人信息嵌入 MP3 ID3 标签
5. 设置页 5 个开关（自动同步、删除已移除、自动数据补全、补全封面/歌词/艺人）在后端未被实际使用
6. 点击"同步到服务器本地"后没有开始下载提示
7. 任务日志没有区分"音乐下载"和"数据补全"两个阶段
8. 文件名可能有乱码，导致 Linux 自动分文件夹

## 一、数据库 Schema 变更

**文件**: `internal/db/db.go`

在 `initTables()` 中为 `downloads` 表新增字段（用 ALTER TABLE 安全迁移）：

* `download_url TEXT` - 歌曲下载 URL（续传需要）

* `total_size INTEGER DEFAULT 0` - 文件总字节数

* `downloaded_size INTEGER DEFAULT 0` - 已下载字节数

* `sub_dir TEXT DEFAULT ''` - 子目录（歌单名）

* `playlist_id INTEGER DEFAULT 0` - 所属歌单 ID

* `phase TEXT DEFAULT 'download'` - 当前阶段: download / metadata / completed

* `cover_downloaded BOOLEAN DEFAULT 0`

* `lyrics_downloaded BOOLEAN DEFAULT 0`

* `artist_completed BOOLEAN DEFAULT 0`

* `id3_embedded BOOLEAN DEFAULT 0`

新增 `sync_tasks` 表用于持久化同步任务：

```sql
CREATE TABLE IF NOT EXISTS sync_tasks (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    playlist_id INTEGER,
    title TEXT,
    status TEXT DEFAULT 'pending',
    current INTEGER DEFAULT 0,
    total INTEGER DEFAULT 0,
    current_file TEXT DEFAULT '',
    current_bytes INTEGER DEFAULT 0,
    total_bytes INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**文件**: `internal/model/song.go` - `DownloadHistory` 结构体增加对应字段

**文件**: `internal/db/download.go` - 新增函数：

* `UpdateDownloadProgress(songID int, downloadedSize, totalSize int64)` - 实时更新下载进度

* `UpdateDownloadPhase(songID int, phase string)` - 更新阶段

* `UpdateMetadataProgress(songID int, field string)` - 更新元数据进度

* `GetPendingDownloads()` - 获取未完成下载的记录（用于断点续传恢复）

* `SaveSyncTask(task *SyncTask)` / `UpdateSyncTask(task *SyncTask)` - 同步任务持久化

## 二、文件名 Sanitization 修复

**文件**: `internal/download/engine.go` - `sanitizeFilename` 函数

修复乱码和路径分隔符问题：

1. 移除 Unicode 控制字符（零宽字符等）
2. 移除路径分隔符 `/` 和 `\`（防止创建子目录）
3. 移除 Windows 非法字符 `<>:"|?*`
4. 移除首尾空格和点
5. 按 rune 计算长度限制（避免截断多字节字符）
6. 空名后备为 `unknown`

## 三、Task 结构体增强

**文件**: `internal/service/task.go`

Task 结构体新增字段：

* `CurrentFile string` - 当前处理的文件名

* `CurrentBytes int64` - 当前文件已下载字节

* `TotalBytes int64` - 当前文件总字节

新增方法：

* `UpdateTaskCurrentFile(id string, filename string, downloaded, total int64)`

* `SetTaskStatus(id string, status TaskStatus)`

## 四、元数据补全器重构

**文件**: `internal/service/metadata_completer.go`

将 `CompleteMetadata` 拆分为三个独立方法：

* `DownloadAndEmbedCover(song, filePath, coverData []byte)` - 下载封面并嵌入 ID3 APIC

* `DownloadLyrics(song, filePath)` - 下载歌词保存为 `.lrc` 文件

* `EmbedArtistInfo(song, filePath)` - 嵌入艺人信息到 ID3

**封面 bug 修复**: 不再保存独立的 `cover.jpg`，直接将封面数据嵌入 MP3 的 ID3 APIC frame。

**ID3 嵌入**: 使用 `github.com/bogem/id3v2` 库，设置 Title/Artist/Album + APIC 封面。

**歌词文件命名**: 歌词文件名 = 歌曲文件名（去掉扩展名）+ `.lrc`，与歌曲同名。

## 五、下载引擎重构（核心）

**文件**: `internal/download/engine.go`

### 5.1 Engine 结构体改造

* 新增 `metadataQueue chan *DownloadTask` - 数据补全队列

* 新增 `taskService *service.TaskService` - 注入任务服务

* 新增 `playlistPhases map[int]*PlaylistPhase` - 歌单阶段追踪

### 5.2 两阶段下载流程

1. `AddPlaylistTask` 创建"音乐下载"任务（TaskService），所有歌曲入 `downloadQueue`
2. `executeTask` 只负责下载歌曲文件（支持断点续传）
3. 下载完成后将 task 放入 `metadataQueue`（不直接补全）
4. 所有歌曲下载完成后，标记下载任务完成，创建"数据补全"任务
5. `metadataWorkerLoop` 从 `metadataQueue` 取任务执行数据补全

### 5.3 断点续传实现

* 使用 `.partial` 临时文件

* 下载前检查 `.partial` 文件大小，设置 Range header

* 以 Append 模式写入

* 每 256KB 持久化一次进度到 DB

* 下载完成后 rename `.partial` -> 最终文件名

### 5.4 设置开关生效

在 `executeMetadataTask` 中读取设置：

* `data_complete_cover` -> 控制封面下载和嵌入

* `data_complete_lyrics` -> 控制歌词下载

* `data_complete_artist` -> 控制艺人信息嵌入

在 Engine 启动时启动 `settingsWatcher` goroutine：

* `auto_sync` -> 定时执行歌单同步

* `delete_removed` -> 同步后删除不在歌单中的本地文件

* `auto_data_complete` -> 定时扫描未补全的记录

## 六、API 层修改

**文件**: `internal/api/engine.go` - 注入 TaskService 到 Engine

**文件**: `internal/api/download.go` - `downloadPlaylist` 返回 `download_task_id` 和 `metadata_task_id`

**文件**: `internal/api/task.go` - 确保返回 `current_file`、`current_bytes`、`total_bytes` 等新字段

## 七、前端 UI 改进

### 7.1 Downloads.vue - 双任务进度显示

新增"正在进行的任务"区域，显示两个任务卡片：

* 音乐下载任务：进度条 + 当前/总数 + 当前文件名 + 字节进度

* 数据补全任务：进度条 + 当前/总数 + 当前文件名

2 秒轮询 `/api/tasks` 接口获取实时进度。

### 7.2 Playlist.vue - 同步按钮反馈

修改 `syncToLocal` 方法，提示文案改为 `开始下载歌单「xxx」，共 N 首歌曲`。

### 7.3 Settings.vue - 无需修改

设置页面 UI 已完整，只需确保后端实际使用这些设置值。

## 八、实现顺序

1. 数据库迁移（db.go + download.go + model/song.go）
2. 文件名 sanitization 修复（engine.go）
3. Task 结构体增强（service/task.go）
4. 元数据补全器拆分 + ID3 嵌入（metadata\_completer.go）
5. 下载引擎重构（engine.go）- 两阶段 + 断点续传 + 设置读取
6. API 层适配（download.go, engine.go, task.go）
7. 前端 UI（Downloads.vue, Playlist.vue）
8. 构建部署到 SSH 服务器
9. 端到端测试验证

## 九、验证方案

1. 启动下载歌单（50+ 首），观察任务日志是否显示两个任务
2. 下载过程中断容器，重启后验证断点续传
3. 检查 MP3 文件是否包含正确的 ID3 标签（封面、艺人、专辑）
4. 检查 `.lrc` 歌词文件是否正确生成
5. 验证设置开关：关闭"补全封面"后封面不被嵌入
6. 验证文件名无乱码、无路径分隔符
7. 验证 Windows 文件传输详情风格的进度显示

