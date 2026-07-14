package model

import "time"

type Song struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int    `json:"duration"` // 毫秒
	PicURL   string `json:"pic_url"`
	Lyric    string `json:"lyric,omitempty"`
}

type SearchResult struct {
	Songs []Song `json:"songs"`
	Total int    `json:"total"`
}

type Download struct {
	ID                int       `json:"id"`
	SongID            int       `json:"song_id"`
	SongName          string    `json:"song_name"`
	Artist            string    `json:"artist"`
	Album             string    `json:"album"`
	Quality           string    `json:"quality"`
	FilePath          string    `json:"file_path"`
	FileSize          int64     `json:"file_size"`
	Status            string    `json:"status"` // pending, downloading, completed, failed
	ErrorMsg          string    `json:"error_msg"`
	MetadataCompleted bool      `json:"metadata_completed"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Playlist struct {
	ID          int       `json:"id"`
	PlaylistID  int       `json:"playlist_id"`
	Name        string    `json:"name"`
	CreatorID   int       `json:"creator_id"`
	TrackCount  int       `json:"track_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type User struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	Nickname       string    `json:"nickname"`
	AvatarURL      string    `json:"avatar_url"`
	Cookie         string    `json:"-"`
	CookieExpires  time.Time `json:"cookie_expires"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DownloadHistory struct {
	ID               int       `json:"id"`
	SongID           int       `json:"song_id"`
	SongName         string    `json:"song_name"`
	Artist           string    `json:"artist"`
	Album            string    `json:"album"`
	Quality          string    `json:"quality"`
	Status           string    `json:"status"`
	ErrorMsg         string    `json:"error_msg"`
	FilePath         string    `json:"file_path"`
	FileSize         int64     `json:"file_size"`
	MetadataCompleted bool     `json:"metadata_completed"`
	DownloadURL      string    `json:"download_url"`
	TotalSize        int64     `json:"total_size"`
	DownloadedSize   int64     `json:"downloaded_size"`
	SubDir           string    `json:"sub_dir"`
	PlaylistID       int       `json:"playlist_id"`
	Phase            string    `json:"phase"`
	CoverDownloaded  bool      `json:"cover_downloaded"`
	LyricsDownloaded bool      `json:"lyrics_downloaded"`
	ArtistCompleted  bool      `json:"artist_completed"`
	ID3Embedded      bool      `json:"id3_embedded"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type SyncTask struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	PlaylistID   int       `json:"playlist_id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	Current      int       `json:"current"`
	Total        int       `json:"total"`
	CurrentFile  string    `json:"current_file"`
	CurrentBytes int64     `json:"current_bytes"`
	TotalBytes   int64     `json:"total_bytes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ScanRecord struct {
	ID                int       `json:"id"`
	SongID            int       `json:"song_id"`
	SongName          string    `json:"song_name"`
	Artist            string    `json:"artist"`
	Album             string    `json:"album"`
	FilePath          string    `json:"file_path"`
	FileSize          int64     `json:"file_size"`
	ModTime           time.Time `json:"mod_time"`
	SongDownloaded    bool      `json:"song_downloaded"`
	LyricsDownloaded  bool      `json:"lyrics_downloaded"`
	MetadataCompleted bool      `json:"metadata_completed"`
	PlaylistID        int       `json:"playlist_id"`
	SubDir            string    `json:"sub_dir"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
