package jwt

import (
	"github.com/gin-gonic/gin"
)

type LoginManager[T any] interface {
	Refresh(ctx *gin.Context)           // 刷新 token 的 gin.HandlerFunc
	MiddlewareBuilder() gin.HandlerFunc // 登录认证的中间件

	GenerateAccessToken(data T) (string, error)                  // 生成资源 token
	VerifyAccessToken(token string) (RegisteredClaims[T], error) // 校验资源 token
	// GenerateRefreshToken 生成刷新 token.
	// 需要设置 refreshJWTOptions 否则返回 ErrEmptyRefreshOpts 错误.
	GenerateRefreshToken(data T) (string, error)
	// VerifyRefreshToken 校验刷新 token.
	// 需要设置 refreshJWTOptions 否则返回 ErrEmptyRefreshOpts 错误.
	VerifyRefreshToken(token string) (RegisteredClaims[T], error) // 校验刷新 token
}
