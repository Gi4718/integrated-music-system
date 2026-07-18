package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"endfield-music/internal/db"
	"endfield-music/internal/model"
	"endfield-music/internal/service"
	"endfield-music/internal/util"
)

// Engine 下载引擎
type Engine struct {
	netease           *service.NeteaseService
	metadataCompleter *service.MetadataCompleter
	taskService       *service.TaskService
	rateLimiter       *util.RateLimiter
	downloadQueue     chan *DownloadTask
	metadataQueue     chan *DownloadTask
	worker            int
	mu                sync.RWMutex
	tasks             map[int]*DownloadTask
	playlistPhases    map[int]*PlaylistPhase
}

// DownloadTask 下载任务
type DownloadTask struct {
	ID            int
	SongID        int
	SongName      string
	Artist        string
	Album         string
	Quality       string
	SubDir        string
	PlaylistID    int
	FilePath      string
	Status        string
	Progress      float64
	Error         string
	CreatedAt     time.Time
	DownloadURL   string
	TotalSize     int64
	DownloadedSize int64
	Phase         string
	TaskServiceID string // 关联的 TaskService 任务ID
}

// PlaylistPhase 歌单阶段追踪
type PlaylistPhase struct {
	PlaylistID     int
	PlaylistName   string
	DownloadTaskID string
	MetadataTaskID string
	TotalSongs     int
	DownloadDone   int
	MetadataDone   int
	Phase          string
	mu             sync.Mutex
}

// NewEngine 创建下载引擎
func NewEngine(netease *service.NeteaseService, taskService *service.TaskService, concurrency int) *Engine {
	return &Engine{
		netease:           netease,
		metadataCompleter: service.NewMetadataCompleter(netease),
		taskService:       taskService,
		rateLimiter:       util.NewRateLimiter(1 * time.Second),
		downloadQueue:     make(chan *DownloadTask, 200),
		metadataQueue:     make(chan *DownloadTask, 200),
		worker:            concurrency,
		tasks:             make(map[int]*DownloadTask),
		playlistPhases:    make(map[int]*PlaylistPhase),
	}
}

// Start 启动工作协程
func (e *Engine) Start(ctx context.Context) {
	for i := 0; i < e.worker; i++ {
		go e.workerLoop(ctx)
		go e.metadataWorkerLoop(ctx)
	}
	// 恢复未完成的任务（断点续传），异步执行避免阻塞 HTTP 服务器启动
	go e.recoverIncompleteTasks(ctx)
	go e.settingsWatcher(ctx)
	// 启动定时持久化（每10秒）
	go e.persistTasksPeriodically(ctx)
}

// persistTasksPeriodically 每10秒持久化所有活跃任务状态到数据库
func (e *Engine) persistTasksPeriodically(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.persistActiveTasks()
		}
	}
}

// persistActiveTasks 持久化所有活跃任务
func (e *Engine) persistActiveTasks() {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, task := range e.tasks {
		// 只持久化进行中的任务
		if task.Status == "downloading" || task.Status == "pending" {
			// 更新下载进度
			if task.DownloadedSize > 0 || task.TotalSize > 0 {
				db.UpdateDownloadProgress(task.SongID, task.DownloadedSize, task.TotalSize)
			}
			// 更新状态
			db.UpdateDownloadStatus(task.SongID, task.Status)
		}
	}
}

// isTaskCancelled 检查任务是否被取消
func (e *Engine) isTaskCancelled(taskID string) bool {
	if e.taskService == nil {
		return false
	}
	task := e.taskService.GetTask(taskID)
	if task == nil {
		return false
	}
	return task.Status == service.TaskStatusCancelled
}

// recoverIncompleteTasks 从数据库恢复未完成的任务
func (e *Engine) recoverIncompleteTasks(ctx context.Context) {
	pending, err := db.GetPendingDownloads()
	if err != nil {
		fmt.Printf("[recover] failed to get pending downloads: %v\n", err)
		return
	}

	// 按歌单分组重建 PlaylistPhase
	phaseMap := make(map[int]*PlaylistPhase)

	for _, d := range pending {
		if d.Phase == "download" && (d.Status == "pending" || d.Status == "downloading") {
			// 重新计算文件路径（DB中FilePath可能为空，因为只在下载完成后才写入）
			ext := ".mp3"
			if d.Quality == "lossless" {
				ext = ".flac"
			}
			songFormat, _ := db.GetSetting("song_format")
			if songFormat == "" {
				songFormat = "{songName} - {artist}"
			}
			formattedName := strings.ReplaceAll(songFormat, "{songName}", d.SongName)
			formattedName = strings.ReplaceAll(formattedName, "{artist}", d.Artist)
			filename := sanitizeFilename(formattedName) + ext
			baseDir := "/music"
			if d.SubDir != "" {
				baseDir = filepath.Join("/music", d.SubDir)
			}
			computedPath := filepath.Join(baseDir, filename)

			// 检查文件是否已存在（下载已完成但状态未更新）
			if _, err := os.Stat(computedPath); err == nil {
				db.UpdateDownloadStatus(d.SongID, "completed")
				db.UpdateDownloadPhase(d.SongID, "metadata")
				fmt.Printf("[recover] file already exists: %s, marked completed\n", computedPath)
				continue
			}

			// 检查 .partial 文件是否存在，恢复断点续传
			partialPath := computedPath + ".partial"
			if info, err := os.Stat(partialPath); err == nil {
				d.DownloadedSize = info.Size()
				d.FilePath = computedPath
				fmt.Printf("[recover] found partial file %s (%d bytes), resuming\n", partialPath, info.Size())
			} else {
				d.DownloadedSize = 0
				d.FilePath = computedPath
				db.UpdateDownloadStatus(d.SongID, "pending")
			}

			// 获取或创建 TaskService 任务 ID
			var taskServiceID string
			if d.PlaylistID > 0 {
				if phase, ok := phaseMap[d.PlaylistID]; ok {
					taskServiceID = phase.DownloadTaskID
				}
			}

			task := &DownloadTask{
				SongID:         d.SongID,
				SongName:       d.SongName,
				Artist:         d.Artist,
				Album:          d.Album,
				Quality:        d.Quality,
				SubDir:         d.SubDir,
				PlaylistID:     d.PlaylistID,
				FilePath:       d.FilePath,
				Status:         "pending",
				Phase:          "download",
				DownloadURL:    d.DownloadURL,
				TotalSize:      d.TotalSize,
				DownloadedSize: d.DownloadedSize,
				TaskServiceID:  taskServiceID,
			}
			e.mu.Lock()
			task.ID = len(e.tasks) + 1
			e.tasks[task.ID] = task
			e.mu.Unlock()

			// 重建 PlaylistPhase
			if d.PlaylistID > 0 {
				if _, ok := phaseMap[d.PlaylistID]; !ok {
					downloadTask := e.taskService.CreateTask(service.TaskTypeDownload,
						fmt.Sprintf("恢复下载歌单「%s」", d.SubDir), "")
					metadataTask := e.taskService.CreateTask(service.TaskTypeDataComplete,
						fmt.Sprintf("补全歌单「%s」元数据", d.SubDir), "")
					phase := &PlaylistPhase{
						PlaylistID:     d.PlaylistID,
						PlaylistName:   d.SubDir,
						DownloadTaskID: downloadTask.ID,
						MetadataTaskID: metadataTask.ID,
						TotalSongs:     0, // 后续统计
						Phase:          "downloading",
					}
					phaseMap[d.PlaylistID] = phase
					e.mu.Lock()
					e.playlistPhases[d.PlaylistID] = phase
					e.mu.Unlock()
				}
				phaseMap[d.PlaylistID].TotalSongs++
			}

			e.downloadQueue <- task
			fmt.Printf("[recover] queued download task: %s (songID=%d)\n", d.SongName, d.SongID)
		} else if d.Phase == "metadata" && !d.MetadataCompleted {
			task := &DownloadTask{
				SongID:     d.SongID,
				SongName:   d.SongName,
				Artist:     d.Artist,
				Album:      d.Album,
				FilePath:   d.FilePath,
				SubDir:     d.SubDir,
				PlaylistID: d.PlaylistID,
				Status:     "pending",
				Phase:      "metadata",
			}
			e.mu.Lock()
			task.ID = len(e.tasks) + 1
			e.tasks[task.ID] = task
			e.mu.Unlock()

			// 重建 PlaylistPhase（如果还没有）
			if d.PlaylistID > 0 {
				if _, ok := phaseMap[d.PlaylistID]; !ok {
					metadataTask := e.taskService.CreateTask(service.TaskTypeDataComplete,
						fmt.Sprintf("恢复补全歌单「%s」元数据", d.SubDir), "")
					phase := &PlaylistPhase{
						PlaylistID:     d.PlaylistID,
						PlaylistName:   d.SubDir,
						MetadataTaskID: metadataTask.ID,
						Phase:          "metadata",
					}
					phaseMap[d.PlaylistID] = phase
					e.mu.Lock()
					e.playlistPhases[d.PlaylistID] = phase
					e.mu.Unlock()
				}
				phaseMap[d.PlaylistID].TotalSongs++
				phaseMap[d.PlaylistID].DownloadDone++
			}

			e.metadataQueue <- task
			fmt.Printf("[recover] queued metadata task: %s (songID=%d)\n", d.SongName, d.SongID)
		}
	}

	if len(phaseMap) > 0 {
		fmt.Printf("[recover] restored %d playlist phases\n", len(phaseMap))
	}
}

