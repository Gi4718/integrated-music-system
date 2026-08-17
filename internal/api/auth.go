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
// 如果无法从 JWT 获取，则使用第一个系统用户作为 fallback
func getSystemUserID(c *gin.Context) int {
	if id, exists := c.Get("system_user_id"); exists {
		if userID, ok := id.(int); ok && userID > 0 {
			return userID
		}
	}
	// Fallback: 查询第一个系统用户
	var firstUserID int
	err := db.GetDB().QueryRow("SELECT id FROM system_users ORDER BY id LIMIT 1").Scan(&firstUserID)
	if err == nil && firstUserID > 0 {
		return firstUserID
	}
	return 1 // 最终 fallback
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

	// cookie 已过期，返回未登录
	if !cookieValid {
		c.JSON(http.StatusOK, gin.H{
			"logged_in":    false,
			"cookie_valid": false,
		})
		return
	}

	// cookie 未过期，调用网易云API获取最新用户信息（包括VIP状态）
	// 即使API调用失败，也返回已登录状态，但不清除用户数据
	netease := service.NewNeteaseService("http://127.0.0.1:3000")
	accountBody, err := netease.GetUserAccount(user.Cookie)
	
	vipType := 0
	nickname := user.Nickname
	avatarURL := user.AvatarURL
	
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
				if nick, ok := profile["nickname"].(string); ok && nick != "" {
					nickname = nick
				}
				if avatar, ok := profile["avatarUrl"].(string); ok && avatar != "" {
					avatarURL = avatar
				} else if avatar, ok := profile["avatar"].(string); ok && avatar != "" {
					avatarURL = avatar
				}
				if vt, ok := profile["vipType"].(float64); ok {
					vipType = int(vt)
				}
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"logged_in":      true,
		"cookie_valid":   true,
		"cookie_expires": user.CookieExpires.Format(time.RFC3339),
		"vipType":        vipType,
		"user": gin.H{
			"user_id":  user.UserID,
			"nickname": nickname,
			"avatar":   avatarURL,
			"vipType":  vipType,
		},
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

    // 如果用户ID为0，说明API返回的数据异常
    if userID == 0 {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
      return
    }

    // 清理 cookie
    cleanCookie := service.CleanCookie(cookie)

    user := &model.User{
      UserID:        int(userID),
      Nickname:      nickname,
      AvatarURL:     avatarURL,
      Cookie:        cleanCookie,
      CookieExpires: time.Now().Add(30 * 24 * time.Hour),
    }
    systemUserID := getSystemUserID(c)
    if err := db.SaveUserForSystem(systemUserID, user); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "保存用户信息失败"})
      return
    }

    c.JSON(http.StatusOK, gin.H{
      "code":    200,
      "message": "登录成功",
      "cookie":  cleanCookie,
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

    // 如果用户ID为0，说明API返回的数据异常
    if userID == 0 {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
      return
    }

    // 清理 cookie
    cleanCookie := service.CleanCookie(cookie)

    user := &model.User{
      UserID:        int(userID),
      Nickname:      nickname,
      AvatarURL:     avatarURL,
      Cookie:        cleanCookie,
      CookieExpires: time.Now().Add(30 * 24 * time.Hour),
    }
    systemUserID := getSystemUserID(c)
    if err := db.SaveUserForSystem(systemUserID, user); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "保存用户信息失败"})
      return
    }

    c.JSON(http.StatusOK, gin.H{
      "code":    200,
      "message": "登录成功",
      "cookie":  cleanCookie,
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
    Phone   string `json:"phone" binding:"required"`
    Captcha string `json:"captcha" binding:"required"`
    Code    string `json:"code" binding:"required"`
  }
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "参数不完整"})
    return
  }

  netease := service.NewNeteaseService("http://127.0.0.1:3000")
  body, err := netease.LoginByPhoneWith2FA(req.Phone, req.Captcha, req.Code)
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

    // 如果用户ID为0，说明API返回的数据异常
    if userID == 0 {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
      return
    }

    // 清理 cookie
    cleanCookie := service.CleanCookie(cookie)

    user := &model.User{
      UserID:        int(userID),
      Nickname:      nickname,
      AvatarURL:     avatarURL,
      Cookie:        cleanCookie,
      CookieExpires: time.Now().Add(30 * 24 * time.Hour),
    }
    systemUserID := getSystemUserID(c)
    if err := db.SaveUserForSystem(systemUserID, user); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "保存用户信息失败"})
      return
    }

    c.JSON(http.StatusOK, gin.H{
      "code":    200,
      "message": "登录成功",
      "cookie":  cleanCookie,
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

    // 如果用户ID为0，说明API返回的数据异常
    if userID == 0 {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
      return
    }

    // 清理 cookie
    cleanCookie := service.CleanCookie(cookie)

    user := &model.User{
      UserID:        int(userID),
      Nickname:      nickname,
      AvatarURL:     avatarURL,
      Cookie:        cleanCookie,
      CookieExpires: time.Now().Add(30 * 24 * time.Hour),
    }
    systemUserID := getSystemUserID(c)
    if err := db.SaveUserForSystem(systemUserID, user); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "保存用户信息失败"})
      return
    }

    c.JSON(http.StatusOK, gin.H{
      "code":    200,
      "message": "登录成功",
      "cookie":  cleanCookie,
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
