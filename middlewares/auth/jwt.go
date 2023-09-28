package auth

import (
	"net/http"

	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTBuilder[T jwt.Claims] struct {
	publicPaths set.Set[string]
	Handler[T]
}

func NewJWTBuilder[T jwt.Claims](handler Handler[T], opts ...BuilderOption[T]) *JWTBuilder[T] {
	dOpts := JWTBuilder[T]{
		publicPaths: set.NewMapSet[string](0),
		Handler:     handler,
	}

	for _, opt := range opts {
		opt.apply(&dOpts)
	}

	return &dOpts
}

type BuilderOption[T jwt.Claims] interface {
	apply(*JWTBuilder[T])
}

type funcBuilderOption[T jwt.Claims] struct {
	f func(*JWTBuilder[T])
}

func (fdo *funcBuilderOption[T]) apply(do *JWTBuilder[T]) {
	fdo.f(do)
}

func newFuncBuilderOption[T jwt.Claims](f func(*JWTBuilder[T])) *funcBuilderOption[T] {
	return &funcBuilderOption[T]{
		f: f,
	}
}

func WithIgnorePaths[T jwt.Claims](paths ...string) BuilderOption[T] {
	s := set.NewMapSet[string](len(paths))
	for _, path := range paths {
		s.Add(path)
	}
	return newFuncBuilderOption[T](func(b *JWTBuilder[T]) {
		b.publicPaths = s
	})
}

func (b *JWTBuilder[T]) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要校验
		if b.publicPaths.Exist(ctx.Request.URL.Path) {
			return
		}

		tokenStr := b.ExtractTokenString(ctx)
		if tokenStr == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err := b.VerifyToken(ctx, tokenStr)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
