// Package handler 提供 OIDC Token 端点处理喵
package handler

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"
)

// HandleToken 处理 POST /oidc/token 请求喵
func (h *Handler) HandleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	grantType := r.FormValue("grant_type")
	code := r.FormValue("code")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	if grantType != "authorization_code" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
		return
	}
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	// 客户端认证：支持 client_secret_post + client_secret_basic 喵
	client := h.store.FindClient(clientID)
	if client == nil {
		clientID, clientSecret, _ = parseBasicAuth(r)
		client = h.store.FindClient(clientID)
	}
	if client == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_client"})
		return
	}
	if clientSecret != "" && client.ClientSecret != clientSecret {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_client"})
		return
	}

	// 获取临时 code 喵
	tempData := h.store.GetTempCode(code)
	if tempData == nil {
		h.logf("WARN", "Token: 无效 code=%s", maskString(code))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	if tempData.ClientID != client.ClientID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	h.store.DeleteTempCode(code)

	// 生成 id_token 喵
	now := time.Now()
	claims := tempData.ToIDTokenClaims(h.store.GetSetting("issuer"), client.ClientID, now)
	token, err := h.keyManager.GenerateIDToken(claims)
	if err != nil {
		h.logf("ERROR", "生成 id_token 失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
		return
	}

	h.logf("INFO", "Token: 签发 id_token client=%s", client.ClientID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"id_token":     token,
	})
}

func parseBasicAuth(r *http.Request) (username, password string, ok bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Basic ") {
		return "", "", false
	}
	payload, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		return "", "", false
	}
	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		return "", "", false
	}
	return pair[0], pair[1], true
}
