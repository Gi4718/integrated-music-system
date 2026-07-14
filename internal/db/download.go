package db

import (
	"database/sql"
	"endfield-music/internal/model"
	"time"
)

func SaveDownloadHistory(d *model.DownloadHistory) error {
	query := `INSERT INTO downloads (song_id, song_name, artist, album, quality, status, sub_dir, playlist_id, phase, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'download', ?, ?)`
	_, err := dbConn.Exec(query, d.SongID, d.SongName, d.Artist, d.Album, d.Quality, d.Status, d.SubDir, d.PlaylistID, time.Now(), time.Now())
	return err
}

func UpdateDownloadStatus(songID int, status string) error {
	query := `UPDATE downloads SET status = ?, updated_at = ? WHERE id = (SELECT id FROM downloads WHERE song_id = ? ORDER BY id DESC LIMIT 1)`
	_, err := dbConn.Exec(query, status, time.Now(), songID)
	return err
}

func UpdateDownloadProgress(songID int, downloadedSize, totalSize int64) error {
	query := `UPDATE downloads SET downloaded_size = ?, total_size = ?, updated_at = ?
			  WHERE id = (SELECT id FROM downloads WHERE song_id = ? ORDER BY id DESC LIMIT 1)`
	_, err := dbConn.Exec(query, downloadedSize, totalSize, time.Now(), songID)
	return err
}

func UpdateDownloadPhase(songID int, phase string) error {
	query := `UPDATE downloads SET phase = ?, updated_at = ?
			  WHERE id = (SELECT id FROM downloads WHERE song_id = ? ORDER BY id DESC LIMIT 1)`
	_, err := dbConn.Exec(query, phase, time.Now(), songID)
	return err
}

func UpdateDownloadFilePath(songID int, filePath string, fileSize int64) error {
	query := `UPDATE downloads SET file_path = ?, file_size = ?, downloaded_size = ?, total_size = ?, phase = 'metadata', updated_at = ?
			  WHERE id = (SELECT id FROM downloads WHERE song_id = ? ORDER BY id DESC LIMIT 1)`
	_, err := dbConn.Exec(query, filePath, fileSize, fileSize, fileSize, time.Now(), songID)
	return err
}

func UpdateMetadataProgress(songID int, field string) error {
	var query string
	switch field {
	case "cover":
		query = `UPDATE downloads SET cover_downloaded = 1, updated_at = ? WHERE id = (SELECT id FROM downloads WHERE song_id = ? ORDER BY id DESC LIMIT 1)`
	case "lyrics":
		query = `UPDATE downloads SET lyrics_downloaded = 1, updated_at = ? WHERE id = (SELECT id FROM downloads WHERE song_id = ? ORDER BY id DESC LIMIT 1)`
	case "artist":
		query = `UPDATE downloads SET artist_completed = 1, updated_at = ? WHERE id = (SELECT id FROM downloads WHERE song_id = ? ORDER BY id DESC LIMIT 1)`
	case "id3":
		query = `UPDATE downloads SET id3_embedded = 1, updated_at = ? WHERE id = (SELECT id FROM downloads WHERE song_id = ? ORDER BY id DESC LIMIT 1)`
	case "completed":
		query = `UPDATE downloads SET metadata_completed = 1, phase = 'completed', updated_at = ? WHERE id = (SELECT id FROM downloads WHERE song_id = ? ORDER BY id DESC LIMIT 1)`
	default:
		return nil
	}
	_, err := dbConn.Exec(query, time.Now(), songID)
	return err
}

