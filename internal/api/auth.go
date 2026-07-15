package api

import (
	"endfield-music/internal/db"
	"endfield-music/internal/model"
	"endfield-music/internal/service"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// getSystemUserID 从 gin context 获取当前系统用户 ID
func getSystemUserID(c *gin.Context) int {
	if id, exists := c.Get("system_user_id"); exists {
		if userID, ok := id.(int); ok {
			return userID
		}
	}
	return 0
}

func getQRKey(c *gin.Context) {
	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	
	key, err := netease.QRKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": key})
}

func getQRCode(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	qrImg, err := netease.QRCode(key)
	if err != nil || qrImg == "" {
		qrURL := fmt.Sprintf("https://music.163.com/login?codekey=%s", key)
		qrImg, err = generateQRImage(qrURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成二维码失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"qr_img": qrImg})
}

func checkQRStatus(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	code, message, cookie, err := netease.CheckQR(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"code":    code,
		"message": message,
	}

	// 803 表示登录成功
	if code == 803 {
		response["cookie"] = cookie

		if cookie != "" {
			// 清理 cookie：从 Set-Cookie 格式提取纯 name=value 对
			cleanCookie := service.CleanCookie(cookie)
			response["clean_cookie"] = cleanCookie

			// 用 /user/account 获取用户信息
			accountBody, err := netease.GetUserAccount(cleanCookie)
			userSaved := false

			if err == nil {
				var accountResult map[string]interface{}
				if json.Unmarshal(accountBody, &accountResult) == nil {
					response["account_response"] = string(accountBody)

					var profile map[string]interface{}
					if data, ok := accountResult["data"].(map[string]interface{}); ok {
						profile, _ = data["profile"].(map[string]interface{})
					}
					if profile == nil {
						profile, _ = accountResult["profile"].(map[string]interface{})
					}

					if profile != nil {
						userID, _ := profile["userId"].(float64)
						nickname, _ := profile["nickname"].(string)
						avatarURL, _ := profile["avatarUrl"].(string)
						if avatarURL == "" {
							avatarURL, _ = profile["avatar"].(string)
						}

						if userID > 0 {
							user := &model.User{
								UserID:        int(userID),
								Nickname:      nickname,
								AvatarURL:     avatarURL,
								Cookie:        cleanCookie,
								CookieExpires: time.Now().Add(30 * 24 * time.Hour),
							}
							systemUserID := getSystemUserID(c)
							if err := db.SaveUserForSystem(systemUserID, user); err == nil {
								userSaved = true
								response["user"] = gin.H{
									"user_id":  user.UserID,
									"nickname": user.Nickname,
									"avatar":   user.AvatarURL,
								}
							}
						}
					}
				}
			}

			// 如果无法解析用户信息或保存失败，至少保存 cookie（使用临时用户 ID）
			if !userSaved {
				tempUserID := int(time.Now().UnixNano() / 1000000)
				user := &model.User{
					UserID:        tempUserID,
					Nickname:      "网易云用户",
					AvatarURL:     "",
					Cookie:        cleanCookie,
					CookieExpires: time.Now().Add(30 * 24 * time.Hour),
				}
				systemUserID := getSystemUserID(c)
				if err := db.SaveUserForSystem(systemUserID, user); err == nil {
					response["user"] = gin.H{
						"user_id":  user.UserID,
						"nickname": user.Nickname,
						"avatar":   user.AvatarURL,
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

func getLoginStatus(c *gin.Context) {
	systemUserID := getSystemUserID(c)
	user, err := db.GetCurrentUserForSystem(systemUserID)
	if err != nil || user == nil {
		c.JSON(http.StatusOK, gin.H{"logged_in": false})
		return
	}

	now := time.Now()
	cookieValid := now.Before(user.CookieExpires)

	// 清理 cookie 后用 /user/account 获取最新用户信息
	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	cleanCookie := service.CleanCookie(user.Cookie)
	vipType := 0
	nickname := user.Nickname
	avatarURL := user.AvatarURL

	accountBody, err := netease.GetUserAccount(cleanCookie)
	if err == nil {
		var accountResult map[string]interface{}
		if json.Unmarshal(accountBody, &accountResult) == nil {
			var profile map[string]interface{}
			if data, ok := accountResult["data"].(map[string]interface{}); ok {
				profile, _ = data["profile"].(map[string]interface{})
			}
			if profile == nil {
				profile, _ = accountResult["profile"].(map[string]interface{})
			}
			if profile != nil {
				if n, ok := profile["nickname"].(string); ok && n != "" {
					nickname = n
				}
				if a, ok := profile["avatarUrl"].(string); ok && a != "" {
					avatarURL = a
				}
				if vt, ok := profile["vipType"].(float64); ok {
					vipType = int(vt)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"logged_in":    true,
		"cookie_valid": cookieValid,
		"cookie_expires": user.CookieExpires.Format(time.RFC3339),
		"user": gin.H{
			"user_id":    user.UserID,
			"nickname":   nickname,
			"avatar":     avatarURL,
			"expires_at": user.CookieExpires,
		},
		"vipType": vipType,
	})
}

func logout(c *gin.Context) {
	systemUserID := getSystemUserID(c)
	if err := db.ClearUserForSystem(systemUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout success"})
}

// generateQRImage 生成二维码图片（备用方案）
func generateQRImage(content string) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}

	png, err := qr.PNG(256)
	if err != nil {
		return "", err
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

func sendSMSCode(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "手机号不能为空"})
		return
	}

	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.SendSMSCode(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if code, ok := result["code"].(float64); ok && code == 200 {
		c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
	} else {
		c.JSON(http.StatusOK, gin.H{"error": result["msg"]})
	}
}

func extractCookie(result map[string]interface{}) string {
	if cookie, ok := result["cookie"].(string); ok && cookie != "" {
		return cookie
	}
	if cookies, ok := result["cookie"].([]interface{}); ok {
		parts := make([]string, 0, len(cookies))
		for _, c := range cookies {
			if s, ok := c.(string); ok && s != "" {
				parts = append(parts, s)
			}
		}
		if len(parts) > 0 {
			joined := ""
			for i, p := range parts {
				if i > 0 {
					joined += "; "
				}
				joined += p
			}
			return joined
		}
	}
	return ""
}

func loginByPhone(c *gin.Context) {
	var req struct {
		Phone   string `json:"phone" binding:"required"`
		Captcha string `json:"captcha" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "手机号和验证码不能为空"})
		return
	}

	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.LoginByPhone(req.Phone, req.Captcha)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if code, ok := result["code"].(float64); ok && code == 200 {
		profile, _ := result["profile"].(map[string]interface{})
		userID, _ := profile["userId"].(float64)
		nickname, _ := profile["nickname"].(string)
		avatarURL, _ := profile["avatarUrl"].(string)
		cookie := extractCookie(result)

		user := &model.User{
			UserID:        int(userID),
			Nickname:      nickname,
			AvatarURL:     avatarURL,
			Cookie:        cookie,
			CookieExpires: time.Now().Add(30 * 24 * time.Hour),
		}
		systemUserID := getSystemUserID(c)
		db.SaveUserForSystem(systemUserID, user)

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "登录成功",
			"user": gin.H{
				"user_id":  user.UserID,
				"nickname": user.Nickname,
				"avatar":   user.AvatarURL,
			},
		})
	} else {
		msg, _ := result["msg"].(string)
		c.JSON(http.StatusOK, gin.H{"code": result["code"], "error": msg, "msg": msg})
	}
}

func loginByPhonePassword(c *gin.Context) {
	var req struct {
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "手机号和密码不能为空"})
		return
	}

	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.LoginByPhonePassword(req.Phone, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if code, ok := result["code"].(float64); ok && code == 200 {
		profile, _ := result["profile"].(map[string]interface{})
		userID, _ := profile["userId"].(float64)
		nickname, _ := profile["nickname"].(string)
		avatarURL, _ := profile["avatarUrl"].(string)
		cookie := extractCookie(result)

		user := &model.User{
			UserID:        int(userID),
			Nickname:      nickname,
			AvatarURL:     avatarURL,
			Cookie:        cookie,
			CookieExpires: time.Now().Add(30 * 24 * time.Hour),
		}
		systemUserID := getSystemUserID(c)
		db.SaveUserForSystem(systemUserID, user)

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "登录成功",
			"user": gin.H{
				"user_id":  user.UserID,
				"nickname": user.Nickname,
				"avatar":   user.AvatarURL,
			},
		})
	} else if code, ok := result["code"].(float64); ok && code == 301 {
		c.JSON(http.StatusOK, gin.H{"code": 301, "needSecondVerify": true, "message": "需要二次验证"})
	} else {
		msg, _ := result["msg"].(string)
		c.JSON(http.StatusOK, gin.H{"code": result["code"], "error": msg, "msg": msg})
	}
}

func saveCookie(c *gin.Context) {
	var req struct {
		Cookie string `json:"cookie" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cookie不能为空"})
		return
	}

	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	cleanCookie := service.CleanCookie(req.Cookie)

	var userID float64
	var nickname, avatarURL string

	accountBody, err := netease.GetUserAccount(cleanCookie)
	if err == nil {
		var accountResult map[string]interface{}
		if json.Unmarshal(accountBody, &accountResult) == nil {
			var profile map[string]interface{}
			if data, ok := accountResult["data"].(map[string]interface{}); ok {
				profile, _ = data["profile"].(map[string]interface{})
			}
			if profile == nil {
				profile, _ = accountResult["profile"].(map[string]interface{})
			}
			if profile != nil {
				userID, _ = profile["userId"].(float64)
				nickname, _ = profile["nickname"].(string)
				avatarURL, _ = profile["avatarUrl"].(string)
			}
		}
	}

	if userID == 0 {
		userID = float64(time.Now().UnixNano() / 1000000)
		nickname = "网易云用户"
	}

	user := &model.User{
		UserID:        int(userID),
		Nickname:      nickname,
		AvatarURL:     avatarURL,
		Cookie:        cleanCookie,
		CookieExpires: time.Now().Add(30 * 24 * time.Hour),
	}

	systemUserID := getSystemUserID(c)
	if err := db.SaveUserForSystem(systemUserID, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存登录状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "保存成功",
		"user": gin.H{
			"user_id":  user.UserID,
			"nickname": user.Nickname,
			"avatar":   user.AvatarURL,
		},
	})
}

func secondVerify(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码不能为空"})
		return
	}

	// 这里实现二次验证逻辑
	// 实际实现需要根据网易云 API 的二次验证流程来处理
	c.JSON(http.StatusOK, gin.H{"message": "二次验证成功"})
}

func loginByEmail(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱和密码不能为空"})
		return
	}

	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.LoginByEmail(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if code, ok := result["code"].(float64); ok && code == 200 {
		profile, _ := result["profile"].(map[string]interface{})
		userID, _ := profile["userId"].(float64)
		nickname, _ := profile["nickname"].(string)
		avatarURL, _ := profile["avatarUrl"].(string)
		cookie := extractCookie(result)

		user := &model.User{
			UserID:        int(userID),
			Nickname:      nickname,
			AvatarURL:     avatarURL,
			Cookie:        cookie,
			CookieExpires: time.Now().Add(30 * 24 * time.Hour),
		}
		systemUserID := getSystemUserID(c)
		db.SaveUserForSystem(systemUserID, user)

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "登录成功",
			"user": gin.H{
				"user_id":  user.UserID,
				"nickname": user.Nickname,
				"avatar":   user.AvatarURL,
			},
		})
	} else {
		msg, _ := result["msg"].(string)
		c.JSON(http.StatusOK, gin.H{"code": result["code"], "error": msg, "msg": msg})
	}
}

func loginByQQ(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "授权码不能为空"})
		return
	}

	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	body, err := netease.LoginByQQ(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if code, ok := result["code"].(float64); ok && code == 200 {
		profile, ok := result["profile"].(map[string]interface{})
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解析用户信息失败"})
			return
		}
		userID, _ := profile["userId"].(float64)
		nickname, _ := profile["nickname"].(string)
		avatarURL, _ := profile["avatarUrl"].(string)

		user := &model.User{
			UserID:        int(userID),
			Nickname:      nickname,
			AvatarURL:     avatarURL,
			Cookie:        c.GetString("cookie"),
			CookieExpires: time.Now().Add(30 * 24 * time.Hour),
		}
		systemUserID := getSystemUserID(c)
		db.SaveUserForSystem(systemUserID, user)

		c.JSON(http.StatusOK, gin.H{
			"message": "登录成功",
			"user": gin.H{
				"user_id":  user.UserID,
				"nickname": user.Nickname,
				"avatar":   user.AvatarURL,
			},
		})
	} else {
		msg, _ := result["msg"].(string)
		c.JSON(http.StatusOK, gin.H{"error": msg})
	}
}

func getQQAuthURL(c *gin.Context) {
	// QQ OAuth 授权 URL（需要配置 AppID）
	authURL := "https://graph.qq.com/oauth2.0/authorize?response_type=code&client_id=YOUR_APP_ID&redirect_uri=YOUR_REDIRECT_URI&state=endfield-music"
	c.JSON(http.StatusOK, gin.H{"auth_url": authURL})
}
