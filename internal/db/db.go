package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var dbConn *sql.DB

func InitDB(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %w", err)
	}

	var err error
	dbConn, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	if err := dbConn.Ping(); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	if err := initTables(); err != nil {
		return fmt.Errorf("初始化数据库表失败: %w", err)
	}

	return nil
}

func CloseDB() error {
	if dbConn != nil {
		return dbConn.Close()
	}
	return nil
}

func GetDB() *sql.DB {
	return dbConn
}

func initTables() error {
	tables := []string{
		// 用户信息表
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE NOT NULL,
			nickname TEXT,
			avatar_url TEXT,
			cookie TEXT,
			cookie_expires DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 下载历史表
		`CREATE TABLE IF NOT EXISTS downloads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			song_id INTEGER NOT NULL,
			song_name TEXT NOT NULL,
			artist TEXT,
			album TEXT,
			quality TEXT,
			file_path TEXT,
			file_size INTEGER,
			status TEXT DEFAULT 'pending',
			error_msg TEXT,
			metadata_completed BOOLEAN DEFAULT 0,
			download_url TEXT DEFAULT '',
			total_size INTEGER DEFAULT 0,
			downloaded_size INTEGER DEFAULT 0,
			sub_dir TEXT DEFAULT '',
			playlist_id INTEGER DEFAULT 0,
			phase TEXT DEFAULT 'download',
			cover_downloaded BOOLEAN DEFAULT 0,
			lyrics_downloaded BOOLEAN DEFAULT 0,
			artist_completed BOOLEAN DEFAULT 0,
			id3_embedded BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 歌单表
		`CREATE TABLE IF NOT EXISTS playlists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			playlist_id INTEGER UNIQUE NOT NULL,
			name TEXT NOT NULL,
			creator_id INTEGER,
			track_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 设置表
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT NOT NULL,
			value TEXT,
			user_id INTEGER DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(key, user_id)
		)`,

		// 系统用户表（访问控制）
		`CREATE TABLE IF NOT EXISTS system_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login_at DATETIME,
			failed_attempts INTEGER DEFAULT 0,
			locked_until DATETIME
		)`,

		// 同步任务表
		`CREATE TABLE IF NOT EXISTS sync_tasks (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			playlist_id INTEGER DEFAULT 0,
			title TEXT DEFAULT '',
			status TEXT DEFAULT 'pending',
			current INTEGER DEFAULT 0,
			total INTEGER DEFAULT 0,
			current_file TEXT DEFAULT '',
			current_bytes INTEGER DEFAULT 0,
			total_bytes INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 扫盘记录表
		`CREATE TABLE IF NOT EXISTS scan_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			song_id INTEGER NOT NULL,
			song_name TEXT NOT NULL,
			artist TEXT,
			album TEXT,
			file_path TEXT,
			file_size INTEGER DEFAULT 0,
			mod_time DATETIME,
			song_downloaded BOOLEAN DEFAULT 0,
			lyrics_downloaded BOOLEAN DEFAULT 0,
			metadata_completed BOOLEAN DEFAULT 0,
			playlist_id INTEGER DEFAULT 0,
			sub_dir TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(song_id, sub_dir)
		)`,

	}

	for _, table := range tables {
		if _, err := dbConn.Exec(table); err != nil {
			return fmt.Errorf("执行 SQL 失败: %w", err)
		}
	}

	// 迁移：为已有表添加新字段
	migrations := []string{
		`ALTER TABLE downloads ADD COLUMN download_url TEXT DEFAULT ''`,
		`ALTER TABLE downloads ADD COLUMN total_size INTEGER DEFAULT 0`,
		`ALTER TABLE downloads ADD COLUMN downloaded_size INTEGER DEFAULT 0`,
		`ALTER TABLE downloads ADD COLUMN sub_dir TEXT DEFAULT ''`,
		`ALTER TABLE downloads ADD COLUMN playlist_id INTEGER DEFAULT 0`,
		`ALTER TABLE downloads ADD COLUMN phase TEXT DEFAULT 'download'`,
		`ALTER TABLE downloads ADD COLUMN cover_downloaded BOOLEAN DEFAULT 0`,
		`ALTER TABLE downloads ADD COLUMN lyrics_downloaded BOOLEAN DEFAULT 0`,
		`ALTER TABLE downloads ADD COLUMN artist_completed BOOLEAN DEFAULT 0`,
		`ALTER TABLE downloads ADD COLUMN id3_embedded BOOLEAN DEFAULT 0`,
		// 为 users 表添加 system_user_id 字段
		`ALTER TABLE users ADD COLUMN system_user_id INTEGER DEFAULT 0`,
	}

	// 执行迁移
	for _, m := range migrations {
		dbConn.Exec(m) // 忽略错误（字段已存在时会报错，这是预期的）
	}

	// 特殊迁移：settings 表需要重建以支持 user_id
	// 检查是否需要迁移 settings 表
	var hasUserID bool
	err := dbConn.QueryRow("SELECT COUNT(*) FROM pragma_table_info('settings') WHERE name='user_id'").Scan(&hasUserID)
	if err == nil && !hasUserID {
		// 备份旧数据
		dbConn.Exec(`CREATE TABLE IF NOT EXISTS settings_backup AS SELECT * FROM settings`)
		// 删除旧表
		dbConn.Exec(`DROP TABLE settings`)
		// 创建新表
		dbConn.Exec(`CREATE TABLE settings (
			key TEXT NOT NULL,
			value TEXT,
			user_id INTEGER DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(key, user_id)
		)`)
		// 恢复数据（user_id 默认为 0）
		dbConn.Exec(`INSERT INTO settings (key, value, user_id, updated_at) SELECT key, value, 0, updated_at FROM settings_backup`)
		dbConn.Exec(`DROP TABLE settings_backup`)
	}

	// 创建索引（在迁移之后，确保所有字段都存在）
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_downloads_song_id ON downloads(song_id)`,
		`CREATE INDEX IF NOT EXISTS idx_downloads_status ON downloads(status)`,
		`CREATE INDEX IF NOT EXISTS idx_downloads_created_at ON downloads(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_downloads_phase ON downloads(phase)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_tasks_status ON sync_tasks(status)`,
	}

	for _, idx := range indexes {
		dbConn.Exec(idx) // 忽略错误
	}

	return nil
}
