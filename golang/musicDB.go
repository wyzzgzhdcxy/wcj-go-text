package golang

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	db   *sql.DB
	once sync.Once
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

// GetDB 获取数据库实例
func GetDB() *sql.DB {
	once.Do(func() {
		initDB()
	})
	return db
}

func initDB() {
	// 获取应用数据目录
	appDir := getAppDataDir()
	dbPath := filepath.Join(appDir, "music_cache.db")

	// 确保目录存在
	os.MkdirAll(appDir, 0755)

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Println("打开数据库失败:", err)
		return
	}

	// 测试连接
	if err = db.Ping(); err != nil {
		fmt.Println("数据库连接失败:", err)
		return
	}

	// 创建表
	createTables()
}

func createTables() {
	// 搜索历史表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS music_search_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			keyword TEXT NOT NULL UNIQUE,
			results TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Println("创建搜索历史表失败:", err)
	}

	// 歌曲表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS music_songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			song_id INTEGER NOT NULL,
			source TEXT NOT NULL,
			name TEXT NOT NULL,
			artist TEXT NOT NULL,
			album TEXT NOT NULL,
			duration INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(song_id, source)
		)
	`)
	if err != nil {
		fmt.Println("创建歌曲表失败:", err)
	}

	// 音频源表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS music_audio_sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			song_id INTEGER NOT NULL,
			source TEXT NOT NULL,
			audio_url TEXT,
			pic_url TEXT,
			lrc_url TEXT,
			file_size INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(song_id, source)
		)
	`)
	if err != nil {
		fmt.Println("创建音频源表失败:", err)
	}
}

func getAppDataDir() string {
	// Windows 下优先使用 %LOCALAPPDATA% (%AppData%\Local)，与原硬编码路径语义一致；
	// 缺失时回退到用户主目录或临时目录，避免换用户/换机器后写错位置。
	dir := os.Getenv("LOCALAPPDATA")
	if dir == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			dir = filepath.Join(home, "AppData", "Local")
		} else {
			dir = os.TempDir()
		}
	}
	return filepath.Join(dir, "wtools", "data")
}

// SearchMusicCache 搜索音乐（先查缓存），返回JSON原始数据
func SearchMusicCache(keyword string) (string, bool) {
	GetDB() // 确保数据库已初始化
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
		return "", false // 缓存未命中
	}

	return resultsJSON, true // 缓存命中
}

// SaveSearchResults 保存搜索结果到缓存，传入JSON字符串
func SaveSearchResults(keyword string, resultsJSON string) error {
	GetDB() // 确保数据库已初始化
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return fmt.Errorf("关键词为空")
	}

	// 插入或更新搜索历史
	_, err := db.Exec(`
		INSERT INTO music_search_history (keyword, results, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(keyword) DO UPDATE SET results = excluded.results, updated_at = excluded.updated_at
	`, keyword, resultsJSON)
	if err != nil {
		return err
	}

	// 解析JSON并保存歌曲信息
	var songs []map[string]any
	if err := json.Unmarshal([]byte(resultsJSON), &songs); err != nil {
		return nil // JSON解析失败不影响主流程
	}

	for _, song := range songs {
		songID := int64(0)
		if v, ok := song["id"].(float64); ok {
			songID = int64(v)
		}
		source := ""
		if v, ok := song["source"].(string); ok {
			source = v
		}
		name := ""
		if v, ok := song["name"].(string); ok {
			name = v
		}
		artist := ""
		if v, ok := song["artist"].(string); ok {
			artist = v
		}
		album := ""
		if v, ok := song["album"].(string); ok {
			album = v
		}
		duration := 0
		if v, ok := song["duration"].(float64); ok {
			duration = int(v)
		}

		if songID > 0 && source != "" {
			_, err = db.Exec(`
				INSERT INTO music_songs (song_id, source, name, artist, album, duration)
				VALUES (?, ?, ?, ?, ?, ?)
				ON CONFLICT(song_id, source) DO UPDATE SET name = excluded.name, artist = excluded.artist, album = excluded.album, duration = excluded.duration
			`, songID, source, name, artist, album, duration)
			if err != nil {
				fmt.Println("保存歌曲失败:", err)
			}
		}
	}

	return nil
}

// SaveAudioSource 保存音频源信息
func SaveAudioSource(songID int64, source string, audioURL, picURL, lrcURL string, fileSize int64) error {
	GetDB() // 确保数据库已初始化
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

// GetAudioSource 获取音频源信息
func GetAudioSource(songID int64, source string) (*MusicAudioSource, error) {
	GetDB() // 确保数据库已初始化
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var audioSource MusicAudioSource
	err := db.QueryRow(`
		SELECT id, song_id, source, audio_url, pic_url, lrc_url, file_size, created_at
		FROM music_audio_sources
		WHERE song_id = ? AND source = ?
	`, songID, source).Scan(
		&audioSource.ID, &audioSource.SongID, &audioSource.Source,
		&audioSource.AudioURL, &audioSource.PicURL, &audioSource.LrcURL,
		&audioSource.FileSize, &audioSource.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &audioSource, nil
}

// Close 关闭数据库
func Close() {
	if db != nil {
		db.Close()
	}
}