package jwt

import (
	"testing"
	"time"

	"github.com/ecodeclub/ekit/bean/option"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	var genIDFn func() string
	tests := []struct {
		name          string
		expire        time.Duration
		encryptionKey string
		want          Options
	}{
		{
			name:          "normal",
			expire:        10 * time.Minute,
			encryptionKey: "sign key",
			want: Options{
				Expire:        10 * time.Minute,
				EncryptionKey: "sign key",
				DecryptKey:    "sign key",
				Method:        jwt.SigningMethodHS256,
				genIDFn:       genIDFn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOptions(tt.expire, tt.encryptionKey)
			got.genIDFn = genIDFn
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithDecryptKey(t *testing.T) {
	tests := []struct {
		name string
		fn   func() option.Option[Options]
		want string
	}{
		{
			name: "normal",
			fn: func() option.Option[Options] {
				return nil
			},
			want: encryptionKey,
		},
		{
			name: "set_another_key",
			fn: func() option.Option[Options] {
				return WithDecryptKey("another sign key")
			},
			want: "another sign key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.fn() == nil {
				got = NewOptions(defaultExpire, encryptionKey).
					DecryptKey
			} else {
				got = NewOptions(defaultExpire, encryptionKey,
					tt.fn()).DecryptKey
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithMethod(t *testing.T) {
	tests := []struct {
		name string
		fn   func() option.Option[Options]
		want jwt.SigningMethod
	}{
		{
			name: "normal",
			fn: func() option.Option[Options] {
				return nil
			},
			want: jwt.SigningMethodHS256,
		},
		{
			name: "set_another_method",
			fn: func() option.Option[Options] {
				return WithMethod(jwt.SigningMethodHS384)
			},
			want: jwt.SigningMethodHS384,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got jwt.SigningMethod
			if tt.fn() == nil {
				got = NewOptions(defaultExpire, encryptionKey).
					Method
			} else {
				got = NewOptions(defaultExpire, encryptionKey,
					tt.fn()).Method
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithIssuer(t *testing.T) {
	tests := []struct {
		name string
		fn   func() option.Option[Options]
		want string
	}{
		{
			name: "normal",
			fn: func() option.Option[Options] {
				return nil
			},
		},
		{
			name: "set_another_issuer",
			fn: func() option.Option[Options] {
				return WithIssuer("foo")
			},
			want: "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.fn() == nil {
				got = NewOptions(defaultExpire, encryptionKey).
					Issuer
			} else {
				got = NewOptions(defaultExpire, encryptionKey,
					tt.fn()).Issuer
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithGenIDFunc(t *testing.T) {
	tests := []struct {
		name string
		fn   func() option.Option[Options]
		want string
	}{
		{
			name: "normal",
			fn: func() option.Option[Options] {
				return nil
			},
		},
		{
			name: "set_another_gen_id_func",
			fn: func() option.Option[Options] {
				return WithGenIDFunc(func() string {
					return "unique id"
				})
			},
			want: "unique id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.fn() == nil {
				got = NewOptions(defaultExpire, encryptionKey).
					genIDFn()
			} else {
				got = NewOptions(defaultExpire, encryptionKey,
					tt.fn()).genIDFn()
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