func (e *Engine) workerLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-e.downloadQueue:
			// 检查任务是否被取消
			if task.PlaylistID > 0 {
				e.mu.RLock()
				phase, ok := e.playlistPhases[task.PlaylistID]
				e.mu.RUnlock()
				if ok && e.isTaskCancelled(phase.DownloadTaskID) {
					fmt.Printf("[worker] task cancelled for playlist %s, skipping song %s\n", phase.PlaylistName, task.SongName)
					db.UpdateDownloadStatus(task.SongID, "cancelled")
					e.mu.Lock()
					task.Status = "cancelled"
					e.mu.Unlock()
					e.checkPlaylistPhaseComplete(task.PlaylistID)
					continue
				}
			}
			e.executeTask(ctx, task)
		}
	}
}

func (e *Engine) metadataWorkerLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-e.metadataQueue:
			// 检查任务是否被取消
			if task.PlaylistID > 0 {
				e.mu.RLock()
				phase, ok := e.playlistPhases[task.PlaylistID]
				e.mu.RUnlock()
				if ok && e.isTaskCancelled(phase.MetadataTaskID) {
					fmt.Printf("[metadata-worker] task cancelled for playlist %s, skipping %s\n", phase.PlaylistName, task.SongName)
					continue
				}
			}
			e.executeMetadataTask(ctx, task)
		}
	}
}

// AddTask 添加下载任务
func (e *Engine) AddTask(songID int, quality string) (int, error) {
	return e.AddTaskWithSubDir(songID, quality, "", 0)
}

