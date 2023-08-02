package jwt

import (
	jwt "github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	BaseClaims
	BufferTime int64
	jwt.RegisteredClaims
}
type BaseClaims struct {
	UUID     string
	Username string
	Name     string
}
