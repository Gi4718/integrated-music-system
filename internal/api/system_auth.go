package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"endfield-music/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("endfield-music-secret-key-2026")

type SystemLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SystemRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

// Register 系统用户注册
func Register(c *gin.Context) {
	var req SystemRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "detail": err.Error()})
		return
	}

	// 检查是否已有系统用户
	hasUser, err := db.HasSystemUser()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	if hasUser {
		c.JSON(http.StatusForbidden, gin.H{"error": "系统已有管理员账号"})
		return
	}

	// 创建首个系统用户（自动为 admin）
	hash := hashPassword(req.Password)
	_, err = db.GetDB().Exec(
		"INSERT INTO system_users (username, password_hash, role, created_at, failed_attempts) VALUES (?, ?, 'admin', ?, 0)",
		req.Username, hash, time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

// SystemLogin 系统用户登录
func SystemLogin(c *gin.Context) {
	var req SystemLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 验证用户
	user, err := db.AuthenticateSystemUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 生成JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     "admin",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    tokenString,
		"username": user.Username,
		"role":     "admin",
	})
}

// CheckSystemUser 检查是否已有系统用户
func CheckSystemUser(c *gin.Context) {
	hasUser, err := db.HasSystemUser()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"has_user": hasUser})
}

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否已有系统用户，如果没有则允许访问注册页面
		hasUser, _ := db.HasSystemUser()
		if !hasUser {
			c.Next()
			return
		}

		// 获取token（优先从 Authorization header，其次从 query 参数）
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			tokenString = c.Query("token")
		}
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			c.Abort()
			return
		}

		// 移除 "Bearer " 前缀
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// 验证token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token无效或已过期"})
			c.Abort()
			return
		}

		// 从 token 中提取用户信息并注入到 context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, ok := claims["user_id"].(float64); ok {
				c.Set("system_user_id", int(userID))
			}
			if username, ok := claims["username"].(string); ok {
				c.Set("system_username", username)
			}
		}

		c.Next()
	}
}

// hashPassword 使用SHA256加密密码
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