func GetPendingDownloads() ([]model.DownloadHistory, error) {
	query := `SELECT id, song_id, song_name, COALESCE(artist,''), COALESCE(album,''), COALESCE(quality,''), status,
			  COALESCE(error_msg,''), COALESCE(file_path,''), COALESCE(file_size, 0),
			  metadata_completed, COALESCE(download_url,''), COALESCE(total_size, 0), COALESCE(downloaded_size, 0),
			  COALESCE(sub_dir,''), playlist_id, COALESCE(phase,'download'),
			  cover_downloaded, lyrics_downloaded, artist_completed, id3_embedded, created_at, updated_at
			  FROM downloads WHERE phase != 'completed' ORDER BY created_at ASC`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.DownloadHistory
	for rows.Next() {
		var d model.DownloadHistory
		err := rows.Scan(&d.ID, &d.SongID, &d.SongName, &d.Artist, &d.Album, &d.Quality, &d.Status, &d.ErrorMsg,
			&d.FilePath, &d.FileSize, &d.MetadataCompleted, &d.DownloadURL, &d.TotalSize, &d.DownloadedSize,
			&d.SubDir, &d.PlaylistID, &d.Phase, &d.CoverDownloaded, &d.LyricsDownloaded, &d.ArtistCompleted,
			&d.ID3Embedded, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

func GetDownloadHistory() ([]model.DownloadHistory, error) {
	query := `SELECT id, song_id, song_name, COALESCE(artist,''), COALESCE(album,''), COALESCE(quality,''), status,
			  COALESCE(error_msg,''), COALESCE(file_path,''), COALESCE(file_size, 0),
			  metadata_completed, COALESCE(download_url,''), COALESCE(total_size, 0), COALESCE(downloaded_size, 0),
			  COALESCE(sub_dir,''), playlist_id, COALESCE(phase,'download'),
			  cover_downloaded, lyrics_downloaded, artist_completed, id3_embedded, created_at, updated_at
			  FROM downloads ORDER BY created_at DESC`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.DownloadHistory
	for rows.Next() {
		var d model.DownloadHistory
		err := rows.Scan(&d.ID, &d.SongID, &d.SongName, &d.Artist, &d.Album, &d.Quality, &d.Status, &d.ErrorMsg,
			&d.FilePath, &d.FileSize, &d.MetadataCompleted, &d.DownloadURL, &d.TotalSize, &d.DownloadedSize,
			&d.SubDir, &d.PlaylistID, &d.Phase, &d.CoverDownloaded, &d.LyricsDownloaded, &d.ArtistCompleted,
			&d.ID3Embedded, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

func GetDownloadBySongID(songID int) (*model.DownloadHistory, error) {
	query := `SELECT id, song_id, song_name, COALESCE(artist,''), COALESCE(album,''), COALESCE(quality,''), status,
			  COALESCE(error_msg,''), COALESCE(file_path,''), COALESCE(file_size, 0),
			  metadata_completed, COALESCE(download_url,''), COALESCE(total_size, 0), COALESCE(downloaded_size, 0),
			  COALESCE(sub_dir,''), playlist_id, COALESCE(phase,'download'),
			  cover_downloaded, lyrics_downloaded, artist_completed, id3_embedded, created_at, updated_at
			  FROM downloads WHERE song_id = ? ORDER BY created_at DESC LIMIT 1`
	row := dbConn.QueryRow(query, songID)
	var d model.DownloadHistory
	err := row.Scan(&d.ID, &d.SongID, &d.SongName, &d.Artist, &d.Album, &d.Quality, &d.Status, &d.ErrorMsg,
		&d.FilePath, &d.FileSize, &d.MetadataCompleted, &d.DownloadURL, &d.TotalSize, &d.DownloadedSize,
		&d.SubDir, &d.PlaylistID, &d.Phase, &d.CoverDownloaded, &d.LyricsDownloaded, &d.ArtistCompleted,
		&d.ID3Embedded, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func SaveSyncTask(t *model.SyncTask) error {
	query := `INSERT INTO sync_tasks (id, type, playlist_id, title, status, current, total, current_file, current_bytes, total_bytes, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := dbConn.Exec(query, t.ID, t.Type, t.PlaylistID, t.Title, t.Status, t.Current, t.Total, t.CurrentFile, t.CurrentBytes, t.TotalBytes, t.CreatedAt, t.UpdatedAt)
	return err
}

func UpdateSyncTask(t *model.SyncTask) error {
	query := `UPDATE sync_tasks SET status = ?, current = ?, total = ?, current_file = ?, current_bytes = ?, total_bytes = ?, updated_at = ? WHERE id = ?`
	_, err := dbConn.Exec(query, t.Status, t.Current, t.Total, t.CurrentFile, t.CurrentBytes, t.TotalBytes, time.Now(), t.ID)
	return err
}

func GetActiveSyncTasks() ([]model.SyncTask, error) {
	query := `SELECT id, type, playlist_id, title, status, current, total, current_file, current_bytes, total_bytes, created_at, updated_at
			  FROM sync_tasks WHERE status IN ('pending', 'running') ORDER BY created_at ASC`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.SyncTask
	for rows.Next() {
		var t model.SyncTask
		err := rows.Scan(&t.ID, &t.Type, &t.PlaylistID, &t.Title, &t.Status, &t.Current, &t.Total, &t.CurrentFile, &t.CurrentBytes, &t.TotalBytes, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func GetPlaylist(playlistID int) (*model.Playlist, error) {
	query := `SELECT id, playlist_id, name, creator_id, track_count, created_at FROM playlists WHERE playlist_id = ?`
	row := dbConn.QueryRow(query, playlistID)
	var p model.Playlist
	err := row.Scan(&p.ID, &p.PlaylistID, &p.Name, &p.CreatorID, &p.TrackCount, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func SavePlaylist(p *model.Playlist) error {
	query := `INSERT INTO playlists (playlist_id, name, creator_id, track_count, created_at)
			  VALUES (?, ?, ?, ?, ?)
			  ON CONFLICT(playlist_id) DO UPDATE SET
			  name = excluded.name, track_count = excluded.track_count`
	_, err := dbConn.Exec(query, p.PlaylistID, p.Name, p.CreatorID, p.TrackCount, time.Now())
	return err
}

// 扫盘记录相关
func SaveScanRecord(songID int, songName, artist, album, filePath string, fileSize int64, modTime time.Time, songDownloaded, lyricsDownloaded, metadataCompleted bool, playlistID int, subDir string) error {
	query := `INSERT INTO scan_records (song_id, song_name, artist, album, file_path, file_size, mod_time, song_downloaded, lyrics_downloaded, metadata_completed, playlist_id, sub_dir, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			  ON CONFLICT(song_id, sub_dir) DO UPDATE SET
			  song_name = excluded.song_name,
			  artist = excluded.artist,
			  album = excluded.album,
			  file_path = excluded.file_path,
			  file_size = excluded.file_size,
			  mod_time = excluded.mod_time,
			  song_downloaded = excluded.song_downloaded,
			  lyrics_downloaded = excluded.lyrics_downloaded,
			  metadata_completed = excluded.metadata_completed,
			  updated_at = excluded.updated_at`
	_, err := dbConn.Exec(query, songID, songName, artist, album, filePath, fileSize, modTime, songDownloaded, lyricsDownloaded, metadataCompleted, playlistID, subDir, time.Now(), time.Now())
	return err
}

func GetScanRecordsBySubDir(subDir string) ([]model.ScanRecord, error) {
	query := `SELECT id, song_id, song_name, artist, album, file_path, file_size, mod_time, song_downloaded, lyrics_downloaded, metadata_completed, playlist_id, sub_dir, created_at, updated_at
			  FROM scan_records WHERE sub_dir = ? ORDER BY created_at DESC`
	rows, err := dbConn.Query(query, subDir)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.ScanRecord
	for rows.Next() {
		var r model.ScanRecord
		err := rows.Scan(&r.ID, &r.SongID, &r.SongName, &r.Artist, &r.Album, &r.FilePath, &r.FileSize, &r.ModTime, &r.SongDownloaded, &r.LyricsDownloaded, &r.MetadataCompleted, &r.PlaylistID, &r.SubDir, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func GetScanRecordBySongIDAndSubDir(songID int, subDir string) (*model.ScanRecord, error) {
	query := `SELECT id, song_id, song_name, artist, album, file_path, file_size, mod_time, song_downloaded, lyrics_downloaded, metadata_completed, playlist_id, sub_dir, created_at, updated_at
			  FROM scan_records WHERE song_id = ? AND sub_dir = ? LIMIT 1`
	row := dbConn.QueryRow(query, songID, subDir)
	var r model.ScanRecord
	err := row.Scan(&r.ID, &r.SongID, &r.SongName, &r.Artist, &r.Album, &r.FilePath, &r.FileSize, &r.ModTime, &r.SongDownloaded, &r.LyricsDownloaded, &r.MetadataCompleted, &r.PlaylistID, &r.SubDir, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// 查找任意已下载的歌曲（跨歌单）
func GetAnyDownloadedSong(songID int) (*model.DownloadHistory, error) {
	query := `SELECT id, song_id, song_name, COALESCE(artist,''), COALESCE(album,''), COALESCE(quality,''), status,
			  COALESCE(error_msg,''), COALESCE(file_path,''), COALESCE(file_size, 0),
			  metadata_completed, COALESCE(download_url,''), COALESCE(total_size, 0), COALESCE(downloaded_size, 0),
			  COALESCE(sub_dir,''), playlist_id, COALESCE(phase,'download'),
			  cover_downloaded, lyrics_downloaded, artist_completed, id3_embedded, created_at, updated_at
			  FROM downloads WHERE song_id = ? AND status = 'completed' AND file_path != '' ORDER BY updated_at DESC LIMIT 1`
	row := dbConn.QueryRow(query, songID)
	var d model.DownloadHistory
	err := row.Scan(&d.ID, &d.SongID, &d.SongName, &d.Artist, &d.Album, &d.Quality, &d.Status, &d.ErrorMsg,
		&d.FilePath, &d.FileSize, &d.MetadataCompleted, &d.DownloadURL, &d.TotalSize, &d.DownloadedSize,
		&d.SubDir, &d.PlaylistID, &d.Phase, &d.CoverDownloaded, &d.LyricsDownloaded, &d.ArtistCompleted,
		&d.ID3Embedded, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}
