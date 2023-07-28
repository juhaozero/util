package jwt

import (
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type CustomClaims struct {
	BaseClaims
	BufferTime int64
	jwt.RegisteredClaims
}
type BaseClaims struct {
	UUID     uuid.UUID
	ID       uint
	Username string
	Name     string
}
