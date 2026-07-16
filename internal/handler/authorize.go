// Package handler 提供 OIDC 授权端点处理喵
// 先展示 QQ/微信/支付宝通道选择页面，用户选择后跳转彩虹喵
package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"oidc-proxy/internal/model"
)

// HandleAuthorize 处理 GET /oidc/authorize 喵
// 如果带 channel 参数 → 调用彩虹 API 获取跳转 URL → 302 跳转喵
// 如果不带 channel → 展示通道选择页面喵
func (h *Handler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	responseType := query.Get("response_type")
	state := query.Get("state")
	nonce := query.Get("nonce")
	channel := query.Get("channel") // qq / wx / alipay

	// 校验必要参数喵
	if clientID == "" || redirectURI == "" || responseType != "code" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "缺少必要参数喵",
		})
		return
	}

	// 校验 client_id 喵
	client := h.store.FindClient(clientID)
	if client == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":             "unauthorized_client",
			"error_description": "未知的 client_id 喵",
		})
		return
	}

	// 如果用户选好了通道，执行跳转喵
	if channel != "" {
		h.handleChannelRedirect(w, r, clientID, redirectURI, state, nonce, channel)
		return
	}

	// 否则展示通道选择页面喵
	h.renderChannelPage(w, r, clientID, redirectURI, state, nonce)
}

// handleChannelRedirect 用户选好通道后，调用彩虹 API 获取跳转 URL，302 跳转喵
func (h *Handler) handleChannelRedirect(w http.ResponseWriter, r *http.Request,
	clientID, redirectURI, state, nonce, channel string) {

	// 保存授权请求信息喵
	rainState := generateRandom(16)
	authReq := model.AuthRequest{
		ClientID:    clientID,
		RedirectURI: redirectURI,
		State:       state,
		Nonce:       nonce,
		RainState:   rainState,
		Channel:     channel,
	}
	h.storeState(rainState, authReq)

	// 构造回调地址喵
	issuer := h.store.GetSetting("issuer")
	callbackURL := fmt.Sprintf("%s/oidc/callback", issuer)

	// 调用彩虹 API 获取真实的登录跳转 URL 喵
	loginResp, err := h.rainClient.GetLoginURL(channel, callbackURL)
	if err != nil {
		h.logf("ERROR", "获取彩虹登录 URL 失败: channel=%s, err=%v", channel, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "获取登录地址失败喵: " + err.Error(),
		})
		return
	}

	h.logf("INFO", "用户选择通道 %s，跳转彩虹 -> %s", channel, maskString(loginResp.URL))

	// 302 跳转到真实的 QQ/微信/支付宝登录页喵
	http.Redirect(w, r, loginResp.URL, http.StatusFound)
}

// ────────── 通道选择页面 ──────────

func (h *Handler) renderChannelPage(w http.ResponseWriter, r *http.Request,
	clientID, redirectURI, state, nonce string) {

	baseParams := fmt.Sprintf("client_id=%s&redirect_uri=%s&response_type=code&state=%s&nonce=%s",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
		url.QueryEscape(nonce),
	)

	// 纯 HTML 通道选择页面喵
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>选择登录方式</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","Microsoft YaHei","PingFang SC",sans-serif;background:#f5f6fa;display:flex;justify-content:center;align-items:center;min-height:100vh}
.card{background:#fff;border-radius:12px;padding:44px 40px;box-shadow:0 1px 3px rgba(0,0,0,.06),0 4px 12px rgba(0,0,0,.04);width:100%;max-width:400px}
.card h2{font-size:20px;color:#1d1d1f;margin-bottom:4px;font-weight:600;letter-spacing:-.3px}
.card .sub{color:#86868b;font-size:13px;margin-bottom:32px}
.channels{display:flex;flex-direction:column;gap:12px}
.chn{display:flex;align-items:center;padding:14px 18px;border-radius:8px;text-decoration:none;font-size:15px;color:#1d1d1f;border:1px solid #e5e5ea;transition:all .2s;background:#fafafa}
.chn:hover{border-color:#bbb;background:#f5f5f7}
.chn .icon{width:32px;height:32px;margin-right:14px;border-radius:6px;flex-shrink:0;object-fit:cover}
.chn .arrow{margin-left:auto;color:#c7c7cc;font-size:14px}
.divider{display:flex;align-items:center;margin:28px 0;color:#c7c7cc;font-size:12px}
.divider::before,.divider::after{content:"";flex:1;height:1px;background:#e5e5ea}
.divider span{padding:0 12px}
.footer{text-align:center;margin-top:4px;color:#c7c7cc;font-size:11px}
</style>
</head>
<body>
<div class="card">
<h2>选择登录方式</h2>
<p class="sub">请选择一种方式继续</p>
<div class="channels">
<a class="chn chn-qq" href="?` + baseParams + `&channel=qq">
<img src="/static/picture/qq.png" class="icon" alt="QQ"><span>QQ 登录</span><span class="arrow">&rsaquo;</span>
</a>
<a class="chn chn-wechat" href="?` + baseParams + `&channel=wx">
<img src="/static/picture/wx.png" class="icon" alt="微信"><span>微信登录</span><span class="arrow">&rsaquo;</span>
</a>
<a class="chn chn-alipay" href="?` + baseParams + `&channel=alipay">
<img src="/static/picture/alipay.png" class="icon" alt="支付宝"><span>支付宝登录</span><span class="arrow">&rsaquo;</span>
</a>
</div>
<div class="divider"><span>其他方式</span></div>
<p class="footer">BY OIDC</p>
</div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// ────────── 工具函数 ──────────

func (h *Handler) storeState(state string, req model.AuthRequest) {
	h.stateMux.Lock()
	defer h.stateMux.Unlock()
	h.authStates[state] = req
}

func (h *Handler) getState(state string) (model.AuthRequest, bool) {
	h.stateMux.RLock()
	defer h.stateMux.RUnlock()
	req, ok := h.authStates[state]
	return req, ok
}

func (h *Handler) deleteState(state string) {
	h.stateMux.Lock()
	defer h.stateMux.Unlock()
	delete(h.authStates, state)
}

func generateRandom(bytes int) string {
	b := make([]byte, bytes)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func buildRedirectURI(base string, params map[string]string) string {
	u, _ := url.Parse(base)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func maskString(s string) string {
	if len(s) > 12 {
		return s[:12] + "..."
	}
	return s
}

// EscapeHTML 简单 HTML 转义喵
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&#34;")
	return s
}
