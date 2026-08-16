package db

import (
	"database/sql"
	"fmt"
)

// SyncPlaylist 同步歌单
type SyncPlaylist struct {
	ID           int
	PlaylistID   int
	PlaylistName string
	UserID       int
	Enabled      bool
	CreatedAt    sql.NullTime
}

// GetSyncPlaylists 获取用户的同步歌单列表
func GetSyncPlaylists(userID int) ([]SyncPlaylist, error) {
	rows, err := dbConn.Query(
		"SELECT id, playlist_id, playlist_name, user_id, enabled, created_at FROM sync_playlists WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []SyncPlaylist
	for rows.Next() {
		var p SyncPlaylist
		if err := rows.Scan(&p.ID, &p.PlaylistID, &p.PlaylistName, &p.UserID, &p.Enabled, &p.CreatedAt); err != nil {
			return nil, err
		}
		playlists = append(playlists, p)
	}
	return playlists, nil
}

// AddSyncPlaylist 添加同步歌单
func AddSyncPlaylist(playlistID int, playlistName string, userID int) error {
	_, err := dbConn.Exec(
		"INSERT OR REPLACE INTO sync_playlists (playlist_id, playlist_name, user_id, enabled) VALUES (?, ?, ?, 1)",
		playlistID, playlistName, userID,
	)
	return err
}

// RemoveSyncPlaylist 移除同步歌单
func RemoveSyncPlaylist(playlistID int, userID int) error {
	_, err := dbConn.Exec(
		"DELETE FROM sync_playlists WHERE playlist_id = ? AND user_id = ?",
		playlistID, userID,
	)
	return err
}

// ToggleSyncPlaylist 切换歌单同步状态
func ToggleSyncPlaylist(playlistID int, userID int, enabled bool) error {
	_, err := dbConn.Exec(
		"UPDATE sync_playlists SET enabled = ? WHERE playlist_id = ? AND user_id = ?",
		enabled, playlistID, userID,
	)
	return err
}

// GetEnabledSyncPlaylists 获取用户启用的同步歌单ID列表
func GetEnabledSyncPlaylists(userID int) ([]int, error) {
	rows, err := dbConn.Query(
		"SELECT playlist_id FROM sync_playlists WHERE user_id = ? AND enabled = 1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlistIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		playlistIDs = append(playlistIDs, id)
	}
	return playlistIDs, nil
}

// SyncPlaylistsExist 检查用户是否已配置同步歌单
func SyncPlaylistsExist(userID int) (bool, error) {
	var count int
	err := dbConn.QueryRow("SELECT COUNT(*) FROM sync_playlists WHERE user_id = ?", userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// BatchUpdateSyncPlaylists 批量更新同步歌单列表
func BatchUpdateSyncPlaylists(userID int, playlistIDs []int) error {
	// 先删除所有旧记录
	_, err := dbConn.Exec("DELETE FROM sync_playlists WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("删除旧同步歌单失败: %w", err)
	}

	// 插入新记录
	for _, playlistID := range playlistIDs {
		_, err := dbConn.Exec(
			"INSERT INTO sync_playlists (playlist_id, playlist_name, user_id, enabled) VALUES (?, '', ?, 1)",
			playlistID, userID,
		)
		if err != nil {
			return fmt.Errorf("插入同步歌单失败: %w", err)
		}
	}
	return nil
}