// AddTaskWithSubDir 添加带子目录的下载任务
func (e *Engine) AddTaskWithSubDir(songID int, quality string, subDir string, playlistID int) (int, error) {
	e.rateLimiter.Wait()

	body, err := e.netease.GetSongDetail(songID)
	if err != nil {
		return 0, fmt.Errorf("获取歌曲详情失败: %w", err)
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	songs, ok := result["songs"].([]interface{})
	if !ok || len(songs) == 0 {
		return 0, fmt.Errorf("未找到歌曲")
	}

	song := songs[0].(map[string]interface{})
	songName := ""
	if v, ok := song["name"].(string); ok {
		songName = v
	}
	artist := ""
	if ar, ok := song["ar"].([]interface{}); ok && len(ar) > 0 {
		if a, ok := ar[0].(map[string]interface{}); ok {
			artist, _ = a["name"].(string)
		}
	}
	album := ""
	if al, ok := song["al"].(map[string]interface{}); ok {
		album, _ = al["name"].(string)
	}

	// 单曲下载需要创建 TaskService 任务以显示在任务日志中
	var taskServiceID string
	if playlistID == 0 && e.taskService != nil {
		tsk := e.taskService.CreateTask(service.TaskTypeDownload,
			fmt.Sprintf("下载单曲「%s」", songName), "")
		tsk.Total = 1
		taskServiceID = tsk.ID
	}

	task := &DownloadTask{
		SongID:        songID,
		SongName:      songName,
		Artist:        artist,
		Album:         album,
		Quality:       quality,
		SubDir:        subDir,
		PlaylistID:    playlistID,
		Status:        "pending",
		CreatedAt:     time.Now(),
		TaskServiceID: taskServiceID,
	}

	e.mu.Lock()
	task.ID = len(e.tasks) + 1
	e.tasks[task.ID] = task
	e.mu.Unlock()

	db.SaveDownloadHistory(&model.DownloadHistory{
		SongID:     songID,
		SongName:   songName,
		Artist:     artist,
		Album:      album,
		Quality:    quality,
		Status:     "pending",
		SubDir:     subDir,
		PlaylistID: playlistID,
	})

	e.downloadQueue <- task
	return task.ID, nil
}

// AddPlaylistTask 添加歌单下载任务（两阶段）
func (e *Engine) AddPlaylistTask(playlistID int, quality string) ([]int, string, string, error) {
	fmt.Printf("[AddPlaylistTask] starting for playlistID=%d, quality=%s\n", playlistID, quality)
	
	e.rateLimiter.Wait()
	fmt.Printf("[AddPlaylistTask] rate limiter passed\n")

	// 重复任务检测：检查是否已有同歌单的活跃任务
	e.mu.Lock()
	if phase, ok := e.playlistPhases[playlistID]; ok {
		phase.mu.Lock()
		isActive := phase.Phase == "scanning" || phase.Phase == "downloading" || phase.Phase == "metadata"
		phase.mu.Unlock()
		if isActive {
			e.mu.Unlock()
			fmt.Printf("[AddPlaylistTask] duplicate task detected for playlistID=%d (phase=%s)\n", playlistID, phase.Phase)
			return nil, "", "", fmt.Errorf("该歌单已有下载任务正在进行中")
		}
		// 旧任务已结束（完成/失败/终止），清理旧 phase 和旧任务，允许重试
		fmt.Printf("[AddPlaylistTask] cleaning up old phase for playlistID=%d (phase=%s)\n", playlistID, phase.Phase)
		// 取消该歌单的旧任务
		for _, t := range e.tasks {
			if t.PlaylistID == playlistID && (t.Status == "pending" || t.Status == "downloading") {
				t.Status = "cancelled"
			}
		}
		delete(e.playlistPhases, playlistID)
	}
	e.mu.Unlock()

	cookie, _ := db.GetCookie()
	body, err := e.netease.GetPlaylistDetail(playlistID, cookie)
	if err != nil {
		fmt.Printf("[AddPlaylistTask] failed to get playlist detail: %v\n", err)
		return nil, "", "", fmt.Errorf("获取歌单详情失败: %w", err)
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	playlist, ok := result["playlist"].(map[string]interface{})
	if !ok {
		fmt.Printf("[AddPlaylistTask] failed to parse playlist\n")
		return nil, "", "", fmt.Errorf("解析歌单失败")
	}

	playlistName := ""
	if name, ok := playlist["name"].(string); ok {
		playlistName = sanitizeFilename(name)
	}

	// 使用 trackIds 获取完整歌单（tracks 只返回前1000首）
	var trackIDs []int
	seen := make(map[int]bool)
	if tids, ok := playlist["trackIds"].([]interface{}); ok {
		for _, tid := range tids {
			if t, ok := tid.(map[string]interface{}); ok {
				if id, ok := t["id"]; ok {
					var songID int
					switch v := id.(type) {
					case float64:
						songID = int(v)
					}
					if songID > 0 && !seen[songID] {
						seen[songID] = true
						trackIDs = append(trackIDs, songID)
					}
				}
			}
		}
	}
	// 回退到 tracks
	if len(trackIDs) == 0 {
		if tracks, ok := playlist["tracks"].([]interface{}); ok && len(tracks) > 0 {
			for _, t := range tracks {
				if track, ok := t.(map[string]interface{}); ok {
					if id, ok := track["id"]; ok {
						var songID int
						switch v := id.(type) {
						case float64:
							songID = int(v)
						}
						if songID > 0 && !seen[songID] {
							seen[songID] = true
							trackIDs = append(trackIDs, songID)
						}
					}
				}
			}
		}
	}

	fmt.Printf("[AddPlaylistTask] raw trackIDs count: %d (after dedup)\n", len(trackIDs))

	if len(trackIDs) == 0 {
		fmt.Printf("[AddPlaylistTask] no tracks found\n")
		return nil, "", "", fmt.Errorf("歌单无曲目")
	}

	// 创建扫描任务
	scanTask := e.taskService.CreateTask(service.TaskTypeScan,
		fmt.Sprintf("扫描歌单「%s」", playlistName), "")
	scanTask.Total = len(trackIDs)

	// 创建歌单阶段追踪（扫描阶段）
	phase := &PlaylistPhase{
		PlaylistID:   playlistID,
		PlaylistName: playlistName,
		TotalSongs:   len(trackIDs),
		Phase:        "scanning",
	}
	e.mu.Lock()
	e.playlistPhases[playlistID] = phase
	e.mu.Unlock()

	// 异步执行扫描
	go e.asyncScanAndDownload(playlistID, playlistName, trackIDs, quality, scanTask)

	return nil, "", "", nil
}

// SongInfo 扫描阶段获取的歌曲信息（用于复用，避免重复 API 调用）
type SongInfo struct {
	SongID int
	Name   string
	Artist string
	Album  string
}

// asyncScanAndDownload 异步执行扫描和下载
func (e *Engine) asyncScanAndDownload(playlistID int, playlistName string, trackIDs []int, quality string, scanTask *service.Task) {
	ctx := context.Background()
	
	// 执行扫描
	e.taskService.SetTaskStatus(scanTask.ID, service.TaskStatusRunning)
	skippedSongIDs, remainingInfos, copiedSongIDs := e.scanPlaylistSongs(ctx, trackIDs, quality, playlistName, playlistID, scanTask)
	
	skippedCount := len(skippedSongIDs)
	copiedCount := len(copiedSongIDs)
	remainingCount := len(remainingInfos)
	
	fmt.Printf("[asyncScanAndDownload] scan complete: %d remaining, %d skipped, %d copied\n", remainingCount, skippedCount, copiedCount)
	
	// 完成扫描任务
	e.taskService.UpdateTaskProgress(scanTask.ID, len(trackIDs), len(trackIDs))
	e.taskService.CompleteTask(scanTask.ID)
	
	// 创建下载和补全任务
	downloadTask := e.taskService.CreateTask(service.TaskTypeDownload,
		fmt.Sprintf("下载歌单「%s」", playlistName), "")
	downloadTask.Total = len(trackIDs)
	
	metadataTask := e.taskService.CreateTask(service.TaskTypeDataComplete,
		fmt.Sprintf("补全歌单「%s」元数据", playlistName), "")
	metadataTask.Total = len(trackIDs)
	
	// 更新 phase
	e.mu.Lock()
	phase, ok := e.playlistPhases[playlistID]
	if ok {
		phase.DownloadTaskID = downloadTask.ID
		phase.MetadataTaskID = metadataTask.ID
		phase.Phase = "downloading"
	}
	e.mu.Unlock()
	
	// 为跳过的歌曲创建虚拟已完成任务
	for _, songID := range skippedSongIDs {
		history, _ := db.GetDownloadBySongID(songID)
		if history == nil {
			continue
		}
		
		virtualTask := &DownloadTask{
			SongID:     songID,
			SongName:   history.SongName,
			Artist:     history.Artist,
			Album:      history.Album,
			Quality:    history.Quality,
			SubDir:     playlistName,
			PlaylistID: playlistID,
			FilePath:   history.FilePath,
			Status:     "completed",
			Phase:      "download",
			CreatedAt:  time.Now(),
		}
		e.mu.Lock()
		virtualTask.ID = len(e.tasks) + 1
		e.tasks[virtualTask.ID] = virtualTask
		e.mu.Unlock()
	}
	
	// 为复制的歌曲创建虚拟已完成任务
	for _, songID := range copiedSongIDs {
		history, _ := db.GetAnyDownloadedSong(songID)
		if history == nil {
			continue
		}
		
		// 复制文件到目标目录
		targetDir := "/music"
		if playlistName != "" {
			targetDir = filepath.Join("/music", playlistName)
		}
		os.MkdirAll(targetDir, 0755)
		
		ext := filepath.Ext(history.FilePath)
		songFormat, _ := db.GetSetting("song_format")
		if songFormat == "" {
			songFormat = "{songName} - {artist}"
		}
		formattedName := strings.ReplaceAll(songFormat, "{songName}", history.SongName)
		formattedName = strings.ReplaceAll(formattedName, "{artist}", history.Artist)
		filename := sanitizeFilename(formattedName) + ext
		targetPath := filepath.Join(targetDir, filename)
		
		// 复制文件
		if err := copyFile(history.FilePath, targetPath); err != nil {
			fmt.Printf("[asyncScanAndDownload] failed to copy %s: %v\n", history.SongName, err)
			continue
		}
		
		// 保存下载记录
		db.SaveDownloadHistory(&model.DownloadHistory{
			SongID:     songID,
			SongName:   history.SongName,
			Artist:     history.Artist,
			Album:      history.Album,
			Quality:    history.Quality,
			Status:     "completed",
			FilePath:   targetPath,
			FileSize:   history.FileSize,
			SubDir:     playlistName,
			PlaylistID: playlistID,
			Phase:      "download",
		})
		
		virtualTask := &DownloadTask{
			SongID:     songID,
			SongName:   history.SongName,
			Artist:     history.Artist,
			Album:      history.Album,
			Quality:    history.Quality,
			SubDir:     playlistName,
			PlaylistID: playlistID,
			FilePath:   targetPath,
			Status:     "completed",
			Phase:      "download",
			CreatedAt:  time.Now(),
		}
		e.mu.Lock()
		virtualTask.ID = len(e.tasks) + 1
		e.tasks[virtualTask.ID] = virtualTask
		e.mu.Unlock()
	}
	
	// 更新进度
	totalSkipped := skippedCount + copiedCount
	if totalSkipped > 0 {
		phase.DownloadDone = totalSkipped
		e.taskService.UpdateTaskProgress(downloadTask.ID, totalSkipped, len(trackIDs))
	}
	
	// 全部已完成，进入元数据阶段
	if remainingCount == 0 {
		fmt.Printf("[asyncScanAndDownload] all songs already downloaded, skipping to metadata\n")
		e.taskService.UpdateTaskProgress(downloadTask.ID, len(trackIDs), len(trackIDs))
		e.taskService.CompleteTask(downloadTask.ID)
		e.taskService.SetTaskStatus(metadataTask.ID, service.TaskStatusRunning)
		phase.Phase = "metadata"
		
		// 推入元数据队列（只推入元数据未完成的歌曲）
		history, _ := db.GetDownloadHistory()
		metadataNeeded := 0
		metadataDone := 0
		for _, h := range history {
			if h.PlaylistID == playlistID && h.Status == "completed" && h.FilePath != "" {
				if h.MetadataCompleted {
					// 元数据已完成，跳过
					metadataDone++
					continue
				}
				metadataNeeded++
				task := &DownloadTask{
					SongID:     h.SongID,
					SongName:   h.SongName,
					Artist:     h.Artist,
					Album:      h.Album,
					FilePath:   h.FilePath,
					SubDir:     h.SubDir,
					PlaylistID: playlistID,
					Phase:      "metadata",
				}
				e.mu.Lock()
				task.ID = len(e.tasks) + 1
				e.tasks[task.ID] = task
				e.mu.Unlock()
				e.metadataQueue <- task
			}
		}
		
		// 更新补全任务进度（已完成的直接计入）
		if metadataDone > 0 {
			phase.MetadataDone = metadataDone
			e.taskService.UpdateTaskProgress(metadataTask.ID, metadataDone, len(trackIDs))
		}
		
		// 如果所有歌曲元数据都已完成，直接完成补全任务
		if metadataNeeded == 0 {
			fmt.Printf("[asyncScanAndDownload] all metadata already completed\n")
			e.taskService.UpdateTaskProgress(metadataTask.ID, len(trackIDs), len(trackIDs))
			e.taskService.CompleteTask(metadataTask.ID)
			phase.Phase = "completed"
		}
		
		return
	}
	
	// 创建下载任务（复用扫描阶段获取的歌曲信息，避免重复API调用）
	for _, info := range remainingInfos {
		// 直接创建任务，不再调用AddTaskWithSubDir（避免重复GetSongDetail）
		task := &DownloadTask{
			SongID:     info.SongID,
			SongName:   info.Name,
			Artist:     info.Artist,
			Album:      info.Album,
			Quality:    quality,
			SubDir:     playlistName,
			PlaylistID: playlistID,
			Status:     "pending",
			CreatedAt:  time.Now(),
		}
		
		e.mu.Lock()
		task.ID = len(e.tasks) + 1
		e.tasks[task.ID] = task
		e.mu.Unlock()
		
		db.SaveDownloadHistory(&model.DownloadHistory{
			SongID:     info.SongID,
			SongName:   info.Name,
			Artist:     info.Artist,
			Album:      info.Album,
			Quality:    quality,
			Status:     "pending",
			SubDir:     playlistName,
			PlaylistID: playlistID,
		})
		
		e.downloadQueue <- task
		fmt.Printf("[asyncScanAndDownload] queued download task: %s (songID=%d)\n", info.Name, info.SongID)
	}
}

// scanPlaylistSongs 扫描歌单歌曲，返回：(已跳过的歌曲ID, 需要下载的歌曲信息, 可复制的歌曲ID)
func (e *Engine) scanPlaylistSongs(ctx context.Context, trackIDs []int, quality, playlistName string, playlistID int, scanTask *service.Task) ([]int, []SongInfo, []int) {
	var skipped []int
	var remaining []SongInfo
	var copied []int
	
	// 构建目标目录
	targetDir := "/music"
	if playlistName != "" {
		targetDir = filepath.Join("/music", playlistName)
	}
	
	// 扫描每首歌曲
	for i, songID := range trackIDs {
		// 更新扫描进度
		e.taskService.UpdateTaskProgress(scanTask.ID, i+1, len(trackIDs))
		e.taskService.UpdateTaskCurrentFile(scanTask.ID, fmt.Sprintf("扫描中... (%d/%d)", i+1, len(trackIDs)), 0, 0)
		
		// 先检查数据库是否有下载记录（避免不必要的API调用）
		history, _ := db.GetDownloadBySongID(songID)
		if history != nil && history.FilePath != "" {
			// 数据库有记录，检查文件是否存在
			if _, err := os.Stat(history.FilePath); err == nil {
				// 文件存在，跳过
				skipped = append(skipped, songID)
				fmt.Printf("[scan] song already downloaded: %s (songID=%d)\n", history.SongName, songID)
				
				// 更新扫盘记录
				db.SaveScanRecord(songID, history.SongName, history.Artist, history.Album, history.FilePath,
					history.FileSize, time.Now(), true, true, history.MetadataCompleted, playlistID, playlistName)
				continue
			}
		}
		
		// 检查是否在其他歌单已下载（跨歌单复制）
		if otherHistory, _ := db.GetAnyDownloadedSong(songID); otherHistory != nil {
			if _, err := os.Stat(otherHistory.FilePath); err == nil {
				copied = append(copied, songID)
				fmt.Printf("[scan] song available in other playlist: %s (songID=%d)\n", otherHistory.SongName, songID)
				continue
			}
		}
		
		// 需要下载，获取歌曲详情
		e.rateLimiter.Wait()
		body, err := e.netease.GetSongDetail(songID)
		if err != nil {
			fmt.Printf("[scan] failed to get song detail for %d: %v\n", songID, err)
			remaining = append(remaining, SongInfo{SongID: songID})
			continue
		}
		
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		
		// 检测限速
		if code, ok := result["code"]; ok {
			switch v := code.(type) {
			case float64:
				if v == -1 || v == 429 || v == 301 {
					e.rateLimiter.Increase()
					fmt.Printf("[ratelimit] 扫描API限速，间隔调整为 %v\n", e.rateLimiter.GetInterval())
					remaining = append(remaining, SongInfo{SongID: songID})
					continue
				}
			}
		}
		e.rateLimiter.Decrease()
		
		songs, ok := result["songs"].([]interface{})
		if !ok || len(songs) == 0 {
			remaining = append(remaining, SongInfo{SongID: songID})
			continue
		}
		
		song := songs[0].(map[string]interface{})
		songName := ""
		if v, ok := song["name"].(string); ok {
			songName = v
		}
		artist := ""
		if ar, ok := song["ar"].([]interface{}); ok && len(ar) > 0 {
			if a, ok := ar[0].(map[string]interface{}); ok {
				artist, _ = a["name"].(string)
			}
		}
		album := ""
		if al, ok := song["al"].(map[string]interface{}); ok {
			album, _ = al["name"].(string)
		}
		
		// 检查是否已在本歌单下载（通过文件名匹配）
		ext := ".mp3"
		if quality == "lossless" {
			ext = ".flac"
		}
		songFormat, _ := db.GetSetting("song_format")
		if songFormat == "" {
			songFormat = "{songName} - {artist}"
		}
		formattedName := strings.ReplaceAll(songFormat, "{songName}", songName)
		formattedName = strings.ReplaceAll(formattedName, "{artist}", artist)
		filename := sanitizeFilename(formattedName) + ext
		targetPath := filepath.Join(targetDir, filename)
		
		if info, err := os.Stat(targetPath); err == nil {
			// 文件存在，检查元数据状态
			lyricsPath := strings.TrimSuffix(targetPath, ext) + ".lrc"
			_, lyricsErr := os.Stat(lyricsPath)
			
			metadataCompleted := history != nil && history.MetadataCompleted
			
			// 保存扫盘记录
			db.SaveScanRecord(songID, songName, artist, album, targetPath, info.Size(), info.ModTime(),
				true, lyricsErr == nil, metadataCompleted, playlistID, playlistName)
			
			skipped = append(skipped, songID)
			fmt.Printf("[scan] song already exists: %s\n", songName)
			continue
		}
		
		// 需要下载，保存歌曲信息（避免后续重复调用API）
		remaining = append(remaining, SongInfo{
			SongID: songID,
			Name:   songName,
			Artist: artist,
			Album:  album,
		})
	}
	
	fmt.Printf("[scanPlaylistSongs] result: %d skipped, %d copied, %d remaining\n", len(skipped), len(copied), len(remaining))
	return skipped, remaining, copied
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	return nil
}


// executeTask 执行下载任务（第一阶段：只下载歌曲文件）
func (e *Engine) executeTask(ctx context.Context, task *DownloadTask) {
	e.mu.Lock()
	task.Status = "downloading"
	e.mu.Unlock()

	// 立即更新 TaskService 状态为 running，让前端可见
	if e.taskService != nil {
		if task.PlaylistID > 0 {
			// 歌单下载：更新歌单任务进度
			e.mu.RLock()
			phase, ok := e.playlistPhases[task.PlaylistID]
			e.mu.RUnlock()
			if ok {
				e.taskService.SetTaskStatus(phase.DownloadTaskID, service.TaskStatusRunning)
				// 更新当前下载计数（正在下载也算进度）
				e.mu.RLock()
				downloading := 0
				for _, t := range e.tasks {
					if t.PlaylistID == task.PlaylistID && (t.Status == "completed" || t.Status == "downloading") {
						downloading++
					}
				}
				e.mu.RUnlock()
				e.taskService.UpdateTaskProgress(phase.DownloadTaskID, downloading, phase.TotalSongs)
			}
		} else if task.TaskServiceID != "" {
			// 单曲下载：更新单曲任务状态
			e.taskService.SetTaskStatus(task.TaskServiceID, service.TaskStatusRunning)
			e.taskService.UpdateTaskProgress(task.TaskServiceID, 0, 1)
		}
	}

	db.UpdateDownloadStatus(task.SongID, "downloading")

	br := 320000
	if task.Quality == "standard" {
		br = 128000
	} else if task.Quality == "lossless" {
		br = 999000
	}

	e.rateLimiter.Wait()

	cookie, _ := db.GetCookie()
	body, err := e.netease.GetSongURL(task.SongID, br, cookie)
	if err != nil {
		e.failTask(task, err.Error())
		e.checkPlaylistPhaseComplete(task.PlaylistID)
		return
	}

	var urlResult map[string]interface{}
	json.Unmarshal(body, &urlResult)

	// 检测限速
	if code, ok := urlResult["code"]; ok {
		switch v := code.(type) {
		case float64:
			if v == -1 || v == 429 || v == 301 {
				e.rateLimiter.Increase()
				fmt.Printf("[ratelimit] 下载API限速，间隔调整为 %v\n", e.rateLimiter.GetInterval())
				e.failTask(task, "API 限速，稍后重试")
				e.checkPlaylistPhaseComplete(task.PlaylistID)
				return
			}
		}
	}

	data, ok := urlResult["data"].([]interface{})
	if !ok || len(data) == 0 {
		e.failTask(task, "获取歌曲 URL 失败")
		e.checkPlaylistPhaseComplete(task.PlaylistID)
		return
	}

	songData := data[0].(map[string]interface{})
	url, _ := songData["url"].(string)
	if url == "" {
		code, _ := songData["code"].(float64)
		if code == -110 {
			e.failTask(task, "需要 VIP 或版权限制")
		} else if code == -100 {
			e.failTask(task, "歌曲不存在")
		} else {
			e.failTask(task, fmt.Sprintf("获取 URL 失败 (code=%.0f)", code))
		}
		e.checkPlaylistPhaseComplete(task.PlaylistID)
		return
	}

	// API 调用成功，逐步降低限速间隔
	e.rateLimiter.Decrease()

	ext := ".mp3"
	if task.Quality == "lossless" {
		ext = ".flac"
	}

	// 从设置读取文件名格式，默认 {songName} - {artist}
	songFormat, _ := db.GetSetting("song_format")
	if songFormat == "" {
		songFormat = "{songName} - {artist}"
	}
	formattedName := strings.ReplaceAll(songFormat, "{songName}", task.SongName)
	formattedName = strings.ReplaceAll(formattedName, "{artist}", task.Artist)
	filename := sanitizeFilename(formattedName) + ext

	baseDir := "/music"
	if task.SubDir != "" {
		baseDir = filepath.Join("/music", task.SubDir)
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			e.failTask(task, fmt.Sprintf("创建目录失败: %v", err))
			e.checkPlaylistPhaseComplete(task.PlaylistID)
			return
		}
	}
	filePath := filepath.Join(baseDir, filename)
	task.FilePath = filePath

	// 更新 TaskService 当前文件
	if e.taskService != nil && task.SubDir != "" {
		e.mu.RLock()
		phase, ok := e.playlistPhases[task.PlaylistID]
		e.mu.RUnlock()
		if ok {
			e.taskService.UpdateTaskCurrentFile(phase.DownloadTaskID, filename, 0, 0)
		}
	}

	// 下载文件（支持断点续传）
	err = e.downloadFile(ctx, url, filePath, task)
	if err != nil {
		e.failTask(task, err.Error())
		e.checkPlaylistPhaseComplete(task.PlaylistID)
		return
	}

	// 下载完成，更新 DB
	fileInfo, _ := os.Stat(filePath)
	var fileSize int64
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}
	db.UpdateDownloadFilePath(task.SongID, filePath, fileSize)

	// 更新任务状态
	e.mu.Lock()
	task.Status = "completed"
	task.Progress = 100
	e.mu.Unlock()

	db.UpdateDownloadStatus(task.SongID, "completed")

	// 更新 TaskService 任务状态（单曲下载）
	if e.taskService != nil && task.TaskServiceID != "" {
		e.taskService.UpdateTaskProgress(task.TaskServiceID, 1, 1)
		e.taskService.CompleteTask(task.TaskServiceID)
	}

	// 检查歌单阶段
	e.checkPlaylistPhaseComplete(task.PlaylistID)
}

