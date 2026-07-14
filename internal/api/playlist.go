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
	user, err := db.GetCurrentUser()
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.GetUserPlaylists(user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	var playlists []map[string]interface{}
	if playlistData, ok := result["playlist"].([]interface{}); ok {
		for _, p := range playlistData {
			playlist := p.(map[string]interface{})
			pl := map[string]interface{}{
				"id":          playlist["id"],
				"name":        playlist["name"],
				"track_count": playlist["trackCount"],
				"cover":       playlist["coverImgUrl"],
			}
			playlists = append(playlists, pl)

			// 缓存到数据库
			db.SavePlaylist(&model.Playlist{
				PlaylistID: int(playlist["id"].(float64)),
				Name:       playlist["name"].(string),
				CreatorID:  user.UserID,
				TrackCount: int(playlist["trackCount"].(float64)),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
}

func getPlaylistDetail(c *gin.Context) {
	playlistIDStr := c.Query("id")
	playlistID, err := strconv.Atoi(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "歌单ID无效"})
		return
	}

	cookie, _ := db.GetCookie()
	fmt.Printf("[playlist] cookie length: %d, playlistID: %d\n", len(cookie), playlistID)
	netease := service.NewNeteaseService("http://127.0.0.1:3000")
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

	fmt.Printf("[playlist] API response code: %v\n", result["code"])

	var tracks []map[string]interface{}
	if playlist, ok := result["playlist"].(map[string]interface{}); ok {
		// 检查 tracks 数组
		var trackList []interface{}
		if tl, ok := playlist["tracks"].([]interface{}); ok {
			trackList = tl
			fmt.Printf("[playlist] tracks array length: %d\n", len(trackList))
		} else {
			fmt.Printf("[playlist] tracks field type: %T, value: %v\n", playlist["tracks"], playlist["tracks"])
		}

		// 检查 trackIds
		if tids, ok := playlist["trackIds"].([]interface{}); ok {
			fmt.Printf("[playlist] trackIds array length: %d\n", len(tids))
		} else {
			fmt.Printf("[playlist] trackIds field type: %T, value: %v\n", playlist["trackIds"], playlist["trackIds"])
		}

		for _, t := range trackList {
			track, ok := t.(map[string]interface{})
			if !ok {
				continue
			}

			// 从 tracks 提取基础信息
			var name string
			if v, exists := track["name"]; exists && v != nil {
				name, _ = v.(string)
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
			album := ""
			if al, ok := track["al"].(map[string]interface{}); ok {
				album, _ = al["name"].(string)
			}
			duration := 0
			if dt, ok := track["dt"].(float64); ok {
				duration = int(dt / 1000)
			}

			// 对 null 字段，回退到批量 API 补充
			needDetail := name == "" || artists == "" || album == ""
			if needDetail {
				if id, ok := track["id"]; ok {
					detailBody, err := netease.GetSongDetailBatch(fmt.Sprintf("%v", id))
					if err == nil {
						var detailResult map[string]interface{}
						json.Unmarshal(detailBody, &detailResult)
						if songs, ok := detailResult["songs"].([]interface{}); ok && len(songs) > 0 {
							song := songs[0].(map[string]interface{})
							if name == "" {
								if v, exists := song["name"]; exists && v != nil {
									name, _ = v.(string)
								}
								if name == "" {
									if tn, ok := song["tns"].([]interface{}); ok && len(tn) > 0 {
										if str, ok := tn[0].(string); ok {
											name = str
										}
									}
								}
								if name == "" {
									if alia, ok := song["alia"].([]interface{}); ok && len(alia) > 0 {
										if str, ok := alia[0].(string); ok {
											name = str
										}
									}
								}
							}
							if artists == "" {
								if ar, ok := song["ar"].([]interface{}); ok {
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
							}
							if album == "" {
								if al, ok := song["al"].(map[string]interface{}); ok {
									album, _ = al["name"].(string)
								}
							}
							if duration == 0 {
								if dt, ok := song["dt"].(float64); ok {
									duration = int(dt / 1000)
								}
							}
						}
					}
				}
			}

			if name == "" {
				name = "未知歌曲"
			}
			if artists == "" {
				artists = "未知歌手"
			}
			if album == "" {
				album = "未知专辑"
			}

			if name == "" || strings.TrimSpace(name) == "" || name == "未知歌曲" {
				fmt.Printf("[playlist] skipping empty track id=%v, name='%s'\n", track["id"], name)
				continue
			}

			tracks = append(tracks, map[string]interface{}{
				"id":       track["id"],
				"name":     name,
				"artist":   artists,
				"album":    album,
				"duration": duration,
			})
		}

		// 如果 tracks 为空，回退到 trackIds 批量获取
		if len(tracks) == 0 {
			var ids []string
			if trackIds, ok := playlist["trackIds"].([]interface{}); ok {
				for _, tid := range trackIds {
					switch v := tid.(type) {
					case float64:
						ids = append(ids, fmt.Sprintf("%.0f", v))
					case map[string]interface{}:
						if id, ok := v["id"]; ok {
							ids = append(ids, fmt.Sprintf("%v", id))
						}
					}
				}
			}
			if len(ids) > 0 {
				for start := 0; start < len(ids); start += 1000 {
					end := start + 1000
					if end > len(ids) {
						end = len(ids)
					}
					batchIds := strings.Join(ids[start:end], ",")
					detailBody, err := netease.GetSongDetailBatch(batchIds)
					if err != nil {
						continue
					}
					var detailResult map[string]interface{}
					json.Unmarshal(detailBody, &detailResult)
					if songs, ok := detailResult["songs"].([]interface{}); ok {
						for _, s := range songs {
							song := s.(map[string]interface{})
							name, _ := song["name"].(string)
							if name == "" {
								if tn, ok := song["tns"].([]interface{}); ok && len(tn) > 0 {
									if str, ok := tn[0].(string); ok {
										name = str
									}
								}
							}
							if name == "" {
								if alia, ok := song["alia"].([]interface{}); ok && len(alia) > 0 {
									if str, ok := alia[0].(string); ok {
										name = str
									}
								}
							}
							if name == "" {
								name = "未知歌曲"
							}

							artists := ""
							if ar, ok := song["ar"].([]interface{}); ok {
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
							if al, ok := song["al"].(map[string]interface{}); ok {
								album, _ = al["name"].(string)
							}
							if album == "" {
								album = "未知专辑"
							}

							duration := 0
							if dt, ok := song["dt"].(float64); ok {
								duration = int(dt / 1000)
							}

							if name == "" || strings.TrimSpace(name) == "" || name == "未知歌曲" {
								continue
							}

							tracks = append(tracks, map[string]interface{}{
								"id":       song["id"],
								"name":     name,
								"artist":   artists,
								"album":    album,
								"duration": duration,
							})
						}
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"playlist": result["playlist"],
		"tracks":   tracks,
	})
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
