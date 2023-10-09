package jwt

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type data struct {
	Foo string `json:"foo"`
}

var (
	defaultExpire = 10 * time.Minute
	defaultClaims = RegisteredClaims[data]{
		Data: data{Foo: "1"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(nowTime.Add(defaultExpire)),
			NotBefore: jwt.NewNumericDate(nowTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
		},
	}
	encryptionKey = "sign key"
	nowTime       = time.UnixMilli(1695571200000)
	defaultOption = &Options{
		Expire:        defaultExpire,
		EncryptionKey: encryptionKey,
		DecryptKey:    encryptionKey,
		Method:        jwt.SigningMethodHS256,
	}
	defaultManager = NewManager[data](defaultOption,
		WithNowFunc[data](func() time.Time {
			return nowTime
		}),
	)
)

func TestManager_GenerateAccessToken(t *testing.T) {
	m := defaultManager
	type testCase[T any] struct {
		name    string
		data    T
		want    string
		wantErr error
	}
	tests := []testCase[data]{
		{
			name: "normal",
			data: data{Foo: "1"},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.UNuVOmAwgR-atNOMVi9JldtT7qGl7LCFuyq4uiYgg_Y",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.GenerateAccessToken(tt.data)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_GenerateRefreshToken(t *testing.T) {
	m := defaultManager
	type testCase[T any] struct {
		name              string
		refreshJWTOptions *Options
		data              T
		want              string
		wantErr           error
	}
	tests := []testCase[data]{
		{
			name: "normal",
			refreshJWTOptions: &Options{
				Expire:        24 * 60 * time.Minute,
				EncryptionKey: "refresh sign key",
				DecryptKey:    "refresh sign key",
				Method:        jwt.SigningMethodHS256,
			},
			data: data{Foo: "1"},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.yb0pocXbtJuZziA6Ugs3wcYOAslrIk1-C_NpKgTrNVw",
		},
		{
			name:    "mistake",
			data:    data{Foo: "1"},
			want:    "",
			wantErr: ErrEmptyRefreshOpts,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.refreshJWTOptions = tt.refreshJWTOptions
			got, err := m.GenerateRefreshToken(tt.data)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_MiddlewareBuilder(t *testing.T) {
	type testCase[T any] struct {
		name       string
		m          *Manager[T]
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
	}
	tests := []testCase[data]{
		{
			// 验证失败
			name: "verify_failed",
			m: NewManager[data](defaultOption,
				WithIgnorePaths[data]("/login")),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.UNuVOmAwgR-atNOMVi9JldtT7qGl7LCFuyq4uiYgg_Y")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			// 提取 token 失败
			name: "extract_token_failed",
			m: NewManager[data](defaultOption,
				WithIgnorePaths[data]("/login")),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer ")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			// 无需认证直接通过
			name: "pass_without_authentication",
			m: NewManager[data](defaultOption,
				WithIgnorePaths[data]("/login")),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/login", nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
		},
		{
			// 验证通过
			name: "pass_the_verification",
			m: NewManager[data](defaultOption,
				WithIgnorePaths[data]("/login"),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695571500000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.UNuVOmAwgR-atNOMVi9JldtT7qGl7LCFuyq4uiYgg_Y")
				return req
			},
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := gin.Default()
			server.Use(tt.m.MiddlewareBuilder())
			tt.m.registerRoutes(server)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tt.wantCode, recorder.Code)
		})
	}
}

func TestManager_Refresh(t *testing.T) {
	type testCase[T any] struct {
		name             string
		m                *Manager[T]
		reqBuilder       func(t *testing.T) *http.Request
		wantCode         int
		wantAccessToken  string
		wantRefreshToken string
	}
	tests := []testCase[data]{
		{
			// 更新资源令牌并轮换刷新令牌
			name: "refresh_access_token_and_rotate_refresh_token",
			m: NewManager[data](defaultOption,
				WithRefreshJWTOptions[data](
					NewOptions(24*60*time.Minute,
						"refresh sign key",
					)),
				WithRotateRefreshToken[data](true),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.yb0pocXbtJuZziA6Ugs3wcYOAslrIk1-C_NpKgTrNVw")
				return req
			},
			wantCode:         http.StatusOK,
			wantAccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjIzNjAwLCJuYmYiOjE2OTU2MjMwMDAsImlhdCI6MTY5NTYyMzAwMH0.5Hv-Gq8RW0xAFBh4WhKc0KDLsdgTEv3RUhPceaM4e5M",
			wantRefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NzA5NDAwLCJuYmYiOjE2OTU2MjMwMDAsImlhdCI6MTY5NTYyMzAwMH0.4R-JmqcKHtsoFOGFDe5SBA2wNV0F-XvnP2Janp6NfZY",
		},
		{
			// 更新资源令牌但轮换刷新令牌生成失败
			name: "refresh_access_token_but_gen_rotate_refresh_token_failed",
			m: NewManager[data](defaultOption,
				WithRefreshJWTOptions[data](
					NewOptions(24*60*time.Minute,
						"refresh sign key",
						WithMethod(jwt.SigningMethodRS256),
					)),
				WithRotateRefreshToken[data](true),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.yb0pocXbtJuZziA6Ugs3wcYOAslrIk1-C_NpKgTrNVw")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			// 更新资源令牌
			name: "refresh_access_token",
			m: NewManager[data](defaultOption,
				WithRefreshJWTOptions[data](
					NewOptions(24*60*time.Minute,
						"refresh sign key",
					)),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.yb0pocXbtJuZziA6Ugs3wcYOAslrIk1-C_NpKgTrNVw")
				return req
			},
			wantCode:        http.StatusOK,
			wantAccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjIzNjAwLCJuYmYiOjE2OTU2MjMwMDAsImlhdCI6MTY5NTYyMzAwMH0.5Hv-Gq8RW0xAFBh4WhKc0KDLsdgTEv3RUhPceaM4e5M",
		},
		{
			// 生成资源令牌失败
			name: "gen_access_token_failed",
			m: NewManager[data](
				&Options{
					Expire:        10 * time.Minute,
					EncryptionKey: encryptionKey,
					DecryptKey:    encryptionKey,
					Method:        jwt.SigningMethodRS256,
				},
				WithRefreshJWTOptions[data](
					NewOptions(24*60*time.Minute,
						"refresh sign key",
					)),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.yb0pocXbtJuZziA6Ugs3wcYOAslrIk1-C_NpKgTrNVw")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			// 刷新令牌认证失败
			name: "refresh_token_verify_failed",
			m: NewManager[data](
				defaultOption,
				WithRefreshJWTOptions[data](
					NewOptions(24*60*time.Minute,
						"refresh sign key",
					)),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695723000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.yb0pocXbtJuZziA6Ugs3wcYOAslrIk1-C_NpKgTrNVw")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			// 没有设置刷新令牌选项
			name: "not_set_refreshJWTOptions",
			m: NewManager[data](
				defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695723000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.yb0pocXbtJuZziA6Ugs3wcYOAslrIk1-C_NpKgTrNVw")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := gin.Default()
			tt.m.registerRoutes(server)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tt.wantCode, recorder.Code)
			if tt.wantCode != http.StatusOK {
				return
			}
			assert.Equal(t, tt.wantAccessToken,
				recorder.Header().Get("x-access-token"))
			assert.Equal(t, tt.wantRefreshToken,
				recorder.Header().Get("x-refresh-token"))
		})
	}
}

func TestManager_VerifyAccessToken(t *testing.T) {
	type testCase[T any] struct {
		name    string
		m       *Manager[T]
		token   string
		want    RegisteredClaims[T]
		wantErr error
	}
	tests := []testCase[data]{
		{
			name:  "normal",
			m:     defaultManager,
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.UNuVOmAwgR-atNOMVi9JldtT7qGl7LCFuyq4uiYgg_Y",
			want:  defaultClaims,
		},
		{
			// token 过期了
			name: "token_expired",
			m: NewManager[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695671200000)
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.UNuVOmAwgR-atNOMVi9JldtT7qGl7LCFuyq4uiYgg_Y",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired)),
		},
		{
			// token 签名错误
			name: "bad_sign_key",
			m: NewManager[data](
				&Options{
					Expire:        defaultExpire,
					EncryptionKey: encryptionKey,
					DecryptKey:    "bad sign key",
					Method:        jwt.SigningMethodHS256,
				},
				WithNowFunc[data](func() time.Time {
					return nowTime
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.UNuVOmAwgR-atNOMVi9JldtT7qGl7LCFuyq4uiYgg_Y",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenSignatureInvalid, jwt.ErrSignatureInvalid)),
		},
		{
			// 错误的 token
			name:  "bad_token",
			m:     defaultManager,
			token: "bad_token",
			wantErr: fmt.Errorf("验证失败: %v: token contains an invalid number of segments",
				jwt.ErrTokenMalformed),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.VerifyAccessToken(tt.token)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_VerifyRefreshToken(t *testing.T) {
	type testCase[T any] struct {
		name    string
		m       *Manager[T]
		token   string
		want    RegisteredClaims[T]
		wantErr error
	}
	tests := []testCase[data]{
		{
			name: "normal",
			m: NewManager[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
				WithRefreshJWTOptions[data](&Options{
					Expire:        24 * 60 * time.Minute,
					EncryptionKey: "refresh sign key",
					DecryptKey:    "refresh sign key",
					Method:        jwt.SigningMethodHS256,
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.yb0pocXbtJuZziA6Ugs3wcYOAslrIk1-C_NpKgTrNVw",
			want: RegisteredClaims[data]{
				Data: data{Foo: "1"},
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(nowTime.Add(24 * 60 * time.Minute)),
					NotBefore: jwt.NewNumericDate(nowTime),
					IssuedAt:  jwt.NewNumericDate(nowTime),
				},
			},
		},
		{
			// token 过期了
			name: "token_expired",
			m: NewManager[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695701200000)
				}),
				WithRefreshJWTOptions[data](&Options{
					Expire:        24 * 60 * time.Minute,
					EncryptionKey: "refresh sign key",
					DecryptKey:    "refresh sign key",
					Method:        jwt.SigningMethodHS256,
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.yb0pocXbtJuZziA6Ugs3wcYOAslrIk1-C_NpKgTrNVw",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired)),
		},
		{
			// token 签名错误
			name: "bad_sign_key",
			m: NewManager[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
				WithRefreshJWTOptions[data](&Options{
					Expire:        24 * 60 * time.Minute,
					EncryptionKey: "bad refresh sign key",
					DecryptKey:    "bad refresh sign key",
					Method:        jwt.SigningMethodHS256,
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.yb0pocXbtJuZziA6Ugs3wcYOAslrIk1-C_NpKgTrNVw",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenSignatureInvalid, jwt.ErrSignatureInvalid)),
		},
		{
			// 错误的 token
			name: "bad_token",
			m: NewManager[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
				WithRefreshJWTOptions[data](&Options{
					Expire:        24 * 60 * time.Minute,
					EncryptionKey: "refresh sign key",
					DecryptKey:    "refresh sign key",
					Method:        jwt.SigningMethodHS256,
				}),
			),
			token: "bad_token",
			wantErr: fmt.Errorf("验证失败: %v: token contains an invalid number of segments",
				jwt.ErrTokenMalformed),
		},
		{
			name: "no_refresh_options",
			m: NewManager[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
			),
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.yb0pocXbtJuZziA6Ugs3wcYOAslrIk1-C_NpKgTrNVw",
			wantErr: ErrEmptyRefreshOpts,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.VerifyRefreshToken(tt.token)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_extractTokenString(t *testing.T) {
	m := defaultManager
	type header struct {
		key   string
		value string
	}
	type testCase[T any] struct {
		name   string
		header header
		want   string
	}
	tests := []testCase[data]{
		{
			name: "normal",
			header: header{
				key:   "authorization",
				value: "Bearer token",
			},
			want: "token",
		},
		{
			name: "mistake_prefix",
			header: header{
				key:   "authorization",
				value: "bearer token",
			},
		},
		{
			name: "no_allow_token_header",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			req, err := http.NewRequest(http.MethodGet, "", nil)
			req.Header.Add(tt.header.key, tt.header.value)
			if err != nil {
				t.Fatal(err)
			}
			ctx.Request = req

			got := m.extractTokenString(ctx)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_verifyTokenAndSetClm(t *testing.T) {
	type testCase[T any] struct {
		name    string
		m       *Manager[T]
		token   string
		want    RegisteredClaims[T]
		wantErr error
	}
	tests := []testCase[data]{
		{
			name:  "normal",
			m:     defaultManager,
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.UNuVOmAwgR-atNOMVi9JldtT7qGl7LCFuyq4uiYgg_Y",
			want:  defaultClaims,
		},
		{
			name: "verify_access_token_failed",
			m: NewManager[data](
				defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695671200000)
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJuYmYiOjE2OTU1NzEyMDAsImlhdCI6MTY5NTU3MTIwMH0.UNuVOmAwgR-atNOMVi9JldtT7qGl7LCFuyq4uiYgg_Y",
			want:  RegisteredClaims[data]{},
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			req, err := http.NewRequest(http.MethodGet, "", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx.Request = req
			err = tt.m.verifyTokenAndSetClm(ctx, tt.token)
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			v, ok := ctx.Get("claims")
			if !ok {
				t.Error("claims设置失败")
			}
			clm, ok := v.(RegisteredClaims[data])
			if !ok {
				t.Error("claims不是 RegisteredClaims[T] 类型")
			}
			assert.Equal(t, tt.want, clm)
		})
	}
}

func TestWithAllowTokenHeader(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() ManagerOption[T]
		want string
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() ManagerOption[data] {
				return nil
			},
			want: "authorization",
		},
		{
			name: "set_another_header",
			fn: func() ManagerOption[data] {
				return WithAllowTokenHeader[data]("jwt")
			},
			want: "jwt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.fn() == nil {
				got = NewManager[data](
					defaultOption,
				).allowTokenHeader
			} else {
				got = NewManager[data](
					defaultOption,
					tt.fn(),
				).allowTokenHeader
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithBearerPrefix(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() ManagerOption[T]
		want string
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() ManagerOption[data] {
				return nil
			},
			want: "Bearer",
		},
		{
			name: "set_another_prefix",
			fn: func() ManagerOption[data] {
				return WithBearerPrefix[data]("jwt")
			},
			want: "jwt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.fn() == nil {
				got = NewManager[data](
					defaultOption,
				).bearerPrefix
			} else {
				got = NewManager[data](
					defaultOption,
					tt.fn(),
				).bearerPrefix
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithClaimsCTXKey(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() ManagerOption[T]
		want string
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() ManagerOption[data] {
				return nil
			},
			want: "claims",
		},
		{
			name: "set_another_ctx_key",
			fn: func() ManagerOption[data] {
				return WithClaimsCTXKey[data]("clm")
			},
			want: "clm",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.fn() == nil {
				got = NewManager[data](
					defaultOption,
				).claimsCTXKey
			} else {
				got = NewManager[data](
					defaultOption,
					tt.fn(),
				).claimsCTXKey
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithExposeAccessHeader(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() ManagerOption[T]
		want string
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() ManagerOption[data] {
				return nil
			},
			want: "x-access-token",
		},
		{
			name: "set_another_header",
			fn: func() ManagerOption[data] {
				return WithExposeAccessHeader[data]("token")
			},
			want: "token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.fn() == nil {
				got = NewManager[data](
					defaultOption,
				).exposeAccessHeader
			} else {
				got = NewManager[data](
					defaultOption,
					tt.fn(),
				).exposeAccessHeader
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithExposeRefreshHeader(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() ManagerOption[T]
		want string
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() ManagerOption[data] {
				return nil
			},
			want: "x-refresh-token",
		},
		{
			name: "set_another_header",
			fn: func() ManagerOption[data] {
				return WithExposeRefreshHeader[data]("refresh-token")
			},
			want: "refresh-token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.fn() == nil {
				got = NewManager[data](
					defaultOption,
				).exposeRefreshHeader
			} else {
				got = NewManager[data](
					defaultOption,
					tt.fn(),
				).exposeRefreshHeader
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithIgnorePaths(t *testing.T) {
	type testCase[T any] struct {
		name  string
		fn    func() ManagerOption[T]
		paths []string
		want  []bool
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() ManagerOption[data] {
				return nil
			},
			want: []bool{},
		},
		{
			name: "all_exists_paths",
			fn: func() ManagerOption[data] {
				return WithIgnorePaths[data]([]string{
					"/login",
					"/signup",
				}...)
			},
			paths: []string{"/login", "/signup"},
			want:  []bool{true, true},
		},
		{
			name: "one_path_does_not_exist",
			fn: func() ManagerOption[data] {
				return WithIgnorePaths[data]([]string{
					"/login",
					"/signup",
				}...)
			},
			paths: []string{"/login", "/profile", "/signup"},
			want:  []bool{true, false, true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ignorePaths set.Set[string]
			if tt.fn() == nil {
				ignorePaths = NewManager[data](
					defaultOption,
				).publicPaths
			} else {
				ignorePaths = NewManager[data](
					defaultOption,
					tt.fn(),
				).publicPaths
			}
			exists := make([]bool, 0, len(tt.paths))
			for _, path := range tt.paths {
				exists = append(exists, ignorePaths.Exist(path))
			}
			assert.Equal(t, tt.want, exists)
		})
	}
}

func TestWithNowFunc(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() ManagerOption[T]
		want time.Time
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() ManagerOption[data] {
				return nil
			},
			want: time.Now(),
		},
		{
			name: "set_another_now_func",
			fn: func() ManagerOption[data] {
				return WithNowFunc[data](func() time.Time {
					return nowTime
				})
			},
			want: nowTime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got time.Time
			if tt.fn() == nil {
				got = NewManager[data](
					defaultOption,
				).nowFunc()
			} else {
				got = NewManager[data](
					defaultOption,
					tt.fn(),
				).nowFunc()
			}
			assert.Equal(t, tt.want.Unix(), got.Unix())
		})
	}
}

func TestWithRefreshJWTOptions(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() ManagerOption[T]
		want *Options
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() ManagerOption[data] {
				return nil
			},
			want: nil,
		},
		{
			name: "set_refresh_jwt_options",
			fn: func() ManagerOption[data] {
				return WithRefreshJWTOptions[data](
					NewOptions(
						24*60*time.Minute,
						"refresh sign key",
					),
				)
			},
			want: &Options{
				Expire:        24 * 60 * time.Minute,
				EncryptionKey: "refresh sign key",
				DecryptKey:    "refresh sign key",
				Method:        jwt.SigningMethodHS256,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *Options
			if tt.fn() == nil {
				got = NewManager[data](
					defaultOption,
				).refreshJWTOptions
			} else {
				got = NewManager[data](
					defaultOption,
					tt.fn(),
				).refreshJWTOptions
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func (m *Manager[T]) registerRoutes(server *gin.Engine) {
	server.GET("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	server.GET("/login", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	server.GET("refresh", m.Refresh)
}
