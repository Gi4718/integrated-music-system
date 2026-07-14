package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"endfield-music/internal/db"
	"endfield-music/internal/service"

	"github.com/gin-gonic/gin"
)

func checkSongURL(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "歌曲 ID 无效"})
		return
	}

	cookie, _ := db.GetCookie()
	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.GetSongURL(id, 320000, cookie)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"available": false, "reason": err.Error()})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		c.JSON(http.StatusOK, gin.H{"available": false, "reason": "未找到音频"})
		return
	}

	songData := data[0].(map[string]interface{})
	url, _ := songData["url"].(string)
	if url == "" {
		code, _ := songData["code"].(float64)
		reason := "无法获取播放链接"
		if code == -110 {
			reason = "需要 VIP 或版权限制"
		} else if code == -100 {
			reason = "歌曲不存在"
		} else if code == 403 {
			reason = "版权限制"
		}
		c.JSON(http.StatusOK, gin.H{"available": false, "reason": reason})
		return
	}

	c.JSON(http.StatusOK, gin.H{"available": true})
}

func streamAudio(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "歌曲 ID 无效"})
		return
	}

	cookie, _ := db.GetCookie()
	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.GetSongURL(id, 320000, cookie)
	if err != nil {
		fmt.Printf("[stream] GetSongURL error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("[stream] JSON parse error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析响应失败"})
		return
	}

	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		fmt.Printf("[stream] No data in response\n")
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到音频"})
		return
	}

	songData := data[0].(map[string]interface{})
	url, _ := songData["url"].(string)
	if url == "" {
		code, _ := songData["code"].(float64)
		fmt.Printf("[stream] Empty URL, code=%.0f\n", code)
		if code == -110 {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要 VIP 或版权限制"})
		} else if code == -100 {
			c.JSON(http.StatusNotFound, gin.H{"error": "歌曲不存在"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("获取 URL 失败 (code=%.0f)", code)})
		}
		return
	}

	fmt.Printf("[stream] Fetching audio from: %s\n", url[:50])

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("[stream] Create request error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://music.163.com/")

	// 传递 Range 请求头
	if rangeHeader := c.GetHeader("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[stream] Fetch error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取音频失败"})
		return
	}
	defer resp.Body.Close()

	// 设置响应头
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Access-Control-Allow-Origin", "*")

	// 传递状态码
	if resp.StatusCode == http.StatusPartialContent {
		c.Status(http.StatusPartialContent)
	} else {
		c.Status(http.StatusOK)
	}

	// 传递 Content-Range 和 Content-Length
	if cr := resp.Header.Get("Content-Range"); cr != "" {
		c.Header("Content-Range", cr)
	}
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		c.Header("Content-Length", cl)
	}

	io.Copy(c.Writer, resp.Body)
}
