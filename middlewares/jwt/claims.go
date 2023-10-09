package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RegisteredClaims[T any] struct {
	Data T `json:"data"`
	jwt.RegisteredClaims
}

type Options struct {
	Issuer        string            // 签发人
	Expire        time.Duration     // 有效期
	EncryptionKey string            // 加密密钥
	DecryptKey    string            // 解密密钥
	Method        jwt.SigningMethod // 签名方式
}

// NewOptions 定义一个JWT Claims配置
// Issuer: 默认使用 "".
// DecryptKey: 默认与 EncryptionKey 相同.
// Method: 默认使用 jwt.SigningMethodHS256 签名方式.
func NewOptions(expire time.Duration, encryptionKey string, opts ...Option) *Options {
	dOpts := Options{
		Expire:        expire,
		EncryptionKey: encryptionKey,
		DecryptKey:    encryptionKey,
		Method:        jwt.SigningMethodHS256,
	}

	for _, opt := range opts {
		opt.apply(&dOpts)
	}

	return &dOpts
}

type Option interface {
	apply(*Options)
}

type funcOption struct {
	f func(handler *Options)
}

func (fdo *funcOption) apply(do *Options) {
	fdo.f(do)
}

func newFuncOption(
	f func(*Options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// WithIssuer 设置签发人.
func WithIssuer(issuer string) Option {
	return newFuncOption(func(o *Options) {
		o.Issuer = issuer
	})
}

// WithDecryptKey 设置解密密钥.
func WithDecryptKey(decryptKey string) Option {
	return newFuncOption(func(o *Options) {
		o.DecryptKey = decryptKey
	})
}

// WithMethod 设置 JWT 的签名方法.
func WithMethod(method jwt.SigningMethod) Option {
	return newFuncOption(func(o *Options) {
		o.Method = method
	})
}
