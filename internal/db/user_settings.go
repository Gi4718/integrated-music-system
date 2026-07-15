package db

import (
	"database/sql"
	"endfield-music/internal/model"
	"time"
)

// GetSetting 获取全局设置（向后兼容）
func GetSetting(key string) (string, error) {
	var value string
	err := dbConn.QueryRow("SELECT value FROM settings WHERE key = ? AND (user_id = 0 OR user_id IS NULL)", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// GetSettingByUser 获取指定用户的设置，如果不存在则回退到全局设置
func GetSettingByUser(userID int, key string) (string, error) {
	if userID <= 0 {
		return GetSetting(key)
	}
	var value string
	err := dbConn.QueryRow("SELECT value FROM settings WHERE key = ? AND user_id = ?", key, userID).Scan(&value)
	if err == sql.ErrNoRows {
		return GetSetting(key)
	}
	return value, err
}

// SetSetting 设置全局设置（向后兼容）
func SetSetting(key, value string) error {
	query := `INSERT INTO settings (key, value, user_id, updated_at) VALUES (?, ?, 0, ?)
		  ON CONFLICT(key, user_id) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`
	_, err := dbConn.Exec(query, key, value, time.Now())
	return err
}

// SetSettingByUser 设置指定用户的设置
func SetSettingByUser(userID int, key, value string) error {
	if userID <= 0 {
		return SetSetting(key, value)
	}
	query := `INSERT INTO settings (key, value, user_id, updated_at) VALUES (?, ?, ?, ?)
		  ON CONFLICT(key, user_id) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`
	_, err := dbConn.Exec(query, key, value, userID, time.Now())
	return err
}

// GetAllSettingsByUser 获取指定用户的所有设置（全局+用户覆盖）
func GetAllSettingsByUser(userID int) (map[string]string, error) {
	settings := make(map[string]string)

	// 先获取全局设置
	rows, err := dbConn.Query("SELECT key, value FROM settings WHERE user_id = 0 OR user_id IS NULL")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		settings[k] = v
	}
	rows.Close()

	// 用用户设置覆盖全局设置
	if userID > 0 {
		rows, err = dbConn.Query("SELECT key, value FROM settings WHERE user_id = ?", userID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			settings[k] = v
		}
		rows.Close()
	}

	return settings, nil
}

// DeleteUserSettings 删除指定用户的所有个性化设置
func DeleteUserSettings(userID int) error {
	_, err := dbConn.Exec("DELETE FROM settings WHERE user_id = ?", userID)
	return err
}

// GetSettingBool 读取布尔类型设置
func GetSettingBool(key string, defaultVal bool) bool {
	val, _ := GetSetting(key)
	if val == "" {
		return defaultVal
	}
	return val == "true"
}

// GetSettingBoolByUser 读取指定用户的布尔类型设置
func GetSettingBoolByUser(userID int, key string, defaultVal bool) bool {
	val, _ := GetSettingByUser(userID, key)
	if val == "" {
		return defaultVal
	}
	return val == "true"
}

// GetMultiUserEnabled 检查是否允许多用户注册
func GetMultiUserEnabled() bool {
	return GetSettingBool("multi_user_enabled", false)
}

// NeteaseUserBinding 系统用户与网易云账号的绑定关系
type NeteaseUserBinding struct {
	SystemUserID   int
	SystemUsername string
	Role           string
	NeteaseUserID  int
	NeteaseNick    string
	Cookie         string
	CookieExpires  time.Time
}

// GetAllNeteaseUserBindings 获取所有系统用户及其绑定的网易云账号
func GetAllNeteaseUserBindings() ([]NeteaseUserBinding, error) {
	query := `SELECT su.id, su.username, COALESCE(su.role, 'user'),
		  COALESCE(u.user_id, 0), COALESCE(u.nickname, ''), COALESCE(u.cookie, ''), COALESCE(u.cookie_expires, '1970-01-01')
		  FROM system_users su
		  LEFT JOIN users u ON u.system_user_id = su.id
		  ORDER BY su.id ASC`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bindings []NeteaseUserBinding
	for rows.Next() {
		var b NeteaseUserBinding
		var cookieExpires string
		err := rows.Scan(&b.SystemUserID, &b.SystemUsername, &b.Role, &b.NeteaseUserID, &b.NeteaseNick, &b.Cookie, &cookieExpires)
		if err != nil {
			return nil, err
		}
		if cookieExpires != "" && cookieExpires != "1970-01-01" {
			b.CookieExpires, _ = time.Parse(time.RFC3339, cookieExpires)
		}
		bindings = append(bindings, b)
	}
	return bindings, nil
}

// SaveUserForSystem 为指定系统用户保存网易云账号信息
func SaveUserForSystem(systemUserID int, user *model.User) error {
	query := `INSERT INTO users (user_id, nickname, avatar_url, cookie, cookie_expires, system_user_id, updated_at)
		  VALUES (?, ?, ?, ?, ?, ?, ?)
		  ON CONFLICT(user_id) DO UPDATE SET
		  nickname = excluded.nickname,
		  avatar_url = excluded.avatar_url,
		  cookie = excluded.cookie,
		  cookie_expires = excluded.cookie_expires,
		  system_user_id = excluded.system_user_id,
		  updated_at = excluded.updated_at`

	_, err := dbConn.Exec(
		query,
		user.UserID,
		user.Nickname,
		user.AvatarURL,
		user.Cookie,
		user.CookieExpires,
		systemUserID,
		time.Now(),
	)
	return err
}

// GetCurrentUserForSystem 获取指定系统用户的网易云账号
func GetCurrentUserForSystem(systemUserID int) (*model.User, error) {
	query := `SELECT id, user_id, nickname, avatar_url, cookie, cookie_expires, created_at, updated_at
		  FROM users WHERE system_user_id = ? ORDER BY updated_at DESC LIMIT 1`

	row := dbConn.QueryRow(query, systemUserID)

	var user model.User
	var cookieExpires sql.NullTime
	err := row.Scan(
		&user.ID,
		&user.UserID,
		&user.Nickname,
		&user.AvatarURL,
		&user.Cookie,
		&cookieExpires,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if cookieExpires.Valid {
		user.CookieExpires = cookieExpires.Time
	}

	return &user, nil
}

// ClearUserForSystem 清除指定系统用户的网易云账号
func ClearUserForSystem(systemUserID int) error {
	_, err := dbConn.Exec(`DELETE FROM users WHERE system_user_id = ?`, systemUserID)
	return err
}

// GetCookieForSystem 获取指定系统用户的网易云cookie
func GetCookieForSystem(systemUserID int) (string, error) {
	user, err := GetCurrentUserForSystem(systemUserID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", nil
	}
	return user.Cookie, nil
}

// IsLoggedInForSystem 检查指定系统用户是否已登录网易云
func IsLoggedInForSystem(systemUserID int) bool {
	user, err := GetCurrentUserForSystem(systemUserID)
	if err != nil || user == nil {
		return false
	}
	return !time.Now().After(user.CookieExpires)
}
