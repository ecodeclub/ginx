package token

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTToken_Generate(t *testing.T) {
	j := NewJWTToken[jwt.RegisteredClaims]("foo")
	nowTime := time.UnixMilli(1695571200000)
	type testCase[T jwt.Claims] struct {
		name    string
		claims  jwt.RegisteredClaims
		want    string
		wantErr error
	}
	tests := []testCase[jwt.RegisteredClaims]{
		{
			name: "生成token",
			claims: jwt.RegisteredClaims{
				Issuer:    "bar",
				Subject:   "1",
				IssuedAt:  jwt.NewNumericDate(nowTime),
				ExpiresAt: jwt.NewNumericDate(nowTime.Add(10 * time.Minute)),
			},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.a1q3jHKedQGbA-Zrn6S21QUpI2ZNYNHoeG5LkxAXRJQ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := j.Generate(tt.claims)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJWTToken_Verify(t *testing.T) {
	j := NewJWTToken[jwt.RegisteredClaims]("foo")
	type testCase[T jwt.Claims] struct {
		name    string
		nowFunc func() time.Time
		token   string
		want    jwt.RegisteredClaims
		wantErr error
	}
	tests := []testCase[jwt.RegisteredClaims]{
		{
			name: "验证通过",
			nowFunc: func() time.Time {
				return time.UnixMilli(1695571500000)
			},
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.a1q3jHKedQGbA-Zrn6S21QUpI2ZNYNHoeG5LkxAXRJQ",
			want: jwt.RegisteredClaims{
				Issuer:    "bar",
				Subject:   "1",
				IssuedAt:  jwt.NewNumericDate(time.UnixMilli(1695571200000)),
				ExpiresAt: jwt.NewNumericDate(time.UnixMilli(1695571200000).Add(10 * time.Minute)),
			},
		},
		{
			name: "token过期了",
			nowFunc: func() time.Time {
				return time.UnixMilli(1695572500000)
			},
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.a1q3jHKedQGbA-Zrn6S21QUpI2ZNYNHoeG5LkxAXRJQ",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired)),
		},
		{
			name: "token签名错误",
			nowFunc: func() time.Time {
				return time.UnixMilli(1695571500000)
			},
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.5OeEzR5tNTGmXwvloac2wYdZvlt8U5UmFdsnpBJ_zb4",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenSignatureInvalid, jwt.ErrSignatureInvalid)),
		},
		{
			name: "错误的token",
			nowFunc: func() time.Time {
				return time.UnixMilli(1695571500000)
			},
			token: "bad_token",
			wantErr: fmt.Errorf("验证失败: %v: token contains an invalid number of segments",
				jwt.ErrTokenMalformed),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j.nowFunc = tt.nowFunc
			got, err := j.Verify(tt.token)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithSigningMethod(t *testing.T) {
	type jwtT = jwt.RegisteredClaims
	type testCase[T jwt.Claims] struct {
		name   string
		method jwt.SigningMethod
		want   jwt.SigningMethod
	}
	tests := []testCase[jwtT]{
		{
			name:   "设置成功",
			method: jwt.SigningMethodHS512,
			want:   jwt.SigningMethodHS512,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewJWTToken[jwtT]("foo", WithSigningMethod[jwtT](tt.method)).method
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithDecryptKey(t *testing.T) {
	type jwtT = jwt.RegisteredClaims
	type testCase[T jwt.Claims] struct {
		name       string
		decryptKey string
		want       string
	}
	tests := []testCase[jwtT]{
		{
			name:       "设置解密密钥成功",
			decryptKey: "decryptKey",
			want:       "decryptKey",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewJWTToken[jwtT]("foo", WithDecryptKey[jwtT](tt.decryptKey)).decryptKey
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithNowFunc(t *testing.T) {
	type jwtT = jwt.RegisteredClaims
	type testCase[T jwt.Claims] struct {
		name    string
		nowFunc func() time.Time
		want    time.Time
	}
	tests := []testCase[jwtT]{
		{
			name: "设置解密密钥成功",
			nowFunc: func() time.Time {
				return time.UnixMilli(1695571200000)
			},
			want: time.UnixMilli(1695571200000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewJWTToken[jwtT]("foo",
				WithNowFunc[jwtT](tt.nowFunc)).nowFunc()
			assert.Equal(t, tt.want, got)
		})
	}
}
