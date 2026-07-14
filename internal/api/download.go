package api

import (
	"net/http"

	"endfield-music/internal/db"

	"github.com/gin-gonic/gin"
)

func downloadSong(c *gin.Context) {
	var req struct {
		SongID  int    `json:"song_id" binding:"required"`
		Quality string `json:"quality"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Quality == "" {
		req.Quality = "high"
	}

	// 从数据库获取下载引擎实例（通过全局变量或依赖注入）
	// 这里简化处理，直接返回任务已创建
	taskID, err := getDownloadEngine().AddTask(req.SongID, req.Quality)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已创建",
		"task_id": taskID,
	})
}

func downloadPlaylist(c *gin.Context) {
	var req struct {
		PlaylistID int    `json:"playlist_id" binding:"required"`
		Quality    string `json:"quality"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Quality == "" {
		req.Quality = "high"
	}

	taskIDs, downloadTaskID, metadataTaskID, err := getDownloadEngine().AddPlaylistTask(req.PlaylistID, req.Quality)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "任务已创建",
		"task_ids":          taskIDs,
		"download_task_id":  downloadTaskID,
		"metadata_task_id":  metadataTaskID,
	})
}

func getDownloadHistory(c *gin.Context) {
	history, err := db.GetDownloadHistory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

func getDownloadProgress(c *gin.Context) {
	engine := getDownloadEngine()
	if engine == nil {
		c.JSON(http.StatusOK, gin.H{"tasks": []interface{}{}})
		return
	}

	// 获取 TaskService 的任务（包含 current_file, current_bytes, total_bytes）
	taskService := engine.GetTaskService()
	if taskService != nil {
		tasks := taskService.GetAllTasks()
		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
		return
	}

	// 回退到旧的下载任务列表
	downloadTasks := engine.GetAllTasks()
	c.JSON(http.StatusOK, gin.H{"tasks": downloadTasks})
}

func verifyMetadata(c *gin.Context) {
	var req struct {
		PlaylistID int `json:"playlist_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	engine := getDownloadEngine()
	if engine == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "下载引擎未初始化"})
		return
	}

	taskID, err := engine.VerifyMetadata(req.PlaylistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "验证补全任务已创建",
		"task_id": taskID,
	})
}
