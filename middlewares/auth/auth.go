package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/ecodeclub/ginx/middlewares/token"
)

type authHandler[T jwt.Claims] struct {
	allowTokenHeader    string
	bearerPrefix        string
	claimsCTXKey        string
	exposeAccessHeader  string
	exposeRefreshHeader string
	token               token.Token[T]
}

func NewAuthHandler[T jwt.Claims](token token.Token[T],
	opts ...authHdlOption[T]) Handler[T] {
	dOpts := defaultAuthHdlOption[T]()
	dOpts.token = token

	for _, opt := range opts {
		opt.apply(&dOpts)
	}

	return &dOpts
}

type authHdlOption[T jwt.Claims] interface {
	apply(*authHandler[T])
}

type funcAuthHdlOption[T jwt.Claims] struct {
	f func(handler *authHandler[T])
}

func (fdo *funcAuthHdlOption[T]) apply(do *authHandler[T]) {
	fdo.f(do)
}

func newFuncAuthHdlOption[T jwt.Claims](
	f func(handler *authHandler[T])) *funcAuthHdlOption[T] {
	return &funcAuthHdlOption[T]{
		f: f,
	}
}

func defaultAuthHdlOption[T jwt.Claims]() authHandler[T] {
	return authHandler[T]{
		allowTokenHeader:    "authorization",
		bearerPrefix:        "Bearer",
		claimsCTXKey:        "claims",
		exposeAccessHeader:  "x-access-token",
		exposeRefreshHeader: "x-refresh-token",
	}
}

func WithAllowTokenHeader[T jwt.Claims](header string) authHdlOption[T] {
	return newFuncAuthHdlOption[T](func(h *authHandler[T]) {
		h.allowTokenHeader = header
	})
}

func WithBearerPrefix[T jwt.Claims](prefix string) authHdlOption[T] {
	return newFuncAuthHdlOption[T](func(h *authHandler[T]) {
		h.bearerPrefix = prefix
	})
}

func WithClaimsCTXKey[T jwt.Claims](key string) authHdlOption[T] {
	return newFuncAuthHdlOption[T](func(h *authHandler[T]) {
		h.claimsCTXKey = key
	})
}

func WithExposeAccessHeader[T jwt.Claims](header string) authHdlOption[T] {
	return newFuncAuthHdlOption[T](func(h *authHandler[T]) {
		h.exposeAccessHeader = header
	})
}

func WithExposeRefreshHeader[T jwt.Claims](header string) authHdlOption[T] {
	return newFuncAuthHdlOption[T](func(h *authHandler[T]) {
		h.exposeRefreshHeader = header
	})
}

// ExtractTokenString 提取 token
func (a *authHandler[T]) ExtractTokenString(ctx *gin.Context) string {
	authCode := ctx.GetHeader(a.allowTokenHeader)
	if authCode == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString(a.bearerPrefix)
	b.WriteString(" ")
	prefix := b.String()
	if strings.HasPrefix(authCode, prefix) {
		return authCode[len(prefix):]
	}
	return ""
}

func (a *authHandler[T]) VerifyToken(ctx *gin.Context, token string) error {
	claims, err := a.token.Verify(token)
	if err != nil {
		return err
	}
	ctx.Set(a.claimsCTXKey, claims)
	return nil
}

func (a *authHandler[T]) SetAccessToken(ctx *gin.Context, claims T) error {
	tokenStr, err := a.token.Generate(claims)
	if err != nil {
		return err
	}
	ctx.Header(a.exposeAccessHeader, tokenStr)
	return nil
}

func (a *authHandler[T]) SetRefreshToken(ctx *gin.Context, claims T) error {
	tokenStr, err := a.token.Generate(claims)
	if err != nil {
		return err
	}
	ctx.Header(a.exposeRefreshHeader, tokenStr)
	return nil
}
