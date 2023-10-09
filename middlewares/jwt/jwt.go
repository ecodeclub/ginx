package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrEmptyRefreshOpts = errors.New("refreshJWTOptions are nil")
)

type Manager[T any] struct {
	publicPaths         set.Set[string] // 存放不需要认证的 path
	allowTokenHeader    string          // 认证的请求头(存放 token 的请求头 key)
	bearerPrefix        string          // 拼接 token 的前缀
	claimsCTXKey        string          // 存放到 gin.Context 的 key
	exposeAccessHeader  string          // 暴露到外部的资源请求头
	exposeRefreshHeader string          // 暴露到外部的刷新请求头

	accessJWTOptions   *Options // 资源 token 选项
	refreshJWTOptions  *Options // 刷新 token 选项
	rotateRefreshToken bool     // 轮换刷新令牌

	nowFunc func() time.Time // 控制 jwt 的时间
}

// NewManager 定义一个 JWTLoginManger
// allowTokenHeader: 默认使用 authorization 为认证请求头.
// bearerPrefix: 默认使用 Bearer 拼接 token.
// claimsCTXKey: 默认使用 claims 为设置到 gin.Context 的key
// exposeAccessHeader: 默认使用 x-access-token 为暴露外部的资源请求头.
// exposeRefreshHeader: 默认使用 x-refresh-token 为暴露外部的刷新请求头.
// refreshJWTOptions: 默认使用 nil 为刷新 token 的配置,
// 如要使用 refresh 功能则需要使用 WithRefreshJWTOptions 添加相关配置.
// rotateRefreshToken: 默认不轮换刷新令牌.
// 该配置需要设置 refreshJWTOptions 才有效.
func NewManager[T any](accessJWTOptions *Options,
	opts ...ManagerOption[T]) *Manager[T] {
	dOpts := defaultManagerOption[T]()
	dOpts.accessJWTOptions = accessJWTOptions

	for _, opt := range opts {
		opt.apply(&dOpts)
	}

	return &dOpts
}

type ManagerOption[T any] interface {
	apply(*Manager[T])
}

type funcManagerOption[T any] struct {
	f func(handler *Manager[T])
}

func (fdo *funcManagerOption[T]) apply(do *Manager[T]) {
	fdo.f(do)
}

func newFuncManagerOption[T any](
	f func(handler *Manager[T])) *funcManagerOption[T] {
	return &funcManagerOption[T]{
		f: f,
	}
}

func defaultManagerOption[T any]() Manager[T] {
	return Manager[T]{
		publicPaths:         set.NewMapSet[string](0),
		allowTokenHeader:    "authorization",
		bearerPrefix:        "Bearer",
		claimsCTXKey:        "claims",
		exposeAccessHeader:  "x-access-token",
		exposeRefreshHeader: "x-refresh-token",
		rotateRefreshToken:  false,
		nowFunc:             time.Now,
	}
}

// WithIgnorePaths 设置忽略资源令牌认证的路径.
// 例如: '/login', '/api/v1/signup'.
func WithIgnorePaths[T any](paths ...string) ManagerOption[T] {
	s := set.NewMapSet[string](len(paths))
	for _, path := range paths {
		s.Add(path)
	}
	return newFuncManagerOption[T](func(l *Manager[T]) {
		l.publicPaths = s
	})
}

// WithAllowTokenHeader 设置允许 token 的请求头.
func WithAllowTokenHeader[T any](header string) ManagerOption[T] {
	return newFuncManagerOption[T](func(m *Manager[T]) {
		m.allowTokenHeader = header
	})
}

// WithBearerPrefix 设置与 token 拼接的前缀.
// 例如: 'Bearer eyx.eyx.x'中的 'Bearer'.
func WithBearerPrefix[T any](prefix string) ManagerOption[T] {
	return newFuncManagerOption[T](func(m *Manager[T]) {
		m.bearerPrefix = prefix
	})
}

// WithClaimsCTXKey 设置放到 gin.Context 中的 key.
func WithClaimsCTXKey[T any](key string) ManagerOption[T] {
	return newFuncManagerOption[T](func(m *Manager[T]) {
		m.claimsCTXKey = key
	})
}

// WithExposeAccessHeader 设置公开资源令牌的请求头.
func WithExposeAccessHeader[T any](header string) ManagerOption[T] {
	return newFuncManagerOption[T](func(m *Manager[T]) {
		m.exposeAccessHeader = header
	})
}

// WithExposeRefreshHeader 设置公开刷新令牌的请求头.
func WithExposeRefreshHeader[T any](header string) ManagerOption[T] {
	return newFuncManagerOption[T](func(m *Manager[T]) {
		m.exposeRefreshHeader = header
	})
}

// WithRefreshJWTOptions 设置刷新令牌相关的配置.
func WithRefreshJWTOptions[T any](refreshOpts *Options) ManagerOption[T] {
	return newFuncManagerOption(func(m *Manager[T]) {
		m.refreshJWTOptions = refreshOpts
	})
}

// WithRotateRefreshToken 设置轮换刷新令牌.
func WithRotateRefreshToken[T any](isRotate bool) ManagerOption[T] {
	return newFuncManagerOption(func(m *Manager[T]) {
		m.rotateRefreshToken = isRotate
	})
}

