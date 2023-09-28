package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler[T jwt.Claims] interface {
	ExtractTokenString(ctx *gin.Context) string
	VerifyToken(ctx *gin.Context, token string) error
	SetAccessToken(ctx *gin.Context, claims T) error
	SetRefreshToken(ctx *gin.Context, claims T) error
}
