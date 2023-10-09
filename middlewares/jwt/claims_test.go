package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	tests := []struct {
		name          string
		expire        time.Duration
		encryptionKey string
		want          *Options
	}{
		{
			name:          "normal",
			expire:        10 * time.Minute,
			encryptionKey: "sign key",
			want: &Options{
				Expire:        10 * time.Minute,
				EncryptionKey: "sign key",
				DecryptKey:    "sign key",
				Method:        jwt.SigningMethodHS256,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOptions(tt.expire, tt.encryptionKey)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithDecryptKey(t *testing.T) {
	tests := []struct {
		name       string
		decryptKey string
		want       *Options
	}{
		{
			name:       "set_another_key",
			decryptKey: "other sign key",
			want: &Options{
				Expire:        defaultExpire,
				EncryptionKey: encryptionKey,
				DecryptKey:    "other sign key",
				Method:        jwt.SigningMethodHS256,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOptions(defaultExpire, encryptionKey,
				WithDecryptKey(tt.decryptKey))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithIssuer(t *testing.T) {
	tests := []struct {
		name   string
		issuer string
		want   *Options
	}{
		{
			name:   "set_issuer",
			issuer: "foo",
			want: &Options{
				Issuer:        "foo",
				Expire:        defaultExpire,
				EncryptionKey: encryptionKey,
				DecryptKey:    encryptionKey,
				Method:        jwt.SigningMethodHS256,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOptions(defaultExpire, encryptionKey,
				WithIssuer(tt.issuer))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithMethod(t *testing.T) {
	tests := []struct {
		name   string
		method jwt.SigningMethod
		want   *Options
	}{
		{
			name:   "set_another_jwt_signing_method",
			method: jwt.SigningMethodHS384,
			want: &Options{
				Expire:        defaultExpire,
				EncryptionKey: encryptionKey,
				DecryptKey:    encryptionKey,
				Method:        jwt.SigningMethodHS384,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOptions(defaultExpire, encryptionKey,
				WithMethod(tt.method))
			assert.Equal(t, tt.want, got)
		})
	}
}