// downloadFile 下载文件（支持断点续传）
func (e *Engine) downloadFile(ctx context.Context, url, dstPath string, task *DownloadTask) error {
	partialPath := dstPath + ".partial"
	var downloadedSize int64 = 0

	// 优先使用 task 中恢复的下载进度（从 DB 恢复的断点续传）
	if task.DownloadedSize > 0 {
		downloadedSize = task.DownloadedSize
		fmt.Printf("[download] resume %s from DB: %d bytes\n", task.SongName, downloadedSize)
	}

	// 检查是否存在 .partial 文件（断点续传）
	if info, err := os.Stat(partialPath); err == nil {
		// 取两者中较大的值
		if info.Size() > downloadedSize {
			downloadedSize = info.Size()
			fmt.Printf("[download] resume %s from partial file: %d bytes\n", task.SongName, downloadedSize)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://music.163.com/")

	if downloadedSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", downloadedSize))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 服务器不支持断点续传，重新开始
	if resp.StatusCode == http.StatusOK && downloadedSize > 0 {
		downloadedSize = 0
		os.Remove(partialPath)
	}

	flag := os.O_CREATE | os.O_WRONLY
	if downloadedSize > 0 && resp.StatusCode == http.StatusPartialContent {
		flag |= os.O_APPEND
	} else {
		downloadedSize = 0
	}

	out, err := os.OpenFile(partialPath, flag, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	total := resp.ContentLength + downloadedSize
	buf := make([]byte, 64*1024)
	var lastPersist int64 = 0

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloadedSize += int64(n)

			// 更新内存进度
			e.mu.Lock()
			if total > 0 {
				task.Progress = float64(downloadedSize) / float64(total) * 100
			}
			task.DownloadedSize = downloadedSize
			task.TotalSize = total
			e.mu.Unlock()

			// 每 512KB 持久化一次到 DB
			if downloadedSize-lastPersist >= 512*1024 {
				db.UpdateDownloadProgress(task.SongID, downloadedSize, total)
				lastPersist = downloadedSize

				// 更新 TaskService 字节进度
				if e.taskService != nil {
					e.mu.RLock()
					phase, ok := e.playlistPhases[task.PlaylistID]
					e.mu.RUnlock()
					if ok {
						e.taskService.UpdateTaskCurrentFile(phase.DownloadTaskID,
							task.SongName, downloadedSize, total)
					}
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	out.Close()

	// 下载完成，重命名
	if err := os.Rename(partialPath, dstPath); err != nil {
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}

// executeMetadataTask 执行数据补全任务（第二阶段）
func (e *Engine) executeMetadataTask(ctx context.Context, task *DownloadTask) {
	if task.FilePath == "" {
		// 从 DB 读取文件路径
		history, _ := db.GetDownloadHistory()
		for _, h := range history {
			if h.SongID == task.SongID && h.FilePath != "" {
				task.FilePath = h.FilePath
				break
			}
		}
	}
	if task.FilePath == "" {
		fmt.Printf("[metadata] skip %s: no file path\n", task.SongName)
		e.completeMetadataTask(task)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(task.FilePath); err != nil {
		fmt.Printf("[metadata] skip %s: file not found\n", task.SongName)
		e.completeMetadataTask(task)
		return
	}

	song := &model.Song{
		ID:     task.SongID,
		Name:   task.SongName,
		Artist: task.Artist,
		Album:  task.Album,
	}

	coverEnabled := getSettingBool("data_complete_cover", true)
	lyricsEnabled := getSettingBool("data_complete_lyrics", true)
	artistEnabled := getSettingBool("data_complete_artist", true)

	// 从 DB 获取已完成的步骤（断点续传）
	history, _ := db.GetDownloadBySongID(task.SongID)
	coverDone := history != nil && history.CoverDownloaded
	lyricsDone := history != nil && history.LyricsDownloaded
	artistDone := history != nil && history.ArtistCompleted

	// 封面嵌入
	if coverEnabled && !coverDone {
		fmt.Printf("[metadata] embedding cover for %s\n", task.SongName)
		if e.taskService != nil {
			e.updateMetadataTaskProgress(task)
		}
		err := e.executeWithRetry("cover", func() error {
			_, err := e.metadataCompleter.DownloadAndEmbedCover(song, task.FilePath)
			return err
		}, task.SongName)
		if err != nil {
			fmt.Printf("[metadata] cover failed %s: %v\n", task.SongName, err)
		} else {
			db.UpdateMetadataProgress(task.SongID, "cover")
		}
	} else if coverEnabled && coverDone {
		fmt.Printf("[metadata] skip cover for %s (already done)\n", task.SongName)
	}

	// 歌词下载
	if lyricsEnabled && !lyricsDone {
		fmt.Printf("[metadata] downloading lyrics for %s\n", task.SongName)
		if e.taskService != nil {
			e.updateMetadataTaskProgress(task)
		}
		err := e.executeWithRetry("lyrics", func() error {
			return e.metadataCompleter.DownloadLyrics(song, task.FilePath)
		}, task.SongName)
		if err != nil {
			fmt.Printf("[metadata] lyrics failed %s: %v\n", task.SongName, err)
		} else {
			db.UpdateMetadataProgress(task.SongID, "lyrics")
		}
	} else if lyricsEnabled && lyricsDone {
		fmt.Printf("[metadata] skip lyrics for %s (already done)\n", task.SongName)
	}

	// 艺人信息嵌入（如果封面已嵌入则艺人信息也已写入）
	if artistEnabled && !coverEnabled && !artistDone {
		fmt.Printf("[metadata] embedding artist info for %s\n", task.SongName)
		if e.taskService != nil {
			e.updateMetadataTaskProgress(task)
		}
		err := e.executeWithRetry("artist", func() error {
			return e.metadataCompleter.EmbedArtistInfo(song, task.FilePath)
		}, task.SongName)
		if err != nil {
			fmt.Printf("[metadata] artist failed %s: %v\n", task.SongName, err)
		} else {
			db.UpdateMetadataProgress(task.SongID, "artist")
		}
	} else if artistEnabled && !coverEnabled && artistDone {
		fmt.Printf("[metadata] skip artist for %s (already done)\n", task.SongName)
	}

	db.UpdateMetadataProgress(task.SongID, "completed")
	e.completeMetadataTask(task)
}

func (e *Engine) updateMetadataTaskProgress(task *DownloadTask) {
	e.mu.RLock()
	phase, ok := e.playlistPhases[task.PlaylistID]
	e.mu.RUnlock()
	if ok {
		e.taskService.UpdateTaskCurrentFile(phase.MetadataTaskID, task.SongName, 0, 0)
	}
}

func (e *Engine) completeMetadataTask(task *DownloadTask) {
	e.mu.RLock()
	phase, ok := e.playlistPhases[task.PlaylistID]
	e.mu.RUnlock()

	if ok {
		phase.mu.Lock()
		phase.MetadataDone++
		current := phase.MetadataDone
		total := phase.TotalSongs
		phase.mu.Unlock()

		if e.taskService != nil {
			e.taskService.UpdateTaskProgress(phase.MetadataTaskID, current, total)
		}

		if current >= total {
			if e.taskService != nil {
				e.taskService.CompleteTask(phase.MetadataTaskID)
			}
			phase.mu.Lock()
			phase.Phase = "completed"
			phase.mu.Unlock()
		}
	}
}

// checkPlaylistPhaseComplete 检查歌单阶段是否完成（包括失败的情况）
func (e *Engine) checkPlaylistPhaseComplete(playlistID int) {
	if playlistID == 0 {
		return
	}

	e.mu.RLock()
	phase, ok := e.playlistPhases[playlistID]
	e.mu.RUnlock()

	if !ok {
		return
	}

	phase.mu.Lock()
	defer phase.mu.Unlock()

	// 统计该歌单的所有任务状态
	e.mu.RLock()
	var downloadDone, downloadFailed, downloadCancelled, metadataDone, metadataFailed int
	for _, t := range e.tasks {
		if t.PlaylistID == playlistID {
			if t.Status == "completed" {
				if t.Phase == "download" || t.Phase == "" {
					downloadDone++
				} else if t.Phase == "metadata" {
					metadataDone++
				}
			} else if t.Status == "failed" {
				if t.Phase == "download" || t.Phase == "" {
					downloadFailed++
				} else if t.Phase == "metadata" {
					metadataFailed++
				}
			} else if t.Status == "cancelled" {
				if t.Phase == "download" || t.Phase == "" {
					downloadCancelled++
				}
			}
		}
	}
	e.mu.RUnlock()

	total := phase.TotalSongs

	// 下载阶段完成（包括成功、失败和取消）
	if phase.Phase == "downloading" && (downloadDone+downloadFailed+downloadCancelled) >= total {
		phase.DownloadDone = downloadDone
		if e.taskService != nil {
			e.taskService.UpdateTaskProgress(phase.DownloadTaskID, downloadDone, total)
			if downloadDone > 0 {
				e.taskService.CompleteTask(phase.DownloadTaskID)
			} else {
				e.taskService.FailTask(phase.DownloadTaskID, "所有歌曲下载失败")
			}
		}

		// 检查是否需要数据补全（跟随同步运行，非独立定时）
		autoComplete := getSettingBool("auto_data_complete", true)
		if autoComplete && downloadDone > 0 && phase.MetadataTaskID != "" {
			// 使用预创建的 metadataTask，更新 Total 为实际下载成功数
			e.taskService.UpdateTaskProgress(phase.MetadataTaskID, 0, downloadDone)
			e.taskService.SetTaskStatus(phase.MetadataTaskID, service.TaskStatusRunning)
			phase.Phase = "metadata"

			// 将所有成功下载的歌曲放入元数据队列
			e.mu.RLock()
			for _, t := range e.tasks {
				if t.PlaylistID == playlistID && t.Status == "completed" {
					t.Phase = "metadata"
					e.metadataQueue <- t
				}
			}
			e.mu.RUnlock()
		} else {
			phase.Phase = "completed"
		}
	}

	// 元数据阶段完成（包括成功和失败）
	if phase.Phase == "metadata" && (metadataDone+metadataFailed) >= phase.DownloadDone {
		phase.MetadataDone = metadataDone
		if e.taskService != nil {
			e.taskService.UpdateTaskProgress(phase.MetadataTaskID, metadataDone, phase.DownloadDone)
			if metadataDone > 0 {
				e.taskService.CompleteTask(phase.MetadataTaskID)
			} else {
				e.taskService.FailTask(phase.MetadataTaskID, "所有元数据补全失败")
			}
		}
		phase.Phase = "completed"
	}
}

// VerifyMetadata 验证并补全歌单元数据
func (e *Engine) VerifyMetadata(playlistID int) (string, error) {
	fmt.Printf("[VerifyMetadata] starting for playlistID=%d\n", playlistID)

	// 获取歌单信息
	cookie, _ := db.GetCookie()
	body, err := e.netease.GetPlaylistDetail(playlistID, cookie)
	if err != nil {
		return "", fmt.Errorf("获取歌单详情失败: %w", err)
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	playlist, ok := result["playlist"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("解析歌单失败")
	}

	playlistName := ""
	if name, ok := playlist["name"].(string); ok {
		playlistName = sanitizeFilename(name)
	}

	// 获取该歌单所有已下载但未完成元数据的歌曲
	history, err := db.GetDownloadHistory()
	if err != nil {
		return "", fmt.Errorf("获取下载历史失败: %w", err)
	}

	var incompleteSongs []model.DownloadHistory
	for _, h := range history {
		if h.PlaylistID == playlistID && h.Phase == "download" && h.Status == "completed" && !h.MetadataCompleted {
			incompleteSongs = append(incompleteSongs, h)
		}
	}

	if len(incompleteSongs) == 0 {
		return "", fmt.Errorf("该歌单所有歌曲元数据已补全")
	}

	fmt.Printf("[VerifyMetadata] found %d songs with incomplete metadata\n", len(incompleteSongs))

	// 创建验证补全任务
	verifyTask := e.taskService.CreateTask(service.TaskTypeDataComplete,
		fmt.Sprintf("验证补全歌单「%s」元数据", playlistName), "")
	verifyTask.Total = len(incompleteSongs)

	// 创建歌单阶段追踪
	phase := &PlaylistPhase{
		PlaylistID:     playlistID,
		PlaylistName:   playlistName,
		MetadataTaskID: verifyTask.ID,
		TotalSongs:     len(incompleteSongs),
		DownloadDone:   len(incompleteSongs),
		Phase:          "metadata",
	}
	e.mu.Lock()
	e.playlistPhases[playlistID] = phase
	e.mu.Unlock()

	// 将未完成的歌曲加入元数据队列
	e.taskService.SetTaskStatus(verifyTask.ID, service.TaskStatusRunning)
	for _, h := range incompleteSongs {
		task := &DownloadTask{
			SongID:     h.SongID,
			SongName:   h.SongName,
			Artist:     h.Artist,
			Album:      h.Album,
			FilePath:   h.FilePath,
			SubDir:     h.SubDir,
			PlaylistID: playlistID,
			Status:     "pending",
			Phase:      "metadata",
		}
		e.mu.Lock()
		task.ID = len(e.tasks) + 1
		e.tasks[task.ID] = task
		e.mu.Unlock()

		e.metadataQueue <- task
		fmt.Printf("[VerifyMetadata] queued metadata task: %s (songID=%d)\n", h.SongName, h.SongID)
	}

	return verifyTask.ID, nil
}

func (e *Engine) failTask(task *DownloadTask, errMsg string) {
	e.mu.Lock()
	task.Status = "failed"
	task.Error = errMsg
	e.mu.Unlock()

	db.UpdateDownloadStatus(task.SongID, "failed")
}

// GetTask 获取任务状态
func (e *Engine) GetTask(id int) *DownloadTask {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.tasks[id]
}

// GetAllTasks 获取所有任务
func (e *Engine) GetAllTasks() []*DownloadTask {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var tasks []*DownloadTask
	for _, t := range e.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// GetTaskService 获取 TaskService（供 API 层使用）
func (e *Engine) GetTaskService() *service.TaskService {
	return e.taskService
}

// settingsWatcher 监控设置变化，执行定时任务
func (e *Engine) settingsWatcher(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// 从数据库恢复上次同步时间
	lastSyncTimeStr, _ := db.GetSetting("last_sync_time")
	var lastSyncTime time.Time
	if lastSyncTimeStr != "" {
		lastSyncTime, _ = time.Parse(time.RFC3339, lastSyncTimeStr)
	}

	// 启动时立即检查一次
	if getSettingBool("auto_sync", false) {
		interval := getSyncInterval("sync_interval", "sync_unit", 12)
		if time.Since(lastSyncTime) > interval {
			fmt.Println("[engine] auto sync triggered on startup")
			lastSyncTime = time.Now()
			db.SetSetting("last_sync_time", lastSyncTime.Format(time.RFC3339))
			nextTime := lastSyncTime.Add(interval)
			db.SetSetting("next_sync_time", nextTime.Format(time.RFC3339))
			go e.runAutoSync(ctx)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 自动同步（数据补全跟随同步运行，在 checkPlaylistPhaseComplete 中触发）
			if getSettingBool("auto_sync", false) {
				interval := getSyncInterval("sync_interval", "sync_unit", 12)
				if time.Since(lastSyncTime) > interval {
					fmt.Println("[engine] auto sync triggered")
					lastSyncTime = time.Now()
					// 保存同步时间到数据库
					db.SetSetting("last_sync_time", lastSyncTime.Format(time.RFC3339))
					nextTime := lastSyncTime.Add(interval)
					db.SetSetting("next_sync_time", nextTime.Format(time.RFC3339))
					// 执行自动同步
					e.runAutoSync(ctx)
				}
			}
		}
	}
}

// runAutoSync 执行自动同步：获取用户歌单列表，逐个触发下载
func (e *Engine) runAutoSync(ctx context.Context) {
	fmt.Println("[autoSync] starting auto sync...")

	// 获取当前用户
	user, err := db.GetCurrentUser()
	if err != nil || user == nil {
		fmt.Printf("[autoSync] no user logged in: %v\n", err)
		return
	}

	// 获取用户歌单列表
	body, err := e.netease.GetUserPlaylists(user.UserID)
	if err != nil {
		fmt.Printf("[autoSync] failed to get playlists: %v\n", err)
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	playlistData, ok := result["playlist"].([]interface{})
	if !ok {
		fmt.Println("[autoSync] no playlists found")
		return
	}

	quality, _ := db.GetSetting("quality")
	if quality == "" {
		quality = "high"
	}

	// 创建同步任务日志
	syncTask := e.taskService.CreateTask(service.TaskTypeSync, "自动同步歌单", "")
	e.taskService.SetTaskStatus(syncTask.ID, service.TaskStatusRunning)

	totalPlaylists := 0
	syncedPlaylists := 0

	for _, p := range playlistData {
		playlist := p.(map[string]interface{})
		playlistID := int(playlist["id"].(float64))
		playlistName := ""
		if name, ok := playlist["name"].(string); ok {
			playlistName = name
		}

		fmt.Printf("[autoSync] syncing playlist: %s (id=%d)\n", playlistName, playlistID)

		// 检查是否已有活跃任务
		e.mu.RLock()
		phase, exists := e.playlistPhases[playlistID]
		e.mu.RUnlock()
		if exists {
			phase.mu.Lock()
			isActive := phase.Phase == "scanning" || phase.Phase == "downloading" || phase.Phase == "metadata"
			phase.mu.Unlock()
			if isActive {
				fmt.Printf("[autoSync] skipping playlist %s (active task)\n", playlistName)
				totalPlaylists++
				continue
			}
		}

		totalPlaylists++

		// 触发歌单下载（会自动创建扫描→下载→补全任务）
		_, _, _, err := e.AddPlaylistTask(playlistID, quality)
		if err != nil {
			fmt.Printf("[autoSync] failed to sync playlist %s: %v\n", playlistName, err)
			continue
		}
		syncedPlaylists++
	}

	e.taskService.UpdateTaskProgress(syncTask.ID, syncedPlaylists, totalPlaylists)
	e.taskService.CompleteTask(syncTask.ID)

	fmt.Printf("[autoSync] completed: %d/%d playlists synced\n", syncedPlaylists, totalPlaylists)
}

// autoCompleteMetadata 自动补全未完成元数据的歌曲
func (e *Engine) autoCompleteMetadata() {
	pending, err := db.GetPendingDownloads()
	if err != nil {
		return
	}

	for _, d := range pending {
		if d.Phase == "metadata" && d.FilePath != "" {
			task := &DownloadTask{
				SongID:   d.SongID,
				SongName: d.SongName,
				Artist:   d.Artist,
				Album:    d.Album,
				FilePath: d.FilePath,
				SubDir:   d.SubDir,
			}
			e.metadataQueue <- task
		}
	}
}

func getSettingBool(key string, defaultVal bool) bool {
	val, _ := db.GetSetting(key)
	if val == "" {
		return defaultVal
	}
	return val == "true"
}

func getSyncInterval(intervalKey, unitKey string, defaultHours int) time.Duration {
	intervalStr, _ := db.GetSetting(intervalKey)
	unitStr, _ := db.GetSetting(unitKey)

	hours := defaultHours
	if intervalStr != "" {
		if v, err := fmt.Sscanf(intervalStr, "%d", &hours); v == 0 || err != nil {
			hours = defaultHours
		}
	}

	if unitStr == "day" {
		return time.Duration(hours) * 24 * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

// sanitizeFilename 清理文件名中的特殊字符，防止乱码和路径注入
func sanitizeFilename(name string) string {
	reg := regexp.MustCompile(`[\x00-\x1f\x7f\x80-\x9f\x{200B}\x{200C}\x{200D}\x{FEFF}]`)
	name = reg.ReplaceAllString(name, "")

	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")

	reg2 := regexp.MustCompile(`[<>:"|?*]`)
	name = reg2.ReplaceAllString(name, "")

	name = strings.Trim(name, " .")

	runes := []rune(name)
	if len(runes) > 200 {
		name = string(runes[:200])
	}

	if name == "" {
		name = "unknown"
	}

	return name
}

// executeWithRetry 带频率限制检测的重试执行
func (e *Engine) executeWithRetry(stepName string, fn func() error, songName string) error {
	maxRetries := 5
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		if isRateLimitError(err) {
			waitTime := time.Duration(30*(attempt+1)) * time.Second
			fmt.Printf("[metadata] %s rate limited for %s, waiting %v (attempt %d/%d)\n",
				stepName, songName, waitTime, attempt+1, maxRetries)
			time.Sleep(waitTime)
			continue
		}
		return err
	}
	return fmt.Errorf("%s: 超过最大重试次数", stepName)
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "频率") ||
		strings.Contains(errMsg, "rate limit") ||
		strings.Contains(errMsg, "429") ||
		strings.Contains(errMsg, "频繁") ||
		strings.Contains(errMsg, "too many")
}
