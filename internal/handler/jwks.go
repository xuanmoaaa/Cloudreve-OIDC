// Package handler 提供 JWKS 端点处理喵
package handler

import (
	"net/http"
)

// HandleJWKS 处理 GET /.well-known/jwks.json 请求喵
// 返回 RSA 公钥，用于验证 id_token 签名喵
func (h *Handler) HandleJWKS(w http.ResponseWriter, r *http.Request) {
	// 从密钥管理器获取 JWKS 数据喵
	jwksData, err := h.keyManager.GetJWKS()
	if err != nil {
		h.logf("ERROR", "获取 JWKS 失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
		return
	}

	// 记录访问日志喵
	h.logf("INFO", "JWKS 请求来自 %s", r.RemoteAddr)

	// 返回 JWKS 响应喵
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jwksData)
}
