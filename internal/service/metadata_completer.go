package service

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"endfield-music/internal/model"
	"endfield-music/internal/util"
)

// MetadataCompleter 元数据补全服务
type MetadataCompleter struct {
	netease     *NeteaseService
	rateLimiter *util.RateLimiter
	retryConfig util.RetryConfig
}

// NewMetadataCompleter 创建元数据补全服务
func NewMetadataCompleter(netease *NeteaseService) *MetadataCompleter {
	return &MetadataCompleter{
		netease:     netease,
		rateLimiter: util.NewRateLimiter(1 * time.Second),
		retryConfig: util.DefaultRetryConfig(),
	}
}

// SongDetail 歌曲详情
type SongDetail struct {
	Name   string
	Artist string
	Album  string
	PicURL string
}

// DownloadAndEmbedCover 下载封面并嵌入 MP3 ID3 APIC
func (c *MetadataCompleter) DownloadAndEmbedCover(song *model.Song, filePath string) ([]byte, error) {
	detail, err := c.getSongDetailWithRetry(song.ID)
	if err != nil {
		return nil, fmt.Errorf("获取歌曲详情失败: %w", err)
	}
	if detail.PicURL == "" {
		return nil, fmt.Errorf("无封面 URL")
	}

	coverData, err := c.downloadCoverData(detail.PicURL)
	if err != nil {
		return nil, fmt.Errorf("下载封面失败: %w", err)
	}

	if err := c.embedCoverToMP3(filePath, coverData, detail); err != nil {
		fmt.Printf("[metadata] 嵌入封面失败 %s: %v\n", song.Name, err)
		return coverData, err
	}

	return coverData, nil
}

// DownloadLyrics 下载歌词保存为 .lrc 文件
func (c *MetadataCompleter) DownloadLyrics(song *model.Song, filePath string) error {
	lyric, err := c.getLyricWithRetry(song.ID)
	if err != nil {
		return fmt.Errorf("获取歌词失败: %w", err)
	}
	if lyric == "" {
		return fmt.Errorf("无歌词")
	}

	base := filePath[:len(filePath)-len(filepath.Ext(filePath))]
	lyricPath := base + ".lrc"
	if err := os.WriteFile(lyricPath, []byte(lyric), 0644); err != nil {
		return fmt.Errorf("保存歌词失败: %w", err)
	}
	return nil
}

// EmbedArtistInfo 将艺人信息嵌入 MP3 ID3 标签
func (c *MetadataCompleter) EmbedArtistInfo(song *model.Song, filePath string) error {
	detail, err := c.getSongDetailWithRetry(song.ID)
	if err != nil {
		return fmt.Errorf("获取歌曲详情失败: %w", err)
	}

	if err := c.embedFullID3(filePath, song, detail, nil); err != nil {
		return fmt.Errorf("嵌入 ID3 标签失败: %w", err)
	}
	return nil
}

// embedCoverToMP3 将封面嵌入 MP3（同时写入基本标签信息）
func (c *MetadataCompleter) embedCoverToMP3(filePath string, coverData []byte, detail *SongDetail) error {
	song := &model.Song{
		Name:   detail.Name,
		Artist: detail.Artist,
		Album:  detail.Album,
	}
	return c.embedFullID3(filePath, song, detail, coverData)
}

// embedFullID3 写入完整 ID3v2.3 标签（TIT2/TPE1/TALB + APIC）
func (c *MetadataCompleter) embedFullID3(filePath string, song *model.Song, detail *SongDetail, coverData []byte) error {
	// 读取原始音频数据（跳过旧 ID3 标签）
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	header := make([]byte, 10)
	if _, err := io.ReadFull(f, header); err != nil {
		f.Close()
		return err
	}

	var audioStart int64 = 0
	if string(header[:3]) == "ID3" {
		size := int64(header[6])<<21 | int64(header[7])<<14 | int64(header[8])<<7 | int64(header[9])
		audioStart = 10 + size
	}
	f.Seek(audioStart, io.SeekStart)

	audioData, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		return err
	}

	// 构建 ID3v2.3 标签帧
	var frames bytes.Buffer

	// TIT2 - 标题
	if song.Name != "" {
		writeTextFrame(&frames, "TIT2", song.Name)
	}
	// TPE1 - 艺人
	if song.Artist != "" {
		writeTextFrame(&frames, "TPE1", song.Artist)
	}
	// TALB - 专辑
	if song.Album != "" {
		writeTextFrame(&frames, "TALB", song.Album)
	}
	// APIC - 封面
	if len(coverData) > 0 {
		writeAPICFrame(&frames, coverData)
	}

	// 写入新文件
	tmpPath := filePath + ".tagtmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	tagBody := frames.Bytes()
	tagBodySize := len(tagBody)

	// ID3v2.3 头部 (10 bytes)
	id3Header := make([]byte, 10)
	copy(id3Header, []byte("ID3"))
	id3Header[3] = 3 // version 2.3
	id3Header[4] = 0 // flags
	// syncsafe integer for size
	id3Header[6] = byte((tagBodySize >> 21) & 0x7F)
	id3Header[7] = byte((tagBodySize >> 14) & 0x7F)
	id3Header[8] = byte((tagBodySize >> 7) & 0x7F)
	id3Header[9] = byte(tagBodySize & 0x7F)

	out.Write(id3Header)
	out.Write(tagBody)
	out.Write(audioData)
	out.Close()

	return os.Rename(tmpPath, filePath)
}

