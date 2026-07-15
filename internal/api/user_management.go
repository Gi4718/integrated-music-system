package api

import (
	"endfield-music/internal/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetCurrentUser 获取当前登录的系统用户信息
func GetCurrentUser(c *gin.Context) {
	systemUserID := getSystemUserID(c)
	if systemUserID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 从数据库获取用户信息
	user, err := db.GetSystemUserByID(systemUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}

// ListUsers 列出所有系统用户（仅管理员）
func ListUsers(c *gin.Context) {
	// 检查当前用户是否是管理员
	currentUserID := getSystemUserID(c)
	currentUser, err := db.GetSystemUserByID(currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取当前用户失败"})
		return
	}

	if currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "仅管理员可以查看用户列表"})
		return
	}

	users, err := db.GetAllSystemUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// UpdateUserRole 更新用户角色（仅管理员）
func UpdateUserRole(c *gin.Context) {
	// 检查当前用户是否是管理员
	currentUserID := getSystemUserID(c)
	currentUser, err := db.GetSystemUserByID(currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取当前用户失败"})
		return
	}

	if currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "仅管理员可以修改用户角色"})
		return
	}

	var req struct {
		UserID int    `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required,oneof=admin user"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 不能修改自己的角色
	if req.UserID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能修改自己的角色"})
		return
	}

	if err := db.UpdateUserRole(req.UserID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新角色失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色已更新"})
}

// UpdateUserPassword 更新用户密码（管理员或本人）
func UpdateUserPassword(c *gin.Context) {
	currentUserID := getSystemUserID(c)

	var req struct {
		UserID      int    `json:"user_id" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，密码至少6位"})
		return
	}

	// 检查权限：只能修改自己的密码，或管理员修改他人密码
	if req.UserID != currentUserID {
		currentUser, err := db.GetSystemUserByID(currentUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取当前用户失败"})
			return
		}
		if currentUser.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "只能修改自己的密码"})
			return
		}
	}

	if err := db.UpdateUserPassword(req.UserID, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码已更新"})
}

// DeleteUser 删除用户（仅管理员，不能删除自己）
func DeleteUser(c *gin.Context) {
	// 检查当前用户是否是管理员
	currentUserID := getSystemUserID(c)
	currentUser, err := db.GetSystemUserByID(currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取当前用户失败"})
		return
	}

	if currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "仅管理员可以删除用户"})
		return
	}

	var req struct {
		UserID int `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 不能删除自己
	if req.UserID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除自己"})
		return
	}

	// 删除用户及其关联数据
	if err := db.DeleteSystemUser(req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}

// RegisterUser 新用户注册（需要管理员开启多用户注册）
func RegisterUser(c *gin.Context) {
	// 检查是否允许多用户注册
	multiUserEnabled, err := db.GetSetting("multi_user_enabled")
	if err != nil || multiUserEnabled != "true" {
		c.JSON(http.StatusForbidden, gin.H{"error": "系统未开启多用户注册"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
		Password string `json:"password" binding:"required,min=6,max=50"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：用户名3-20位，密码至少6位"})
		return
	}

	// 检查用户名是否已存在
	if db.UsernameExists(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}

	// 创建新用户
	if err := db.CreateSystemUserWithRole(req.Username, req.Password, "user"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

// GetSystemUserByID 从数据库获取系统用户
func GetSystemUserByID(id int) (*db.SystemUser, error) {
	return db.GetSystemUserByID(id)
}
