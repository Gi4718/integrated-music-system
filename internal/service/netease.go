package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type NeteaseService struct {
	baseURL    string
	httpClient *http.Client
}

func NewNeteaseService(baseURL string) *NeteaseService {
	return &NeteaseService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// QRKey 获取二维码登录 key
func (s *NeteaseService) QRKey() (string, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/login/qr/key?timerstamp=%d", s.baseURL, time.Now().UnixMilli()))
	if err != nil {
		return "", fmt.Errorf("请求二维码 key 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Data struct {
			Unikey string `json:"unikey"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	return result.Data.Unikey, nil
}

// QRCode 生成二维码图片
func (s *NeteaseService) QRCode(key string) (string, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/login/qr/create?key=%s&qrimg=true", s.baseURL, url.QueryEscape(key)))
	if err != nil {
		return "", fmt.Errorf("请求二维码失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Data struct {
			QRImg string `json:"qrimg"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	return result.Data.QRImg, nil
}

// CheckQR 检查二维码扫描状态
func (s *NeteaseService) CheckQR(key string) (int, string, string, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/login/qr/check?key=%s&timerstamp=%d", s.baseURL, url.QueryEscape(key), time.Now().UnixMilli()))
	if err != nil {
		return 0, "", "", fmt.Errorf("请求检查状态失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", "", fmt.Errorf("读取响应失败: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, "", "", fmt.Errorf("解析响应失败: %w", err)
	}

	code := 0
	if c, ok := raw["code"].(float64); ok {
		code = int(c)
	}
	message := ""
	if m, ok := raw["message"].(string); ok {
		message = m
	}

	var cookie string
	switch v := raw["cookie"].(type) {
	case string:
		cookie = v
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				parts = append(parts, s)
			}
		}
		for i, p := range parts {
			if i > 0 {
				cookie += "; "
			}
			cookie += p
		}
	}

	return code, message, cookie, nil
}

// SearchSongs 搜索歌曲
func (s *NeteaseService) SearchSongs(keyword string, limit int, offset int) ([]byte, error) {
	params := url.Values{}
	params.Set("keywords", keyword)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))

	resp, err := s.httpClient.Get(fmt.Sprintf("%s/search?%s", s.baseURL, params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetSongURL 获取歌曲播放 URL
func (s *NeteaseService) GetSongURL(songID int, br int, cookie string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/song/url?id=%d&br=%d", s.baseURL, songID, br), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取歌曲 URL 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetSongDetail 获取歌曲详情
func (s *NeteaseService) GetSongDetail(songID int) ([]byte, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/song/detail?ids=%d", s.baseURL, songID))
	if err != nil {
		return nil, fmt.Errorf("获取歌曲详情失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetSongDetailBatch 批量获取歌曲详情
func (s *NeteaseService) GetSongDetailBatch(ids string) ([]byte, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/song/detail?ids=%s", s.baseURL, ids))
	if err != nil {
		return nil, fmt.Errorf("批量获取歌曲详情失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetLyric 获取歌词
func (s *NeteaseService) GetLyric(songID int) ([]byte, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/lyric?id=%d", s.baseURL, songID))
	if err != nil {
		return nil, fmt.Errorf("获取歌词失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetUserPlaylists 获取用户歌单
func (s *NeteaseService) GetUserPlaylists(uid int) ([]byte, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/user/playlist?uid=%d", s.baseURL, uid))
	if err != nil {
		return nil, fmt.Errorf("获取用户歌单失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetPlaylistDetail 获取歌单详情
func (s *NeteaseService) GetPlaylistDetail(playlistID int, cookie string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/playlist/detail?id=%d", s.baseURL, playlistID), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取歌单详情失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// SubscribePlaylist 收藏歌单
func (s *NeteaseService) SubscribePlaylist(playlistID int, cookie string) ([]byte, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/playlist/subscribe?id=%d", s.baseURL, playlistID), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("收藏歌单失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetLoginStatus 获取登录状态
func (s *NeteaseService) GetLoginStatus(cookie string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/login/status", s.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Cookie", cookie)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取登录状态失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// SendSMSCode 发送手机验证码
func (s *NeteaseService) SendSMSCode(phone string) ([]byte, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/captcha/sent?phone=%s", s.baseURL, url.QueryEscape(phone)))
	if err != nil {
		return nil, fmt.Errorf("发送验证码失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// VerifySMSCode 验证手机验证码
func (s *NeteaseService) VerifySMSCode(phone, captcha string) ([]byte, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/captcha/verify?phone=%s&captcha=%s", s.baseURL, url.QueryEscape(phone), url.QueryEscape(captcha)))
	if err != nil {
		return nil, fmt.Errorf("验证验证码失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// LoginByPhone 手机号登录
func (s *NeteaseService) LoginByPhone(phone, captcha string) ([]byte, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/login/cellphone?phone=%s&captcha=%s", s.baseURL, url.QueryEscape(phone), url.QueryEscape(captcha)))
	if err != nil {
		return nil, fmt.Errorf("手机号登录失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// LoginByPhonePassword 手机号密码登录
func (s *NeteaseService) LoginByPhonePassword(phone, password string) ([]byte, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/login/cellphone?phone=%s&password=%s", s.baseURL, url.QueryEscape(phone), url.QueryEscape(password)))
	if err != nil {
		return nil, fmt.Errorf("手机号密码登录失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// LoginByPhoneWith2FA 带二次验证的手机号登录
func (s *NeteaseService) LoginByPhoneWith2FA(phone, captcha, code string) ([]byte, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/login/cellphone?phone=%s&captcha=%s&code=%s", s.baseURL, url.QueryEscape(phone), url.QueryEscape(captcha), url.QueryEscape(code)))
	if err != nil {
		return nil, fmt.Errorf("二次验证登录失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// LoginByEmail 邮箱登录
func (s *NeteaseService) LoginByEmail(email, password string) ([]byte, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/login?email=%s&password=%s", s.baseURL, url.QueryEscape(email), url.QueryEscape(password)))
	if err != nil {
		return nil, fmt.Errorf("邮箱登录失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// LoginByQQ QQ 登录（需要 OAuth 回调）
func (s *NeteaseService) LoginByQQ(code string) ([]byte, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/login/qq?code=%s", s.baseURL, url.QueryEscape(code)))
	if err != nil {
		return nil, fmt.Errorf("QQ 登录失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetRecommendSongs 获取推荐歌曲
func (s *NeteaseService) GetRecommendSongs(cookie string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/recommend/songs", s.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取推荐歌曲失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetRecommendPlaylists 获取推荐歌单
func (s *NeteaseService) GetRecommendPlaylists(cookie string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/personalized", s.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取推荐歌单失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// CleanCookie 从 Set-Cookie 格式中提取纯 cookie 值
// 输入: "MUSIC_U=xxx; Max-Age=15552000; Expires=...; Path=/;;__csrf=yyy; ..."
// 输出: "MUSIC_U=xxx; __csrf=yyy"
func CleanCookie(rawCookie string) string {
	if rawCookie == "" {
		return ""
	}

	// 按分号分割所有 cookie 片段
	parts := strings.Split(rawCookie, ";")
	var cookiePairs []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 跳过 Set-Cookie 属性（Max-Age, Expires, Path, Domain, Secure, HttpOnly, SameSite）
		lower := strings.ToLower(part)
		if strings.HasPrefix(lower, "max-age=") ||
			strings.HasPrefix(lower, "expires=") ||
			strings.HasPrefix(lower, "path=") ||
			strings.HasPrefix(lower, "domain=") ||
			strings.HasPrefix(lower, "secure") ||
			strings.HasPrefix(lower, "httponly") ||
			strings.HasPrefix(lower, "samesite=") {
			continue
		}

		// 保留 name=value 格式的 cookie
		if strings.Contains(part, "=") {
			cookiePairs = append(cookiePairs, part)
		}
	}

	return strings.Join(cookiePairs, "; ")
}

// GetUserAccount 通过 /user/account 获取用户信息
func (s *NeteaseService) GetUserAccount(cookie string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user/account", s.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetHTTPClient 获取 HTTP 客户端（供 CLI 使用）
func (s *NeteaseService) GetHTTPClient() *http.Client {
	return s.httpClient
}
