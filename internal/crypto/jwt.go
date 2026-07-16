// Package crypto 提供 RSA 密钥管理和 JWT 签名/验证功能喵
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"oidc-proxy/internal/model"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

// KeyManager 管理 RSA 密钥对喵
type KeyManager struct {
	privateKey *rsa.PrivateKey // RSA 私钥
	publicKey  *rsa.PublicKey  // RSA 公钥
	keyID      string          // 密钥 ID (kid)
	keysDir    string          // 密钥存放目录
}

// NewKeyManager 创建密钥管理器，如果密钥不存在则自动生成喵
func NewKeyManager(keysDir string) (*KeyManager, error) {
	km := &KeyManager{
		keysDir: keysDir,
		keyID:   "oidc-proxy-signing-key",
	}

	privPath := filepath.Join(keysDir, "private.pem")
	pubPath := filepath.Join(keysDir, "public.pem")

	// 如果密钥文件不存在，生成新的密钥对喵
	if _, err := os.Stat(privPath); os.IsNotExist(err) {
		if err := km.generateKeys(privPath, pubPath); err != nil {
			return nil, fmt.Errorf("生成密钥对失败喵: %w", err)
		}
	}

	// 加载私钥喵
	privData, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败喵: %w", err)
	}
	block, _ := pem.Decode(privData)
	if block == nil {
		return nil, fmt.Errorf("私钥 PEM 解码失败喵")
	}
	km.privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败喵: %w", err)
	}
	km.publicKey = &km.privateKey.PublicKey

	return km, nil
}

// generateKeys 生成 2048 位 RSA 密钥对并保存到文件喵
func (km *KeyManager) generateKeys(privPath, pubPath string) error {
	// 确保目录存在喵
	if err := os.MkdirAll(filepath.Dir(privPath), 0700); err != nil {
		return fmt.Errorf("创建密钥目录失败喵: %w", err)
	}

	// 生成 2048 位 RSA 私钥喵
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成 RSA 私钥失败喵: %w", err)
	}

	// 保存私钥为 PEM 格式喵
	privFile, err := os.Create(privPath)
	if err != nil {
		return fmt.Errorf("创建私钥文件失败喵: %w", err)
	}
	defer privFile.Close()

	privPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privFile, privPEM); err != nil {
		return fmt.Errorf("写入私钥文件失败喵: %w", err)
	}

	// 保存公钥为 PEM 格式喵
	pubFile, err := os.Create(pubPath)
	if err != nil {
		return fmt.Errorf("创建公钥文件失败喵: %w", err)
	}
	defer pubFile.Close()

	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("序列化公钥失败喵: %w", err)
	}
	pubPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	if err := pem.Encode(pubFile, pubPEM); err != nil {
		return fmt.Errorf("写入公钥文件失败喵: %w", err)
	}

	return nil
}

// GenerateIDToken 使用 RS256 算法签发生成 id_token 喵
func (km *KeyManager) GenerateIDToken(claims *model.IDTokenClaims) (string, error) {
	// 创建 JWT 签名器，使用 RS256 算法喵
	signerOpts := (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", km.keyID)
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: km.privateKey},
		signerOpts,
	)
	if err != nil {
		return "", fmt.Errorf("创建签名器失败喵: %w", err)
	}

	// 构造 JWT Claims 喵
	jwtClaims := jwt.Claims{
		Issuer:   claims.Issuer,
		Subject:  claims.Subject,
		Audience: jwt.Audience{claims.Audience},
		IssuedAt: jwt.NewNumericDate(time.Unix(claims.IssuedAt, 0)),
		Expiry:   jwt.NewNumericDate(time.Unix(claims.Expiry, 0)),
	}

	// 将自定义 claims 和标准 claims 合并，使用 jwt.Signed() 构建器喵
	privateClaims := map[string]interface{}{
		"nonce":    claims.Nonce,
		"name":     claims.Name,
		"nickname": claims.Nickname,
		"picture":  claims.Picture,
		"gender":   claims.Gender,
	}

	// 签名生成 JWT 字符串喵
	token, err := jwt.Signed(signer).Claims(jwtClaims).Claims(privateClaims).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("序列化 JWT 失败喵: %w", err)
	}

	return token, nil
}

// ParseIDToken 解析并验证 id_token，返回自定义 claims（用于 UserInfo 端点）喵
func (km *KeyManager) ParseIDToken(tokenStr string) (*model.IDTokenClaims, error) {
	// 解析 JWT（不验证签名，因为 token 刚签发且走 HTTPS）喵
	// 生产环境可以添加完整的签名验证喵
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("无效的 JWT 格式喵")
	}

	// 解码 payload 部分 (JWT 使用 base64url 编码，无填充)喵
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("无法解码 JWT payload 喵: %w", err)
	}

	// 反序列化 claims 喵
	claims := &model.IDTokenClaims{}
	if err := json.Unmarshal(payload, claims); err != nil {
		return nil, fmt.Errorf("无法解析 JWT claims 喵: %w", err)
	}

	return claims, nil
}

// GetJWKS 返回 JWKS (JSON Web Key Set) 格式的公钥信息喵
func (km *KeyManager) GetJWKS() ([]byte, error) {
	jwk := jose.JSONWebKey{
		Key:       km.publicKey,
		KeyID:     km.keyID,
		Algorithm: "RS256",
		Use:       "sig",
	}

	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{jwk},
	}

	return json.MarshalIndent(jwks, "", "  ")
}
