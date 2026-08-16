package api

import (
	"net/http"

	"endfield-music/internal/db"

	"github.com/gin-gonic/gin"
)

// GetSyncPlaylists 获取同步歌单列表
func GetSyncPlaylists(c *gin.Context) {
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	playlists, err := db.GetSyncPlaylists(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取同步歌单失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    playlists,
	})
}

// UpdateSyncPlaylists 更新同步歌单列表
func UpdateSyncPlaylists(c *gin.Context) {
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		PlaylistIDs []int `json:"playlist_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := db.BatchUpdateSyncPlaylists(userID, req.PlaylistIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新同步歌单失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "同步歌单已更新",
	})
}

// ToggleSyncPlaylist 切换歌单同步状态
func ToggleSyncPlaylist(c *gin.Context) {
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		PlaylistID int  `json:"playlist_id"`
		Enabled    bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := db.ToggleSyncPlaylist(req.PlaylistID, userID, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "切换同步状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "同步状态已更新",
	})
}

// getCurrentUserID 获取当前登录用户ID
func getCurrentUserID(c *gin.Context) int {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int); ok {
			return id
		}
	}
	return 0
}
