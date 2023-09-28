package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTToken[T jwt.Claims] struct {
	encryptionKey string // 加密密钥
	decryptKey    string // 解密密钥
	nowFunc       func() time.Time
	method        jwt.SigningMethod
}

// NewJWTToken
// method: 默认签名加密方式使用 SH256
// decryptKey: 因默认使用对称加密所以与 encryptionKey 相同
func NewJWTToken[T jwt.Claims](encryptionKey string, opts ...Option[T]) *JWTToken[T] {
	dOpts := defaultOption[T]()
	dOpts.encryptionKey = encryptionKey
	dOpts.decryptKey = encryptionKey

	for _, opt := range opts {
		opt.apply(&dOpts)
	}

	return &dOpts
}

type Option[T jwt.Claims] interface {
	apply(*JWTToken[T])
}

type funcOption[T jwt.Claims] struct {
	f func(*JWTToken[T])
}

func (fdo *funcOption[T]) apply(do *JWTToken[T]) {
	fdo.f(do)
}

func newFuncOption[T jwt.Claims](f func(*JWTToken[T])) *funcOption[T] {
	return &funcOption[T]{
		f: f,
	}
}

func defaultOption[T jwt.Claims]() JWTToken[T] {
	return JWTToken[T]{
		nowFunc: time.Now,
		method:  jwt.SigningMethodHS256,
	}
}

func WithDecryptKey[T jwt.Claims](decryptKey string) Option[T] {
	return newFuncOption(func(j *JWTToken[T]) {
		j.decryptKey = decryptKey
	})
}

func WithNowFunc[T jwt.Claims](nowFunc func() time.Time) Option[T] {
	return newFuncOption(func(j *JWTToken[T]) {
		j.nowFunc = nowFunc
	})
}

func WithSigningMethod[T jwt.Claims](method jwt.SigningMethod) Option[T] {
	return newFuncOption(func(j *JWTToken[T]) {
		j.method = method
	})
}

// Generate 生成 jwt token.
func (j *JWTToken[T]) Generate(claims T) (string, error) {
	token := jwt.NewWithClaims(j.method, claims)
	return token.SignedString([]byte(j.encryptionKey))
}

// Verify 验证token.验证不通过则返回 error.
func (j *JWTToken[T]) Verify(token string) (T, error) {
	var claimsZero T
	claims := claimsZero
	var claimsPtr any = &claims
	t, err := jwt.ParseWithClaims(token, claimsPtr.(jwt.Claims),
		func(*jwt.Token) (interface{}, error) {
			return []byte(j.decryptKey), nil
		},
		jwt.WithTimeFunc(j.nowFunc),
	)
	if err != nil || !t.Valid {
		return claimsZero, fmt.Errorf("验证失败: %v", err)
	}

	return claims, nil
}
