package sqllite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// MusicAudioSource 音乐音频源信息
type MusicAudioSource struct {
	ID        int64  `json:"id"`
	SongID    int64  `json:"song_id"`
	Source    string `json:"source"`
	AudioURL  string `json:"audio_url"`
	PicURL    string `json:"pic_url"`
	LrcURL    string `json:"lrc_url"`
	FileSize  int64  `json:"file_size"`
	CreatedAt string `json:"created_at"`
}

// musicDB 单例句柄，对应 music_cache.db。
var musicDB *sql.DB

func getMusicDB() *sql.DB {
	if musicDB != nil {
		return musicDB
	}
	db := getOrOpen("music_cache.db")
	if db == nil {
		return nil
	}
	execSchema(db, musicSchema, "创建 music 表失败:")
	musicDB = db
	return musicDB
}

var musicSchema = []string{
	`CREATE TABLE IF NOT EXISTS music_search_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		keyword TEXT NOT NULL UNIQUE,
		results TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS music_songs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		song_id INTEGER NOT NULL,
		source TEXT NOT NULL,
		name TEXT NOT NULL,
		artist TEXT NOT NULL,
		album TEXT NOT NULL,
		duration INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(song_id, source)
	)`,
	`CREATE TABLE IF NOT EXISTS music_audio_sources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		song_id INTEGER NOT NULL,
		source TEXT NOT NULL,
		audio_url TEXT,
		pic_url TEXT,
		lrc_url TEXT,
		file_size INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(song_id, source)
	)`,
}

// SearchMusicCache 搜索音乐缓存；命中返回 (resultsJSON, true)，未命中返回 ("", false)。
func SearchMusicCache(keyword string) (string, bool) {
	db := getMusicDB()
	if db == nil {
		return "", false
	}
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return "", false
	}
	var resultsJSON string
	err := db.QueryRow("SELECT results FROM music_search_history WHERE keyword = ?", keyword).Scan(&resultsJSON)
	if err != nil {
		return "", false
	}
	return resultsJSON, true
}

// SaveSearchResults 保存搜索结果到缓存，resultsJSON 是搜索结果数组的 JSON 字符串。
func SaveSearchResults(keyword, resultsJSON string) error {
	db := getMusicDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return fmt.Errorf("关键词为空")
	}
	if _, err := db.Exec(`
		INSERT INTO music_search_history (keyword, results, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(keyword) DO UPDATE SET results = excluded.results, updated_at = excluded.updated_at
	`, keyword, resultsJSON); err != nil {
		return err
	}

	var songs []map[string]any
	if err := json.Unmarshal([]byte(resultsJSON), &songs); err != nil {
		return nil
	}
	for _, song := range songs {
		songID := int64(0)
		if v, ok := song["id"].(float64); ok {
			songID = int64(v)
		}
		source, _ := song["source"].(string)
		name, _ := song["name"].(string)
		artist, _ := song["artist"].(string)
		album, _ := song["album"].(string)
		duration := 0
		if v, ok := song["duration"].(float64); ok {
			duration = int(v)
		}
		if songID > 0 && source != "" {
			if _, err := db.Exec(`
				INSERT INTO music_songs (song_id, source, name, artist, album, duration)
				VALUES (?, ?, ?, ?, ?, ?)
				ON CONFLICT(song_id, source) DO UPDATE SET name = excluded.name, artist = excluded.artist, album = excluded.album, duration = excluded.duration
			`, songID, source, name, artist, album, duration); err != nil {
				fmt.Println("保存歌曲失败:", err)
			}
		}
	}
	return nil
}

// SaveAudioSource 保存/更新一条音频源记录。
func SaveAudioSource(songID int64, source, audioURL, picURL, lrcURL string, fileSize int64) error {
	db := getMusicDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := db.Exec(`
		INSERT INTO music_audio_sources (song_id, source, audio_url, pic_url, lrc_url, file_size)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(song_id, source) DO UPDATE SET audio_url = excluded.audio_url, pic_url = excluded.pic_url, lrc_url = excluded.lrc_url, file_size = excluded.file_size
	`, songID, source, audioURL, picURL, lrcURL, fileSize)
	return err
}

// GetAudioSource 获取 (songID, source) 对应的音频源记录。
func GetAudioSource(songID int64, source string) (*MusicAudioSource, error) {
	db := getMusicDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var out MusicAudioSource
	err := db.QueryRow(`
		SELECT id, song_id, source, audio_url, pic_url, lrc_url, file_size, created_at
		FROM music_audio_sources
		WHERE song_id = ? AND source = ?
	`, songID, source).Scan(
		&out.ID, &out.SongID, &out.Source,
		&out.AudioURL, &out.PicURL, &out.LrcURL,
		&out.FileSize, &out.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &out, nil
}