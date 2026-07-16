// Package handler 定义 Handler 结构体和公共工具方法喵
package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"oidc-proxy/internal/crypto"
	"oidc-proxy/internal/model"
	"oidc-proxy/internal/rainbow"
	"oidc-proxy/internal/store"
)

// Handler 包含所有 HTTP 处理器共享的依赖喵
type Handler struct {
	store      *store.SQLiteStore           // SQLite 存储（设置+客户端+日志）
	keyManager *crypto.KeyManager           // RSA 密钥管理器
	rainClient *rainbow.Client              // 彩虹 API 客户端
	authStates map[string]model.AuthRequest // 彩虹授权 state 映射
	stateMux   sync.RWMutex                 // state 映射锁
	startTime  time.Time                    // 服务启动时间
}

// NewHandler 创建新的 Handler 实例喵
func NewHandler(s *store.SQLiteStore, km *crypto.KeyManager, rc *rainbow.Client) *Handler {
	return &Handler{
		store:      s,
		keyManager: km,
		rainClient: rc,
		authStates: make(map[string]model.AuthRequest),
		startTime:  time.Now(),
	}
}

// GetIssuer 从数据库读取 issuer 喵
func (h *Handler) GetIssuer() string {
	return h.store.GetSetting("issuer")
}

// logf 统一的日志输出方法喵
func (h *Handler) logf(level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[%s] %s", level, msg)
}

// writeJSON 写入 JSON 响应喵
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[ERROR] JSON 编码失败: %v", err)
	}
}
