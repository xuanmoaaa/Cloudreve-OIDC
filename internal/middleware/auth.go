// Package middleware 提供管理员认证中间件喵
package middleware

import (
	"net/http"

	"oidc-proxy/internal/store"
)

// AdminAuth 返回一个检查管理员登录态的 HTTP 中间件喵
func AdminAuth(s *store.SQLiteStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("admin_session")
			if err != nil || cookie.Value == "" {
				http.Redirect(w, r, "/admin/login", http.StatusFound)
				return
			}
			if !s.IsSessionValid(cookie.Value) {
				http.Redirect(w, r, "/admin/login", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
