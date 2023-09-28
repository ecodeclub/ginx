package token

import (
	"github.com/golang-jwt/jwt/v5"
)

type Token[T jwt.Claims] interface {
	Generate(claims T) (string, error)
	Verify(token string) (T, error)
}