// WithNowFunc 设置当前时间.
// 一般用于测试.
func WithNowFunc[T any](nowFunc func() time.Time) ManagerOption[T] {
	return newFuncManagerOption(func(m *Manager[T]) {
		m.nowFunc = nowFunc
	})
}

// Refresh 刷新 token 的 gin.HandlerFunc
func (m *Manager[T]) Refresh(ctx *gin.Context) {
	if m.refreshJWTOptions == nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	tokenStr := m.extractTokenString(ctx)
	clm, err := m.VerifyRefreshToken(tokenStr)
	if err != nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	accessToken, err := m.GenerateAccessToken(clm.Data)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Header(m.exposeAccessHeader, accessToken)

	// 轮换刷新令牌
	if m.rotateRefreshToken {
		refreshToken, err := m.GenerateRefreshToken(clm.Data)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		ctx.Header(m.exposeRefreshHeader, refreshToken)
	}
	ctx.Status(http.StatusOK)
}

// MiddlewareBuilder 登录认证的中间件
func (m *Manager[T]) MiddlewareBuilder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要校验
		if m.publicPaths.Exist(ctx.Request.URL.Path) {
			return
		}

		tokenStr := m.extractTokenString(ctx)
		if tokenStr == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err := m.verifyTokenAndSetClm(ctx, tokenStr)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

// extractTokenString 提取 token 字符串.
func (m *Manager[T]) extractTokenString(ctx *gin.Context) string {
	authCode := ctx.GetHeader(m.allowTokenHeader)
	if authCode == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString(m.bearerPrefix)
	b.WriteString(" ")
	prefix := b.String()
	if strings.HasPrefix(authCode, prefix) {
		return authCode[len(prefix):]
	}
	return ""
}

// verifyTokenAndSetClm 校验 access token 并把 claims 设置到 gin.Context 中.
func (m *Manager[T]) verifyTokenAndSetClm(ctx *gin.Context, token string) error {
	claims, err := m.VerifyAccessToken(token)
	if err != nil {
		return err
	}
	ctx.Set(m.claimsCTXKey, claims)
	return nil
}

// GenerateAccessToken 生成资源 token.
func (m *Manager[T]) GenerateAccessToken(data T) (string, error) {
	nowTime := m.nowFunc()
	claims := RegisteredClaims[T]{
		Data: data,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.accessJWTOptions.Issuer,
			ExpiresAt: jwt.NewNumericDate(nowTime.Add(m.accessJWTOptions.Expire)),
			NotBefore: jwt.NewNumericDate(nowTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
		},
	}

	token := jwt.NewWithClaims(m.accessJWTOptions.Method, claims)
	return token.SignedString([]byte(m.accessJWTOptions.EncryptionKey))
}

// VerifyAccessToken 校验资源 token.
func (m *Manager[T]) VerifyAccessToken(token string) (RegisteredClaims[T], error) {
	t, err := jwt.ParseWithClaims(token, &RegisteredClaims[T]{},
		func(*jwt.Token) (interface{}, error) {
			return []byte(m.accessJWTOptions.DecryptKey), nil
		},
		jwt.WithTimeFunc(m.nowFunc),
	)
	if err != nil || !t.Valid {
		return RegisteredClaims[T]{}, fmt.Errorf("验证失败: %v", err)
	}
	clm, _ := t.Claims.(*RegisteredClaims[T])
	return *clm, nil
}

// GenerateRefreshToken 生成刷新 token.
// 需要设置 refreshJWTOptions 否则返回 ErrEmptyRefreshOpts 错误.
func (m *Manager[T]) GenerateRefreshToken(data T) (string, error) {
	if m.refreshJWTOptions == nil {
		return "", ErrEmptyRefreshOpts
	}

	nowTime := m.nowFunc()
	claims := RegisteredClaims[T]{
		Data: data,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.refreshJWTOptions.Issuer,
			ExpiresAt: jwt.NewNumericDate(nowTime.Add(m.refreshJWTOptions.Expire)),
			NotBefore: jwt.NewNumericDate(nowTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
		},
	}

	token := jwt.NewWithClaims(m.refreshJWTOptions.Method, claims)
	return token.SignedString([]byte(m.refreshJWTOptions.EncryptionKey))
}

// VerifyRefreshToken 校验刷新 token.
// 需要设置 refreshJWTOptions 否则返回 ErrEmptyRefreshOpts 错误.
func (m *Manager[T]) VerifyRefreshToken(token string) (RegisteredClaims[T], error) {
	if m.refreshJWTOptions == nil {
		return RegisteredClaims[T]{}, ErrEmptyRefreshOpts
	}
	t, err := jwt.ParseWithClaims(token, &RegisteredClaims[T]{},
		func(*jwt.Token) (interface{}, error) {
			return []byte(m.refreshJWTOptions.DecryptKey), nil
		},
		jwt.WithTimeFunc(m.nowFunc),
	)
	if err != nil || !t.Valid {
		return RegisteredClaims[T]{}, fmt.Errorf("验证失败: %v", err)
	}
	clm, _ := t.Claims.(*RegisteredClaims[T])
	return *clm, nil
}
