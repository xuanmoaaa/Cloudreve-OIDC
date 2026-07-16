// Package handler 提供 OIDC 发现端点处理喵
package handler

import (
	"fmt"
	"net/http"
)

// HandleDiscovery 处理 GET /.well-known/openid-configuration 请求喵
func (h *Handler) HandleDiscovery(w http.ResponseWriter, r *http.Request) {
	issuer := h.store.GetSetting("issuer")

	cfg := map[string]interface{}{
		"issuer":                                issuer,
		"authorization_endpoint":                fmt.Sprintf("%s/oidc/authorize", issuer),
		"token_endpoint":                        fmt.Sprintf("%s/oidc/token", issuer),
		"userinfo_endpoint":                     fmt.Sprintf("%s/oidc/userinfo", issuer),
		"jwks_uri":                              fmt.Sprintf("%s/.well-known/jwks.json", issuer),
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "profile"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"claims_supported":                      []string{"iss", "sub", "aud", "exp", "iat", "nonce", "name", "nickname", "picture", "gender"},
	}

	h.logf("INFO", "OIDC Discovery 请求来自 %s", r.RemoteAddr)
	writeJSON(w, http.StatusOK, cfg)
}
