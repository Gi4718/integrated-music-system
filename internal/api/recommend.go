package api

import (
	"encoding/json"
	"net/http"

	"endfield-music/internal/db"
	"endfield-music/internal/service"

	"github.com/gin-gonic/gin"
)

func getRecommendSongs(c *gin.Context) {
	cookie, _ := db.GetCookie()
	netease := service.NewNeteaseService("http://127.0.0.1:3000")

	body, err := netease.GetRecommendSongs(cookie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "songs": []interface{}{}})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusOK, gin.H{"songs": []interface{}{}})
		return
	}

	var songs []map[string]interface{}

	if data, ok := result["data"].(map[string]interface{}); ok {
		if list, ok := data["dailySongs"].([]interface{}); ok {
			for _, s := range list {
				song := parseRecommendSong(s)
				if song != nil {
					songs = append(songs, song)
				}
				if len(songs) >= 10 {
					break
				}
			}
		}
	}

	if songs == nil {
		songs = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{"songs": songs})
}

func getRecommendPlaylists(c *gin.Context) {
	cookie, _ := db.GetCookie()
	netease := service.NewNeteaseService("http://127.0.0.1:3000")

	body, err := netease.GetRecommendPlaylists(cookie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "playlists": []interface{}{}})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusOK, gin.H{"playlists": []interface{}{}})
		return
	}

	var playlists []map[string]interface{}

	if recommend, ok := result["result"].([]interface{}); ok {
		for _, p := range recommend {
			if pm, ok := p.(map[string]interface{}); ok {
				playlists = append(playlists, map[string]interface{}{
					"id":         pm["id"],
					"name":       pm["name"],
					"cover":      pm["picUrl"],
					"trackCount": pm["trackCount"],
				})
			}
		}
	}

	if playlists == nil {
		playlists = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
}

func parseRecommendSong(s interface{}) map[string]interface{} {
	song, ok := s.(map[string]interface{})
	if !ok {
		return nil
	}

	name, _ := song["name"].(string)
	if name == "" {
		return nil
	}

	artist := ""
	if ar, ok := song["ar"].([]interface{}); ok && len(ar) > 0 {
		if ar0, ok := ar[0].(map[string]interface{}); ok {
			artist, _ = ar0["name"].(string)
		}
	}

	album := ""
	var picURL string
	if al, ok := song["al"].(map[string]interface{}); ok {
		album, _ = al["name"].(string)
		picURL, _ = al["picUrl"].(string)
	}

	duration := 0
	if d, ok := song["dt"].(float64); ok {
		duration = int(d) / 1000
	}

	return map[string]interface{}{
		"id":       song["id"],
		"name":     name,
		"artist":   artist,
		"album":    album,
		"duration": duration,
		"pic_url":  picURL,
	}
}
