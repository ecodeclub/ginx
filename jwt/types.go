package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Manager jwt 管理器.
type Manager[T any] interface {
	// MiddlewareBuilder 创建登录认证的中间件.
	MiddlewareBuilder() *MiddlewareBuilder[T]

	// Refresh 刷新 token 的 gin.HandlerFunc.
	// 需要设置 refreshJWTOptions 否则会出现 500 的 http 状态码.
	Refresh(ctx *gin.Context)

	// GenerateAccessToken 生成资源 token.
	GenerateAccessToken(data T) (string, error)

	// VerifyAccessToken 校验资源 token.
	VerifyAccessToken(token string, opts ...jwt.ParserOption) (RegisteredClaims[T], error)

	// GenerateRefreshToken 生成刷新 token.
	// 需要设置 refreshJWTOptions 否则返回 errEmptyRefreshOpts 错误.
	GenerateRefreshToken(data T) (string, error)

	// VerifyRefreshToken 校验刷新 token.
	// 需要设置 refreshJWTOptions 否则返回 errEmptyRefreshOpts 错误.
	VerifyRefreshToken(token string, opts ...jwt.ParserOption) (RegisteredClaims[T], error)

	// SetClaims 设置 claims 到 key=`claims` 的 gin.Context 中.
	SetClaims(ctx *gin.Context, claims RegisteredClaims[T])
}

type RegisteredClaims[T any] struct {
	Data T `json:"data"`
	jwt.RegisteredClaims
}
