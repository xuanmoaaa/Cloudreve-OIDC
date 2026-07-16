// Package handler 提供彩虹回调端点处理喵
package handler

import (
	"net/http"
	"time"

	"oidc-proxy/internal/model"
)

// HandleCallback 处理 GET /oidc/callback（彩虹回调，带 ?type=qq&code=xxx）喵
func (h *Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	code := query.Get("code")
	loginType := query.Get("type") // qq / wx / alipay
	// 注意：彩虹回调不带 state，我们通过 state 参数在 authorize 阶段自己保存的喵
	// 但彩虹 API v2 的回调格式是 ?type=qq&code=xxx，没有 state。
	// 我们需要通过其他方式关联。这里用一个简化方案：直接用 state query 参数喵
	rainState := query.Get("state")

	h.logf("INFO", "收到彩虹回调: type=%s, code=%s, state=%s", loginType, maskString(code), rainState)

	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing code"})
		return
	}

	// 尝试匹配 state 喵
	authReq, ok := h.getState(rainState)
	if !ok {
		// 彩虹 v2 回调不带 state，遍历查找匹配的 type（简化处理）喵
		h.logf("WARN", "state=%s 未匹配，尝试根据 type=%s 查找", rainState, loginType)
		authReq, ok = h.findStateByType(loginType)
		if !ok {
			h.logf("ERROR", "无法匹配授权请求: type=%s", loginType)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid state"})
			return
		}
	}
	h.deleteState(authReq.RainState)

	// 用 code 直接换用户信息喵（彩虹 connect.php?act=callback）喵
	cbResp, err := h.rainClient.ExchangeCallback(loginType, code)
	if err != nil {
		h.logf("ERROR", "用 code 换用户信息失败: %v", err)
		ru := buildRedirectURI(authReq.RedirectURI, map[string]string{
			"error": "server_error", "error_description": "获取用户信息失败", "state": authReq.State,
		})
		http.Redirect(w, r, ru, http.StatusFound)
		return
	}

	// 统一用户信息喵
	userInfo := model.RainbowUserInfo{
		SocialUID: cbResp.SocialUID,
		Nickname:  cbResp.Nickname,
		Avatar:    cbResp.FaceImg,
		Gender:    cbResp.Gender,
		Platform:  loginType,
	}

	// 生成中转站临时 code（5分钟过期）喵
	tempCode := generateRandom(16)
	h.store.SetTempCode(tempCode, model.TempCodeData{
		UserInfo: userInfo,
		Nonce:    authReq.Nonce,
		ClientID: authReq.ClientID,
		Expires:  time.Now().Add(5 * time.Minute),
	})

	// 记录登录日志喵
	clientName := authReq.ClientID
	if c := h.store.FindClient(authReq.ClientID); c != nil {
		clientName = c.Name
	}
	h.store.AddLoginLog(model.LoginLog{
		Time:     time.Now().Format("2006-01-02 15:04:05"),
		Client:   clientName,
		Platform: loginType,
		Nickname: cbResp.Nickname,
	})

	// 302 回第三方应用喵
	ru := buildRedirectURI(authReq.RedirectURI, map[string]string{
		"code":  tempCode,
		"state": authReq.State,
	})

	h.logf("INFO", "回调完成: user=%s, platform=%s", cbResp.Nickname, loginType)
	http.Redirect(w, r, ru, http.StatusFound)
}

// findStateByType 根据 type 查找匹配的 authRequest（彩虹 v2 回调不带 state 时的备选方案）喵
func (h *Handler) findStateByType(channel string) (model.AuthRequest, bool) {
	h.stateMux.RLock()
	defer h.stateMux.RUnlock()
	for _, v := range h.authStates {
		if v.Channel == channel {
			return v, true
		}
	}
	return model.AuthRequest{}, false
}
