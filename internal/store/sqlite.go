// Package store 提供基于 SQLite 的持久化存储喵
// 替代原来的 config.yaml + data.json + sync.Map 喵
package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"oidc-proxy/internal/model"
)

// SQLiteStore SQLite 数据库存储，实现 Store 接口喵
type SQLiteStore struct {
	db       *sql.DB  // 数据库连接
	codes    sync.Map // key: code(string), value: model.TempCodeData（临时code放内存）
	sessions sync.Map // key: token(string), value: expires(time.Time)（session放内存）
}

// NewSQLiteStore 创建 SQLite 存储实例，自动建表喵
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败喵: %w", err)
	}

	// 连接池配置喵
	db.SetMaxOpenConns(1) // SQLite 单写，避免 busy
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	s := &SQLiteStore{db: db}

	// 自动建表喵
	if err := s.initTables(); err != nil {
		return nil, fmt.Errorf("初始化数据库表失败喵: %w", err)
	}

	// 插入默认设置（仅首次）喵
	if err := s.seedDefaults(); err != nil {
		return nil, fmt.Errorf("写入默认设置失败喵: %w", err)
	}

	// 启动过期数据清理喵
	go s.cleanupLoop()

	log.Printf("[INFO] SQLite 数据库初始化成功: %s", dbPath)
	return s, nil
}

// ────────── 建表 ──────────

