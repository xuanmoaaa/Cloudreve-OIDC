// Package model 定义项目中使用的所有数据结构喵
package model

import "time"

// ────────── 彩虹聚合登录 API 数据结构（基于 connect.php） ──────────

// RainLoginResp connect.php?act=login 返回结构喵
type RainLoginResp struct {
	Code   int    `json:"code"`   // 0=成功
	Msg    string `json:"msg"`    // 错误信息
	Type   string `json:"type"`   // qq/wx/alipay
	URL    string `json:"url"`    // 跳转登录地址
	QRCode string `json:"qrcode"` // 扫码地址（仅微信/支付宝返回）
}

// RainCallbackResp connect.php?act=callback 返回结构（直接用 code 换用户信息）喵
type RainCallbackResp struct {
	Code        int    `json:"code"`         // 0=成功, 2=未完成
	Msg         string `json:"msg"`          // 错误信息
	Type        string `json:"type"`         // qq/wx/alipay
	SocialUID   string `json:"social_uid"`   // 第三方用户 UID
	AccessToken string `json:"access_token"` // 用户 token
	Nickname    string `json:"nickname"`     // 昵称
	FaceImg     string `json:"faceimg"`      // 头像
	Gender      string `json:"gender"`       // 性别
	Location    string `json:"location"`     // 所在地
	IP          string `json:"ip"`           // 登录 IP
}

// RainQueryResp connect.php?act=query 返回结构喵
type RainQueryResp struct {
	Code        int    `json:"code"`
	Msg         string `json:"msg"`
	Type        string `json:"type"`
	SocialUID   string `json:"social_uid"`
	AccessToken string `json:"access_token"`
	Nickname    string `json:"nickname"`
	FaceImg     string `json:"faceimg"`
	Gender      string `json:"gender"`
	Location    string `json:"location"`
}

// RainbowUserInfo 统一用户信息（从彩虹 API 映射到 OIDC claims）喵
type RainbowUserInfo struct {
	SocialUID string `json:"social_uid"` // 第三方 UID
	Nickname  string `json:"nickname"`   // 昵称
	Avatar    string `json:"avatar"`     // 头像
	Gender    string `json:"gender"`     // 性别
	Platform  string `json:"platform"`   // qq/wx/alipay
}

// ────────── OIDC 相关数据结构 ──────────

// AuthRequest 保存 OIDC 授权请求的临时数据喵
type AuthRequest struct {
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`
	State       string `json:"state"`
	Nonce       string `json:"nonce"`
	RainState   string `json:"rain_state"` // 彩虹 state
	Channel     string `json:"channel"`    // 登录通道 qq/wx/alipay
}

// TempCodeData 临时 code 对应的数据喵
type TempCodeData struct {
	UserInfo RainbowUserInfo `json:"user_info"`
	Nonce    string          `json:"nonce"`
	ClientID string          `json:"client_id"`
	Expires  time.Time       `json:"expires"`
}

// OIDCClient OIDC 客户端喵
type OIDCClient struct {
	Name         string `json:"name"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	CreatedAt    string `json:"created_at"`
}

// LoginLog 登录日志喵
type LoginLog struct {
	Time     string `json:"time"`
	Client   string `json:"client"`
	Platform string `json:"platform"`
	Nickname string `json:"nickname"`
}

// IDTokenClaims id_token 的 JWT Claims 喵
type IDTokenClaims struct {
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	Audience string `json:"aud"`
	IssuedAt int64  `json:"iat"`
	Expiry   int64  `json:"exp"`
	Nonce    string `json:"nonce"`
	Name     string `json:"name"`
	Nickname string `json:"nickname,omitempty"`
	Picture  string `json:"picture,omitempty"`
	Gender   string `json:"gender,omitempty"`
}

// ────────── 彩虹客户端配置（仅用于 rainbow 包） ──────────

type RainConfig struct {
	AppID      string
	AppKey     string
	ConnectURL string // connect.php 基础 URL
}

// ────────── 辅助方法 ──────────

func (t *TempCodeData) ToIDTokenClaims(issuer, audience string, now time.Time) *IDTokenClaims {
	return &IDTokenClaims{
		Issuer:   issuer,
		Subject:  t.UserInfo.SocialUID,
		Audience: audience,
		IssuedAt: now.Unix(),
		Expiry:   now.Add(1 * time.Hour).Unix(),
		Nonce:    t.Nonce,
		Name:     t.UserInfo.Nickname,
		Nickname: t.UserInfo.Nickname,
		Picture:  t.UserInfo.Avatar,
		Gender:   t.UserInfo.Gender,
	}
}
