package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/ecodeclub/ginx/middlewares/token"
)

type myClaims struct {
	Foo string `json:"foo"`
	jwt.RegisteredClaims
}

var jwtToken = token.NewJWTToken[myClaims]("foo")

func TestNewAuthHandler(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name  string
		token token.Token[T]
		opts  []authHdlOption[T]
		want  Handler[T]
	}
	tests := []testCase[myClaims]{
		{
			name:  "normal_default_creates",
			token: jwtToken,
			opts:  []authHdlOption[myClaims]{},
			want: &authHandler[myClaims]{
				allowTokenHeader:    "authorization",
				bearerPrefix:        "Bearer",
				claimsCTXKey:        "claims",
				exposeAccessHeader:  "x-access-token",
				exposeRefreshHeader: "x-refresh-token",
				token:               jwtToken,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAuthHandler(tt.token, tt.opts...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithAllowTokenHeader(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name   string
		header string
		want   Handler[T]
	}
	tests := []testCase[myClaims]{
		{
			name:   "normal_set_allow_token_handler",
			header: "auth",
			want: &authHandler[myClaims]{
				allowTokenHeader:    "auth",
				bearerPrefix:        "Bearer",
				claimsCTXKey:        "claims",
				exposeAccessHeader:  "x-access-token",
				exposeRefreshHeader: "x-refresh-token",
				token:               jwtToken,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAuthHandler[myClaims](jwtToken,
				WithAllowTokenHeader[myClaims](tt.header))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithBearerPrefix(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name   string
		prefix string
		want   Handler[T]
	}
	tests := []testCase[myClaims]{
		{
			name:   "normal_set_bearer_prefix",
			prefix: "jwt",
			want: &authHandler[myClaims]{
				allowTokenHeader:    "authorization",
				bearerPrefix:        "jwt",
				claimsCTXKey:        "claims",
				exposeAccessHeader:  "x-access-token",
				exposeRefreshHeader: "x-refresh-token",
				token:               jwtToken,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAuthHandler[myClaims](jwtToken,
				WithBearerPrefix[myClaims](tt.prefix))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithClaimsCTXKey(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name         string
		claimsCTXKey string
		want         Handler[T]
	}
	tests := []testCase[myClaims]{
		{
			name:         "normal_set_claims_ctx_key",
			claimsCTXKey: "clm",
			want: &authHandler[myClaims]{
				allowTokenHeader:    "authorization",
				bearerPrefix:        "Bearer",
				claimsCTXKey:        "clm",
				exposeAccessHeader:  "x-access-token",
				exposeRefreshHeader: "x-refresh-token",
				token:               jwtToken,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAuthHandler[myClaims](jwtToken,
				WithClaimsCTXKey[myClaims](tt.claimsCTXKey))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithExposeAccessHeader(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name               string
		exposeAccessHeader string
		want               Handler[T]
	}
	tests := []testCase[myClaims]{
		{
			name:               "normal_set_expose_access_header",
			exposeAccessHeader: "access",
			want: &authHandler[myClaims]{
				allowTokenHeader:    "authorization",
				bearerPrefix:        "Bearer",
				claimsCTXKey:        "claims",
				exposeAccessHeader:  "access",
				exposeRefreshHeader: "x-refresh-token",
				token:               jwtToken,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAuthHandler[myClaims](jwtToken,
				WithExposeAccessHeader[myClaims](tt.exposeAccessHeader))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithExposeRefreshHeader(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name                string
		exposeRefreshHeader string
		want                Handler[T]
	}
	tests := []testCase[myClaims]{
		{
			name:                "normal_set_expose_refresh_Header",
			exposeRefreshHeader: "refresh",
			want: &authHandler[myClaims]{
				allowTokenHeader:    "authorization",
				bearerPrefix:        "Bearer",
				claimsCTXKey:        "claims",
				exposeAccessHeader:  "x-access-token",
				exposeRefreshHeader: "refresh",
				token:               jwtToken,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAuthHandler[myClaims](jwtToken,
				WithExposeRefreshHeader[myClaims](tt.exposeRefreshHeader))
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_authHandler_ExtractTokenString(t *testing.T) {
	a := NewAuthHandler[myClaims](jwtToken)
	type testCase[T jwt.Claims] struct {
		name       string
		reqBuilder func(t *testing.T) *http.Request
		want       string
	}
	tests := []testCase[myClaims]{
		{
			name: "normal_extract_token",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.a1q3jHKedQGbA-Zrn6S21QUpI2ZNYNHoeG5LkxAXRJQ")
				return req
			},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.a1q3jHKedQGbA-Zrn6S21QUpI2ZNYNHoeG5LkxAXRJQ",
		},
		{
			name: "bad_extract_token",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.a1q3jHKedQGbA-Zrn6S21QUpI2ZNYNHoeG5LkxAXRJQ")
				return req
			},
			want: "",
		},
		{
			name: "header_value_not_found",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = tt.reqBuilder(t)

			got := a.ExtractTokenString(ctx)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_authHandler_SetAccessToken(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name     string
		jwtToken token.Token[T]
		claims   T
		want     string
		wantErr  error
	}
	tests := []testCase[myClaims]{
		{
			name: "normal_set_access_token",
			jwtToken: token.NewJWTToken[myClaims]("foo",
				token.WithNowFunc[myClaims](func() time.Time {
					return time.UnixMilli(1695571200000)
				})),
			claims: myClaims{
				Foo: "bar",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  "bar",
					Subject: "1",
					IssuedAt: jwt.NewNumericDate(
						time.UnixMilli(1695571200000)),
					ExpiresAt: jwt.NewNumericDate(
						time.UnixMilli(1695571200000).
							Add(10 * time.Minute)),
				},
			},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.DkipgOka6QyyyvhW3IKLTnnWDQVuTBeGO5vb3Poj7ZY",
		},
		{
			name: "bad_claims",
			jwtToken: token.NewJWTToken[myClaims]("foo",
				token.WithSigningMethod[myClaims](jwt.SigningMethodRS512)),
			claims: myClaims{
				Foo: "bar",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  "bar",
					Subject: "1",
					IssuedAt: jwt.NewNumericDate(
						time.UnixMilli(1695571200000)),
					ExpiresAt: jwt.NewNumericDate(
						time.UnixMilli(1695571000000)),
				},
			},
			wantErr: errors.New("key is invalid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAuthHandler[myClaims](tt.jwtToken)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			req, err := http.NewRequest(http.MethodGet, "", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx.Request = req

			err = a.SetAccessToken(ctx, tt.claims)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want,
				recorder.Header().Get("x-access-token"))
		})
	}
}

func Test_authHandler_SetRefreshToken(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name     string
		jwtToken token.Token[T]
		claims   T
		want     string
		wantErr  error
	}
	tests := []testCase[myClaims]{
		{
			name: "normal_set_refresh_token",
			jwtToken: token.NewJWTToken[myClaims]("foo",
				token.WithNowFunc[myClaims](func() time.Time {
					return time.UnixMilli(1695571200000)
				})),
			claims: myClaims{
				Foo: "bar",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  "bar",
					Subject: "2",
					IssuedAt: jwt.NewNumericDate(
						time.UnixMilli(1695571200000)),
					ExpiresAt: jwt.NewNumericDate(
						time.UnixMilli(1695571200000).
							Add(10 * time.Minute)),
				},
			},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJiYXIiLCJzdWIiOiIyIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.gc4-zm430YUBtQIJi07uAxMiMCG1tclhOODNM20fZlM",
		},
		{
			name: "bad_claims",
			jwtToken: token.NewJWTToken[myClaims]("foo",
				token.WithSigningMethod[myClaims](jwt.SigningMethodRS512)),
			claims: myClaims{
				Foo: "bar",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  "bar",
					Subject: "2",
					IssuedAt: jwt.NewNumericDate(
						time.UnixMilli(1695571200000)),
					ExpiresAt: jwt.NewNumericDate(
						time.UnixMilli(1695571000000)),
				},
			},
			wantErr: errors.New("key is invalid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAuthHandler[myClaims](tt.jwtToken)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			req, err := http.NewRequest(http.MethodGet, "", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx.Request = req

			err = a.SetRefreshToken(ctx, tt.claims)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want,
				recorder.Header().Get("x-refresh-token"))
		})
	}
}

func Test_authHandler_VerifyToken(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name     string
		jwtToken token.Token[T]
		token    string
		want     T
		wantErr  error
	}
	tests := []testCase[myClaims]{
		{
			name: "normal_set_claims",
			jwtToken: token.NewJWTToken[myClaims]("foo",
				token.WithNowFunc[myClaims](func() time.Time {
					return time.UnixMilli(1695571500000)
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.DkipgOka6QyyyvhW3IKLTnnWDQVuTBeGO5vb3Poj7ZY",
			want: myClaims{
				Foo: "bar",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  "bar",
					Subject: "1",
					IssuedAt: jwt.NewNumericDate(
						time.UnixMilli(1695571200000)),
					ExpiresAt: jwt.NewNumericDate(
						time.UnixMilli(1695571200000).
							Add(10 * time.Minute)),
				},
			},
		},
		{
			name: "token_expired",
			jwtToken: token.NewJWTToken[myClaims]("foo",
				token.WithNowFunc[myClaims](func() time.Time {
					return time.UnixMilli(1695572500000)
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.DkipgOka6QyyyvhW3IKLTnnWDQVuTBeGO5vb3Poj7ZY",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired)),
		},
		{
			name: "wrong_signature",
			jwtToken: token.NewJWTToken[myClaims]("foo",
				token.WithNowFunc[myClaims](func() time.Time {
					return time.UnixMilli(1695571500000)
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.5AgQBNdf08M3vUIi_N2fVQrlNdrIbMvRw-8smkXATWc",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenSignatureInvalid, jwt.ErrSignatureInvalid)),
		},
		{
			name: "bad_token",
			jwtToken: token.NewJWTToken[myClaims]("foo",
				token.WithNowFunc[myClaims](func() time.Time {
					return time.UnixMilli(1695571500000)
				}),
			),
			token: "bad_token",
			wantErr: fmt.Errorf("验证失败: %v: token contains an invalid number of segments",
				jwt.ErrTokenMalformed),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAuthHandler[myClaims](tt.jwtToken)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			err := a.VerifyToken(ctx, tt.token)
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			claims, ok := ctx.Get("claims")
			if !ok {
				t.Errorf("claims 设置失败")
			}
			clm, ok := claims.(myClaims)
			if !ok {
				t.Errorf("claims 类型错误")
			}
			assert.Equal(t, tt.want, clm)
		})
	}
}
