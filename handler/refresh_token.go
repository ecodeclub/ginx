package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/ecodeclub/ginx/middlewares/auth"
)

type TokenHandler[T jwt.Claims] interface {
	Refresh(ctx *gin.Context)
}

type tokenHandler[T jwt.Claims] struct {
	accessClaims T
	auth.Handler[T]
}

func NewTokenHandler[T jwt.Claims](
	accessClaims T, handler auth.Handler[T]) TokenHandler[T] {
	return &tokenHandler[T]{
		accessClaims: accessClaims,
		Handler:      handler,
	}
}

func (t *tokenHandler[T]) Refresh(ctx *gin.Context) {
	tokenStr := t.ExtractTokenString(ctx)
	err := t.VerifyToken(ctx, tokenStr)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = t.SetAccessToken(ctx, t.accessClaims)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}