// writeTextFrame 写入 ID3v2.3 文本帧 (encoding=3 UTF-8)
func writeTextFrame(buf *bytes.Buffer, frameID, text string) {
	var frame bytes.Buffer
	frame.WriteString(frameID)

	// 帧数据: encoding(1) + text
	var data bytes.Buffer
	data.WriteByte(3) // UTF-8
	data.WriteString(text)

	frameData := data.Bytes()
	sizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuf, uint32(len(frameData)))
	frame.Write(sizeBuf)
	frame.Write([]byte{0, 0}) // flags
	frame.Write(frameData)

	buf.Write(frame.Bytes())
}

// writeAPICFrame 写入 ID3v2.3 APIC 帧
func writeAPICFrame(buf *bytes.Buffer, coverData []byte) {
	var frame bytes.Buffer
	frame.WriteString("APIC")

	var data bytes.Buffer
	data.WriteByte(0)                  // encoding: ISO-8859-1
	data.WriteString("image/jpeg")     // MIME type
	data.WriteByte(0)                  // separator
	data.WriteByte(0x03)               // picture type: Front Cover
	data.WriteByte(0)                  // description: empty (null terminated)
	data.Write(coverData)

	frameData := data.Bytes()
	sizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuf, uint32(len(frameData)))
	frame.Write(sizeBuf)
	frame.Write([]byte{0, 0}) // flags
	frame.Write(frameData)

	buf.Write(frame.Bytes())
}

// downloadCoverData 下载封面图片数据
func (c *MetadataCompleter) downloadCoverData(url string) ([]byte, error) {
	c.rateLimiter.Wait()

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	_, _, err = image.DecodeConfig(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("无效的图片: %w", err)
	}

	resp2, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	return io.ReadAll(resp2.Body)
}

// isRateLimited 检测网易云 API 限速错误
func isRateLimited(result map[string]interface{}) bool {
	if code, ok := result["code"]; ok {
		switch v := code.(type) {
		case float64:
			return v == -1 || v == 429 || v == 301
		case int:
			return v == -1 || v == 429 || v == 301
		}
	}
	return false
}

// getSongDetailWithRetry 带重试的歌曲详情获取（自适应限速）
func (c *MetadataCompleter) getSongDetailWithRetry(songID int) (*SongDetail, error) {
	var detail *SongDetail
	err := util.RetryWithBackoff(func() error {
		c.rateLimiter.Wait()

		body, err := c.netease.GetSongDetail(songID)
		if err != nil {
			return err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return err
		}

		// 检测限速
		if isRateLimited(result) {
			c.rateLimiter.Increase()
			fmt.Printf("[ratelimit] 检测到限速，间隔调整为 %v\n", c.rateLimiter.GetInterval())
			return fmt.Errorf("API 限速")
		}

		songs, ok := result["songs"].([]interface{})
		if !ok || len(songs) == 0 {
			return fmt.Errorf("未找到歌曲")
		}

		song := songs[0].(map[string]interface{})
		detail = &SongDetail{}
		if v, ok := song["name"].(string); ok {
			detail.Name = v
		}

		if ar, ok := song["ar"].([]interface{}); ok && len(ar) > 0 {
			if artist, ok := ar[0].(map[string]interface{}); ok {
				if n, ok := artist["name"].(string); ok {
					detail.Artist = n
				}
			}
		}

		if al, ok := song["al"].(map[string]interface{}); ok {
			if v, ok := al["name"].(string); ok {
				detail.Album = v
			}
			if pic, ok := al["picUrl"].(string); ok {
				detail.PicURL = pic
			}
		}

		c.rateLimiter.Decrease()
		return nil
	}, c.retryConfig)

	return detail, err
}

// getLyricWithRetry 带重试的歌词获取（自适应限速）
func (c *MetadataCompleter) getLyricWithRetry(songID int) (string, error) {
	var lyric string
	err := util.RetryWithBackoff(func() error {
		c.rateLimiter.Wait()

		body, err := c.netease.GetLyric(songID)
		if err != nil {
			return err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return err
		}

		// 检测限速
		if isRateLimited(result) {
			c.rateLimiter.Increase()
			fmt.Printf("[ratelimit] 检测到限速，间隔调整为 %v\n", c.rateLimiter.GetInterval())
			return fmt.Errorf("API 限速")
		}

		if lrc, ok := result["lrc"].(map[string]interface{}); ok {
			if v, ok := lrc["lyric"].(string); ok {
				lyric = v
			}
		}

		c.rateLimiter.Decrease()
		return nil
	}, c.retryConfig)

	return lyric, err
}
