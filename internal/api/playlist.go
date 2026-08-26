package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"endfield-music/internal/db"
	"endfield-music/internal/model"
	"endfield-music/internal/service"

	"github.com/gin-gonic/gin"
)

func getUserPlaylists(c *gin.Context) {
	systemUserID := getSystemUserID(c)
	user, err := db.GetCurrentUserForSystem(systemUserID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.GetUserPlaylists(user.UserID, user.Cookie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	var playlists []map[string]interface{}
	if playlistData, ok := result["playlist"].([]interface{}); ok {
		for _, p := range playlistData {
			if p == nil {
				continue
			}
			playlist, ok := p.(map[string]interface{})
			if !ok || playlist == nil {
				continue
			}
			
			// 安全获取字段值
			id, _ := playlist["id"].(float64)
			name, _ := playlist["name"].(string)
			trackCount, _ := playlist["trackCount"].(float64)
			cover, _ := playlist["coverImgUrl"].(string)
			
			// 跳过无效数据
			if id == 0 || name == "" {
				continue
			}
			
			// 标记是否可写（用户自己创建的歌单）
			creatorID := float64(0)
			if creator, ok := playlist["creator"].(map[string]interface{}); ok {
				if uid, ok := creator["userId"].(float64); ok {
					creatorID = uid
				}
			}
			isWritable := creatorID == float64(user.UserID)

			pl := map[string]interface{}{
				"id":          id,
				"name":        name,
				"track_count": trackCount,
				"cover":       cover,
				"writable":    isWritable,
			}
			playlists = append(playlists, pl)

			// 缓存到数据库
			db.SavePlaylist(&model.Playlist{
				PlaylistID: int(id),
				Name:       name,
				CreatorID:  user.UserID,
				TrackCount: int(trackCount),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": playlists})
}

// addSongToPlaylist 添加歌曲到歌单
func addSongToPlaylist(c *gin.Context) {
	systemUserID := getSystemUserID(c)
	user, err := db.GetCurrentUserForSystem(systemUserID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		PlaylistID int `json:"playlist_id" binding:"required"`
		SongID     int `json:"song_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	cookie, _ := db.GetCookie()
	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.AddSongToPlaylist(req.PlaylistID, req.SongID, cookie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// 检查网易云 API 返回
	if code, ok := result["code"].(float64); ok {
		if code == 200 {
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "添加成功"})
			return
		}
		// 解析错误信息
		message, _ := result["message"].(string)
		if message == "" {
			message = "未知错误"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "添加成功"})
}

func getPlaylistDetail(c *gin.Context) {
	playlistIDStr := c.Query("id")
	playlistID, err := strconv.Atoi(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "歌单ID无效"})
		return
	}

	// 分页参数
	offset := 0
	limit := 100
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}

	cookie, _ := db.GetCookie()
	netease := service.NewNeteaseService("http://127.0.0.1:3000")

	// 获取歌单详情（含 trackIds）
	body, err := netease.GetPlaylistDetail(playlistID, cookie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("解析响应失败: %v", err)})
		return
	}

	// 提取 trackIds
	var allTrackIDs []int
	total := 0
	if playlist, ok := result["playlist"].(map[string]interface{}); ok {
		if tc, ok := playlist["trackCount"].(float64); ok {
			total = int(tc)
		}
		if trackIDs, ok := playlist["trackIds"].([]interface{}); ok {
			for _, id := range trackIDs {
				switch v := id.(type) {
				case float64:
					allTrackIDs = append(allTrackIDs, int(v))
				case map[string]interface{}:
					if songID, ok := v["id"].(float64); ok {
						allTrackIDs = append(allTrackIDs, int(songID))
					}
				}
			}
		}
	}

	// 用 trackIds 的实际数量作为 total（比 trackCount 更准确）
	if len(allTrackIDs) > 0 {
		total = len(allTrackIDs)
	}

	// 分页截取 trackIds
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	tracks := make([]map[string]interface{}, 0)
	if len(allTrackIDs) > 0 && start < end {
		pageIDs := allTrackIDs[start:end]
		// 分批获取歌曲详情（每批最多100首）
		for i := 0; i < len(pageIDs); i += 100 {
			batchEnd := i + 100
			if batchEnd > len(pageIDs) {
				batchEnd = len(pageIDs)
			}
			batchIDs := pageIDs[i:batchEnd]
			idStrs := make([]string, len(batchIDs))
			for j, id := range batchIDs {
				idStrs[j] = strconv.Itoa(id)
			}
			idsParam := strings.Join(idStrs, ",")

			songBody, err := netease.GetSongDetailBatch(idsParam)
			if err != nil {
				continue
			}
			var songResult map[string]interface{}
			if err := json.Unmarshal(songBody, &songResult); err != nil {
				continue
			}
			if songs, ok := songResult["songs"].([]interface{}); ok {
				for _, s := range songs {
					tracks = append(tracks, parseTrack(s.(map[string]interface{})))
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"playlist": result["playlist"],
		"tracks":   tracks,
		"total":    total,
		"offset":   offset,
		"limit":    limit,
	})
}

func parseTrack(track map[string]interface{}) map[string]interface{} {
	name, _ := track["name"].(string)
	if name == "" {
		if tn, ok := track["tns"].([]interface{}); ok && len(tn) > 0 {
			name, _ = tn[0].(string)
		}
	}
	if name == "" {
		if alia, ok := track["alia"].([]interface{}); ok && len(alia) > 0 {
			name, _ = alia[0].(string)
		}
	}
	if name == "" {
		name = "未知歌曲"
	}

	artists := ""
	if ar, ok := track["ar"].([]interface{}); ok {
		names := make([]string, 0)
		for _, a := range ar {
			if artist, ok := a.(map[string]interface{}); ok {
				if n, ok := artist["name"].(string); ok && n != "" {
					names = append(names, n)
				}
			}
		}
		artists = strings.Join(names, "/")
	}
	if artists == "" {
		artists = "未知歌手"
	}

	album := ""
	picURL := ""
	if al, ok := track["al"].(map[string]interface{}); ok {
		album, _ = al["name"].(string)
		picURL, _ = al["picUrl"].(string)
	}
	if album == "" {
		album = "未知专辑"
	}

	duration := 0
	if dt, ok := track["dt"].(float64); ok {
		duration = int(dt / 1000)
	}

	return map[string]interface{}{
		"id":       track["id"],
		"name":     name,
		"artist":   artists,
		"album":    album,
		"pic_url":  picURL,
		"duration": duration,
	}
}

func subscribePlaylist(c *gin.Context) {
	var req struct {
		PlaylistID int `json:"playlist_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	cookie, _ := db.GetCookie()
	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.SubscribePlaylist(req.PlaylistID, cookie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("解析响应失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, result)
}
