package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/ecodeclub/ginx/middlewares/auth"
	"github.com/ecodeclub/ginx/middlewares/token"
)

type myClaims struct {
	Foo string `json:"foo"`
	jwt.RegisteredClaims
}

func Test_token_Refresh(t *testing.T) {
	nowTime := time.UnixMilli(1695571500000)
	type testCase[T jwt.Claims] struct {
		name         string
		hdl          auth.Handler[T]
		reqBuilder   func(t *testing.T) *http.Request
		accessClaims T
		wantCode     int
		wantToken    string
	}
	tests := []testCase[myClaims]{
		{
			name: "normal",
			hdl: auth.NewAuthHandler[myClaims](
				token.NewJWTToken[myClaims]("access-token-key",
					token.WithNowFunc[myClaims](
						func() time.Time {
							return nowTime
						},
					),
					token.WithDecryptKey[myClaims]("refresh-token-key"),
				),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJyZWZyZXNoIiwic3ViIjoiMSIsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.8_LyHqansmkqcXJ1INVJDPI2XUAzd12keCrSltqnCJQ")
				return req
			},
			accessClaims: myClaims{
				Foo: "bar",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:   "access",
					Subject:  "1",
					IssuedAt: jwt.NewNumericDate(nowTime),
					ExpiresAt: jwt.NewNumericDate(
						nowTime.Add(10 * time.Minute)),
				},
			},
			wantCode:  http.StatusOK,
			wantToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJhY2Nlc3MiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcyMTAwLCJpYXQiOjE2OTU1NzE1MDB9.rE74rZg00AtSwvFpVMMYQggfPpgsrK6oiil3PjKKpcA",
		},
		{
			name: "set_access_token_failed",
			hdl: auth.NewAuthHandler[myClaims](
				token.NewJWTToken[myClaims]("access-token-key",
					token.WithNowFunc[myClaims](
						func() time.Time {
							return nowTime
						},
					),
					token.WithSigningMethod[myClaims](jwt.SigningMethodRS256),
					token.WithDecryptKey[myClaims]("refresh-token-key"),
				),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJyZWZyZXNoIiwic3ViIjoiMSIsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.8_LyHqansmkqcXJ1INVJDPI2XUAzd12keCrSltqnCJQ")
				return req
			},
			accessClaims: myClaims{
				Foo: "bar",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:   "access",
					Subject:  "1",
					IssuedAt: jwt.NewNumericDate(nowTime),
					ExpiresAt: jwt.NewNumericDate(
						nowTime.Add(10 * time.Minute)),
				},
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "verify_failed",
			hdl: auth.NewAuthHandler[myClaims](
				token.NewJWTToken[myClaims]("access-token-key",
					token.WithNowFunc[myClaims](
						func() time.Time {
							return nowTime
						},
					),
					token.WithDecryptKey[myClaims]("mistake-refresh-token-key"),
				),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJyZWZyZXNoIiwic3ViIjoiMSIsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.8_LyHqansmkqcXJ1INVJDPI2XUAzd12keCrSltqnCJQ")
				return req
			},
			accessClaims: myClaims{
				Foo: "bar",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:   "access",
					Subject:  "1",
					IssuedAt: jwt.NewNumericDate(nowTime),
					ExpiresAt: jwt.NewNumericDate(
						nowTime.Add(10 * time.Minute)),
				},
			},
			wantCode: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewTokenHandler[myClaims](tt.accessClaims, tt.hdl)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = tt.reqBuilder(t)
			svc.Refresh(ctx)
			assert.Equal(t, tt.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}
			assert.Equal(t, tt.wantToken,
				recorder.Header().Get("x-access-token"))
		})
	}
}

func (t *tokenHandler[T]) registerRoutes(server *gin.Engine) {
	server.GET("/refresh", t.Refresh)
}
