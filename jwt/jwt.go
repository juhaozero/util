package jwt

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"golang.org/x/sync/singleflight"
)

var (
	Single = &singleflight.Group{}
)

type JWT struct {
	SigningKey  []byte
	BufferTime  int64
	ExpiresTime int64
	Issuer      string
}

var (
	TokenExpired     = errors.New("Token is expired")
	TokenNotValidYet = errors.New("Token not active yet")
	TokenMalformed   = errors.New("That's not even a token")
	TokenInvalid     = errors.New("Couldn't handle this token:")
)

// NewJWT 新建一个Jwt
func NewJWT(signingKey string, buffer, expires int64) *JWT {
	return &JWT{
		SigningKey:  []byte(signingKey),
		BufferTime:  buffer,
		ExpiresTime: expires,
	}

}

// CreateClaims 创建令牌
func (j *JWT) CreateClaims(baseClaims BaseClaims) CustomClaims {

	claims := CustomClaims{
		BaseClaims: baseClaims,
		// 缓冲时间内会获得新的token刷新令牌 此时一个用户会存在两个有效令牌
		BufferTime: int64(time.Duration(j.BufferTime) / time.Second),

		RegisteredClaims: jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(time.Now().Add(time.Duration(j.BufferTime))),  // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.ExpiresTime))), // 过期时间 7天  配置文件
			Issuer:    j.Issuer,                                                         // 签名的发行者
		},
	}
	return claims
}

// CreateToken 创建一个token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// CreateTokenByOldToken 旧token 换新token 使用归并回源避免并发问题
func (j *JWT) CreateTokenByOldToken(oldToken string, claims CustomClaims) (string, error) {
	v, err, _ := Single.Do("JWT:"+oldToken, func() (interface{}, error) {
		return j.CreateToken(claims)
	})
	return v.(string), err
}

// ParseToken 解析 token
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
		return nil, TokenInvalid

	} else {
		return nil, TokenInvalid
	}
}
