// Package rainbow 封装彩虹聚合登录 connect.php API 调用喵
package rainbow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"oidc-proxy/internal/model"
)

// Client 彩虹聚合登录 API 客户端喵（使用 connect.php 接口）
type Client struct {
	appID      string // 彩虹 APP ID
	appKey     string // 彩虹 APP KEY
	connectURL string // connect.php URL（如 https://login.mapay.cn/connect.php）
}

// NewClient 创建彩虹 API 客户端喵
func NewClient(appID, appKey, connectURL string) *Client {
	return &Client{
		appID:      appID,
		appKey:     appKey,
		connectURL: connectURL,
	}
}

// UpdateConfig 运行时更新彩虹配置喵
func (c *Client) UpdateConfig(cfg model.RainConfig) {
	c.appID = cfg.AppID
	c.appKey = cfg.AppKey
	c.connectURL = cfg.ConnectURL
}

// ────────── Step1: 获取登录跳转 URL ──────────

// GetLoginURL 调用 connect.php?act=login 获取第三方登录跳转地址喵
// channel: qq / wx / alipay
func (c *Client) GetLoginURL(channel, redirectURI string) (*model.RainLoginResp, error) {
	params := url.Values{}
	params.Set("act", "login")
	params.Set("appid", c.appID)
	params.Set("appkey", c.appKey)
	params.Set("type", channel)
	params.Set("redirect_uri", redirectURI)

	reqURL := fmt.Sprintf("%s?%s", c.connectURL, params.Encode())

	resp, err := http.DefaultClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("获取登录 URL 失败喵: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败喵: %w", err)
	}

	var result model.RainLoginResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败喵: %w, body=%s", err, string(body))
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("彩虹返回错误: code=%d, msg=%s", result.Code, result.Msg)
	}

	return &result, nil
}

// ────────── Step2: code 换用户信息 ──────────

// ExchangeCallback 调用 connect.php?act=callback 用 code 换用户信息喵
func (c *Client) ExchangeCallback(channel, code string) (*model.RainCallbackResp, error) {
	params := url.Values{}
	params.Set("act", "callback")
	params.Set("appid", c.appID)
	params.Set("appkey", c.appKey)
	params.Set("type", channel)
	params.Set("code", code)

	reqURL := fmt.Sprintf("%s?%s", c.connectURL, params.Encode())

	resp, err := http.DefaultClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败喵: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败喵: %w", err)
	}

	var result model.RainCallbackResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败喵: %w, body=%s", err, string(body))
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("彩虹返回错误: code=%d, msg=%s", result.Code, result.Msg)
	}

	return &result, nil
}

// ────────── 查询用户信息 ──────────

// QueryUser 调用 connect.php?act=query 查询用户信息喵
func (c *Client) QueryUser(socialUID string) (*model.RainQueryResp, error) {
	params := url.Values{}
	params.Set("act", "query")
	params.Set("appid", c.appID)
	params.Set("appkey", c.appKey)
	params.Set("social_uid", socialUID)

	reqURL := fmt.Sprintf("%s?%s", c.connectURL, params.Encode())

	resp, err := http.DefaultClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("查询用户信息失败喵: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败喵: %w", err)
	}

	var result model.RainQueryResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败喵: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("彩虹返回错误: code=%d, msg=%s", result.Code, result.Msg)
	}

	return &result, nil
}

// CheckConnectivity 检查连通性喵
func (c *Client) CheckConnectivity() bool {
	resp, err := http.DefaultClient.Get(c.connectURL)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode < 500
}
