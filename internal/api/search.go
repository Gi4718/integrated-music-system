package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"endfield-music/internal/service"

	"github.com/gin-gonic/gin"
)

func searchSongs(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "关键词不能为空"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.SearchSongs(keyword, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	songs := parseSearchResults(result)

	var total interface{}
	if resultData, ok := result["result"].(map[string]interface{}); ok {
		total = resultData["songCount"]
	}

	c.JSON(http.StatusOK, gin.H{
		"songs": songs,
		"total": total,
	})
}

func parseSearchResults(result map[string]interface{}) []map[string]interface{} {
	var songs []map[string]interface{}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		if songsData, ok := resultData["songs"].([]interface{}); ok {
			for _, s := range songsData {
				song := s.(map[string]interface{})
				artist := ""
				if ar, ok := song["artists"].([]interface{}); ok && len(ar) > 0 {
					artist = ar[0].(map[string]interface{})["name"].(string)
				}
				album := ""
				if al, ok := song["album"].(map[string]interface{}); ok {
					album = al["name"].(string)
				}

				songs = append(songs, map[string]interface{}{
					"id":       song["id"],
					"name":     song["name"],
					"artist":   artist,
					"album":    album,
					"duration": song["duration"],
					"pic_url":  song["album"].(map[string]interface{})["picUrl"],
				})
			}
		}
	}

	return songs
}
