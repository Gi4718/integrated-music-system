package db

import (
	"database/sql"
	"errors"
	"time"
)

// SystemUserInfo 系统用户信息（用于列表展示）
type SystemUserInfo struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	LastLoginAt    time.Time `json:"last_login_at"`
	FailedAttempts int       `json:"failed_attempts"`
	LockedUntil    time.Time `json:"locked_until"`
}

// GetAllSystemUsers 获取所有系统用户
func GetAllSystemUsers() ([]SystemUserInfo, error) {
	rows, err := dbConn.Query(
		"SELECT id, username, role, created_at, COALESCE(last_login_at, '1970-01-01'), failed_attempts, COALESCE(locked_until, '1970-01-01') FROM system_users ORDER BY id ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []SystemUserInfo
	for rows.Next() {
		var u SystemUserInfo
		var lastLogin, lockedUntil string
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt, &lastLogin, &u.FailedAttempts, &lockedUntil); err != nil {
			return nil, err
		}
		if lastLogin != "1970-01-01" {
			u.LastLoginAt, _ = time.Parse("2006-01-02", lastLogin)
		}
		if lockedUntil != "1970-01-01" {
			u.LockedUntil, _ = time.Parse("2006-01-02", lockedUntil)
		}
		users = append(users, u)
	}
	return users, nil
}

// UpdateUserRole 更新用户角色
func UpdateUserRole(userID int, role string) error {
	if role != "admin" && role != "user" {
		return errors.New("无效的角色")
	}
	_, err := dbConn.Exec("UPDATE system_users SET role = ? WHERE id = ?", role, userID)
	return err
}

// DeleteSystemUser 删除系统用户（不能删除最后一个管理员）
func DeleteSystemUser(userID int) error {
	// 检查要删除的用户
	var role string
	err := dbConn.QueryRow("SELECT role FROM system_users WHERE id = ?", userID).Scan(&role)
	if err != nil {
		return errors.New("用户不存在")
	}

	if role == "admin" {
		// 检查是否还有其他管理员
		var adminCount int
		dbConn.QueryRow("SELECT COUNT(*) FROM system_users WHERE role = 'admin' AND id != ?", userID).Scan(&adminCount)
		if adminCount == 0 {
			return errors.New("不能删除最后一个管理员")
		}
	}

	_, err = dbConn.Exec("DELETE FROM system_users WHERE id = ?", userID)
	return err
}

// ChangeUserPassword 修改用户密码
func ChangeUserPassword(userID int, newPasswordHash string) error {
	_, err := dbConn.Exec("UPDATE system_users SET password_hash = ?, failed_attempts = 0, locked_until = NULL WHERE id = ?", newPasswordHash, userID)
	return err
}

// GetSystemUserByID 根据ID获取系统用户
func GetSystemUserByID(userID int) (*SystemUser, error) {
	var user SystemUser
	var lastLoginAt sql.NullTime
	var lockedUntil sql.NullTime

	err := dbConn.QueryRow(
		"SELECT id, username, password_hash, created_at, last_login_at, failed_attempts, locked_until FROM system_users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &lastLoginAt, &user.FailedAttempts, &lockedUntil)
	if err != nil {
		return nil, err
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = lockedUntil.Time
	}
	return &user, nil
}

// GetSystemUserByUsername 根据用户名获取系统用户
func GetSystemUserByUsername(username string) (*SystemUser, error) {
	var user SystemUser
	var lastLoginAt sql.NullTime
	var lockedUntil sql.NullTime

	err := dbConn.QueryRow(
		"SELECT id, username, password_hash, created_at, last_login_at, failed_attempts, locked_until FROM system_users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &lastLoginAt, &user.FailedAttempts, &lockedUntil)
	if err != nil {
		return nil, err
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = lockedUntil.Time
	}
	return &user, nil
}
