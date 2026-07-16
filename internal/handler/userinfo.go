// Package handler 提供 OIDC UserInfo 端点处理喵
package handler

import (
	"net/http"
	"strings"
)

// HandleUserInfo 处理 GET /oidc/userinfo 请求喵
func (h *Handler) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing_token"})
		return
	}
	token := authHeader[7:]

	claims, err := h.keyManager.ParseIDToken(token)
	if err != nil {
		h.logf("WARN", "UserInfo: 无效 token: %v", err)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
		return
	}

	h.logf("INFO", "UserInfo 请求: sub=%s", claims.Subject)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sub":      claims.Subject,
		"name":     claims.Name,
		"nickname": claims.Nickname,
		"picture":  claims.Picture,
		"gender":   claims.Gender,
	})
}
