package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/sync/singleflight"
)

var (
	ErrNotInRenewWindow   = errors.New("not in renew window")
	ErrEmptySigningKey    = errors.New("signing key is empty")
	ErrSigningKeyTooShort = errors.New("signing key is too short")
	ErrInvalidExpires     = errors.New("expires time must be positive")
	ErrInvalidToken       = errors.New("token is invalid")
	ErrInvalidAuthHeader  = errors.New("invalid authorization header")

	flight = &singleflight.Group{}
)

const minSigningKeyLen = 32

type CustomClaims struct {
	BaseClaims
	BufferTime int64 `json:"bufferTime"`
	jwt.RegisteredClaims
}

type BaseClaims struct {
	UserId       string         `json:"userId"`
	Username     string         `json:"username"`
	CustomParams map[string]any `json:"customParams,omitempty"`
}

type JWT struct {
	SigningKey  []byte
	BufferTime  time.Duration
	ExpiresTime time.Duration
	Issuer      string
}

// ValidateSigningKey 校验 HS256 签名密钥强度。
func ValidateSigningKey(key string) error {
	if key == "" {
		return ErrEmptySigningKey
	}
	if len(key) < minSigningKeyLen {
		return ErrSigningKeyTooShort
	}
	return nil
}

// NewJWT 新建 JWT 实例，buffer 和 expires 使用标准 time.Duration（如 24*time.Hour）。
func NewJWT(signingKey string, buffer, expires time.Duration, issuer string) (*JWT, error) {
	if err := ValidateSigningKey(signingKey); err != nil {
		return nil, err
	}
	if expires <= 0 {
		return nil, ErrInvalidExpires
	}
	return &JWT{
		SigningKey:  []byte(signingKey),
		BufferTime:  buffer,
		ExpiresTime: expires,
		Issuer:      issuer,
	}, nil
}

// MustNewJWT 新建 JWT 实例，配置非法时 panic。
func MustNewJWT(signingKey string, buffer, expires time.Duration, issuer string) *JWT {
	j, err := NewJWT(signingKey, buffer, expires, issuer)
	if err != nil {
		panic(err)
	}
	return j
}

// CreateClaims 根据业务 claims 构造 JWT claims。
func (j *JWT) CreateClaims(baseClaims BaseClaims) CustomClaims {
	now := time.Now()
	return CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: int64(j.BufferTime / time.Second),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Issuer,
			Subject:   baseClaims.UserId,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-time.Second)),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.ExpiresTime)),
		},
	}
}

// renewClaims 续期 claims
func (j *JWT) renewClaims(old CustomClaims) CustomClaims {
	now := time.Now()
	old.BufferTime = int64(j.BufferTime / time.Second)
	old.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    j.Issuer,
		Subject:   old.UserId,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(-time.Second)),
		ExpiresAt: jwt.NewNumericDate(now.Add(j.ExpiresTime)),
	}
	return old
}

// CreateToken 签发 token。
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// GenerateToken 创建 claims 并签发 token。
func (j *JWT) GenerateToken(baseClaims BaseClaims) (string, error) {
	return j.CreateToken(j.CreateClaims(baseClaims))
}

// CreateTokenByOldToken 旧 token 换新 token，使用 singleflight 避免并发重复刷新。
func (j *JWT) CreateTokenByOldToken(oldToken string, claims CustomClaims) (string, error) {
	v, err, _ := flight.Do(j.singleflightKey(oldToken), func() (any, error) {
		return j.CreateToken(claims)
	})
	if err != nil {
		return "", err
	}
	s, ok := v.(string)
	if !ok {
		return "", ErrInvalidToken
	}
	return s, nil
}

func (j *JWT) singleflightKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "JWT:" + hex.EncodeToString(sum[:])
}

func (j *JWT) parseOptions() []jwt.ParserOption {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	}
	if j.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(j.Issuer))
	}
	return opts
}

func (j *JWT) keyFunc(token *jwt.Token) (any, error) {
	if token.Method != jwt.SigningMethodHS256 {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
	}
	return j.SigningKey, nil
}

// ParseToken 解析并校验 token。
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, j.keyFunc, j.parseOptions()...)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ShouldRefresh 判断 token 是否进入缓冲续期窗口。
func (j *JWT) ShouldRefresh(claims *CustomClaims) bool {
	if claims == nil || claims.ExpiresAt == nil {
		return false
	}
	return time.Until(claims.ExpiresAt.Time) <= j.BufferTime
}

// IsExpired 判断 token 是否已过期。
func (j *JWT) IsExpired(claims *CustomClaims) bool {
	if claims == nil || claims.ExpiresAt == nil {
		return true
	}
	return time.Now().After(claims.ExpiresAt.Time)
}

// RemainingTime 返回 token 剩余有效时间。
func (j *JWT) RemainingTime(claims *CustomClaims) time.Duration {
	if claims == nil || claims.ExpiresAt == nil {
		return 0
	}
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RenewToken 在缓冲窗口内续期 token，返回新 token。
func (j *JWT) RenewToken(claims CustomClaims, token string) (string, error) {
	if !j.ShouldRefresh(&claims) {
		return "", ErrNotInRenewWindow
	}
	return j.CreateTokenByOldToken(token, j.renewClaims(claims))
}

// RefreshToken 解析旧 token 并在缓冲窗口内续期，返回新 token。
func (j *JWT) RefreshToken(oldToken string) (string, error) {
	claims, err := j.ParseToken(oldToken)
	if err != nil {
		return "", err
	}
	if !j.ShouldRefresh(claims) {
		return "", ErrNotInRenewWindow
	}
	return j.CreateTokenByOldToken(oldToken, j.renewClaims(*claims))
}

// ParseBearerToken 从 Authorization 头中提取 Bearer token。
func ParseBearerToken(authHeader string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", ErrInvalidAuthHeader
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	if token == "" {
		return "", ErrInvalidAuthHeader
	}
	return token, nil
}