func (s *SQLiteStore) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS clients (
			client_id TEXT PRIMARY KEY,
			client_secret TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			redirect_uri TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			time TEXT NOT NULL DEFAULT '',
			client TEXT NOT NULL DEFAULT '',
			platform TEXT NOT NULL DEFAULT '',
			nickname TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_time ON logs(time DESC)`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("建表失败: %s, err=%w", q, err)
		}
	}
	return nil
}

// ────────── 默认值 ──────────

func (s *SQLiteStore) seedDefaults() error {
	defaults := map[string]string{
		"issuer":              "https://oidc.example.com",
		"listen":              ":8443",
		"admin_username":      "admin",
		"admin_password":      "admin123",
		"rainbow_app_id":      "",
		"rainbow_app_key":     "",
		"rainbow_connect_url": "https://login.mapay.cn/connect.php",
	}
	for k, v := range defaults {
		_, err := s.db.Exec("INSERT OR IGNORE INTO settings(key,value) VALUES(?,?)", k, v)
		if err != nil {
			return fmt.Errorf("插入默认设置失败: key=%s, err=%w", k, err)
		}
	}
	return nil
}

// ────────── 设置读写 ──────────

// GetSetting 读取单个设置喵
func (s *SQLiteStore) GetSetting(key string) string {
	var val string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key=?", key).Scan(&val)
	if err != nil {
		return ""
	}
	return val
}

// SetSetting 写入单个设置喵
func (s *SQLiteStore) SetSetting(key, value string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO settings(key,value) VALUES(?,?)", key, value)
	return err
}

// GetAllSettings 获取所有设置（返回 map）喵
func (s *SQLiteStore) GetAllSettings() map[string]string {
	rows, err := s.db.Query("SELECT key, value FROM settings ORDER BY key")
	if err != nil {
		return map[string]string{}
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		result[k] = v
	}
	return result
}

// BuildRainConfig 从 settings 构建 RainConfig 喵
func (s *SQLiteStore) BuildRainConfig() model.RainConfig {
	return model.RainConfig{
		AppID:      s.GetSetting("rainbow_app_id"),
		AppKey:     s.GetSetting("rainbow_app_key"),
		ConnectURL: s.GetSetting("rainbow_connect_url"),
	}
}

// ────────── 客户端 CRUD ──────────

// ListClients 获取所有客户端喵
func (s *SQLiteStore) ListClients() ([]model.OIDCClient, error) {
	rows, err := s.db.Query("SELECT client_id, client_secret, name, redirect_uri, created_at FROM clients ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []model.OIDCClient
	for rows.Next() {
		var c model.OIDCClient
		rows.Scan(&c.ClientID, &c.ClientSecret, &c.Name, &c.RedirectURI, &c.CreatedAt)
		clients = append(clients, c)
	}
	if clients == nil {
		clients = []model.OIDCClient{}
	}
	return clients, nil
}

// CreateClient 创建新客户端喵
func (s *SQLiteStore) CreateClient(name, redirectURI string) (*model.OIDCClient, error) {
	c := model.OIDCClient{
		Name:         name,
		ClientID:     "client_" + generateRandomID(12),
		ClientSecret: "secret_" + generateRandomID(16),
		RedirectURI:  redirectURI,
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}
	_, err := s.db.Exec(
		"INSERT INTO clients(client_id,client_secret,name,redirect_uri,created_at) VALUES(?,?,?,?,?)",
		c.ClientID, c.ClientSecret, c.Name, c.RedirectURI, c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// DeleteClient 删除客户端喵
func (s *SQLiteStore) DeleteClient(clientID string) error {
	_, err := s.db.Exec("DELETE FROM clients WHERE client_id=?", clientID)
	return err
}

// FindClient 根据 client_id 查找客户端喵
func (s *SQLiteStore) FindClient(clientID string) *model.OIDCClient {
	var c model.OIDCClient
	err := s.db.QueryRow(
		"SELECT client_id, client_secret, name, redirect_uri, created_at FROM clients WHERE client_id=?",
		clientID,
	).Scan(&c.ClientID, &c.ClientSecret, &c.Name, &c.RedirectURI, &c.CreatedAt)
	if err != nil {
		return nil
	}
	return &c
}

// ────────── 登录日志 ──────────

// AddLoginLog 添加登录日志，只保留最近 50 条喵
func (s *SQLiteStore) AddLoginLog(log model.LoginLog) error {
	_, err := s.db.Exec("INSERT INTO logs(time,client,platform,nickname) VALUES(?,?,?,?)",
		log.Time, log.Client, log.Platform, log.Nickname)
	if err != nil {
		return err
	}
	// 只保留最近 50 条喵
	_, _ = s.db.Exec("DELETE FROM logs WHERE id NOT IN (SELECT id FROM logs ORDER BY id DESC LIMIT 50)")
	return nil
}

// ListLogs 获取最近登录日志喵
func (s *SQLiteStore) ListLogs() ([]model.LoginLog, error) {
	rows, err := s.db.Query("SELECT time, client, platform, nickname FROM logs ORDER BY id DESC LIMIT 50")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.LoginLog
	for rows.Next() {
		var l model.LoginLog
		rows.Scan(&l.Time, &l.Client, &l.Platform, &l.Nickname)
		logs = append(logs, l)
	}
	if logs == nil {
		logs = []model.LoginLog{}
	}
	return logs, nil
}

// ────────── 临时 code（内存） ──────────

// SetTempCode 保存临时 code 喵
func (s *SQLiteStore) SetTempCode(code string, data model.TempCodeData) {
	s.codes.Store(code, data)
}

// GetTempCode 获取临时 code，过期返回 nil 喵
func (s *SQLiteStore) GetTempCode(code string) *model.TempCodeData {
	val, ok := s.codes.Load(code)
	if !ok {
		return nil
	}
	d := val.(model.TempCodeData)
	if time.Now().After(d.Expires) {
		s.codes.Delete(code)
		return nil
	}
	return &d
}

// DeleteTempCode 删除临时 code 喵
func (s *SQLiteStore) DeleteTempCode(code string) {
	s.codes.Delete(code)
}

// ────────── 管理员 session（内存） ──────────

// SetSession 保存管理员 session 喵
func (s *SQLiteStore) SetSession(token string, expires time.Time) {
	s.sessions.Store(token, expires)
}

// IsSessionValid 校验 session 喵
func (s *SQLiteStore) IsSessionValid(token string) bool {
	val, ok := s.sessions.Load(token)
	if !ok {
		return false
	}
	if time.Now().After(val.(time.Time)) {
		s.sessions.Delete(token)
		return false
	}
	return true
}

// DeleteSession 删除 session 喵
func (s *SQLiteStore) DeleteSession(token string) {
	s.sessions.Delete(token)
}

// ────────── 清理 ──────────

func (s *SQLiteStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.codes.Range(func(k, v any) bool {
			if now.After(v.(model.TempCodeData).Expires) {
				s.codes.Delete(k)
			}
			return true
		})
		s.sessions.Range(func(k, v any) bool {
			if now.After(v.(time.Time)) {
				s.sessions.Delete(k)
			}
			return true
		})
	}
}

// Close 关闭数据库连接喵
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// ────────── 工具 ──────────

func generateRandomID(bytes int) string {
	b := make([]byte, bytes)
	rand.Read(b)
	return hex.EncodeToString(b)
}
