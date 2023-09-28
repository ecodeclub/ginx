package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/ecodeclub/ginx/middlewares/token"
)

var authHdl = NewAuthHandler[myClaims](jwtToken)

func TestNewJWTBuilder(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name    string
		handler Handler[T]
		want    *JWTBuilder[T]
	}
	tests := []testCase[myClaims]{
		{
			name:    "normal",
			handler: authHdl,
			want: &JWTBuilder[myClaims]{
				Handler: authHdl,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewJWTBuilder[myClaims](tt.handler)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithIgnorePaths(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name  string
		paths []string
		want  func() *JWTBuilder[T]
	}
	tests := []testCase[myClaims]{
		{
			name: "normal",
			paths: []string{
				"/login",
				"/signup",
			},
			want: func() *JWTBuilder[myClaims] {
				pathSet := set.NewMapSet[string](2)
				pathSet.Add("/login")
				pathSet.Add("/signup")

				return &JWTBuilder[myClaims]{
					publicPaths: pathSet,
					Handler:     authHdl,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewJWTBuilder(authHdl,
				WithIgnorePaths[myClaims](tt.paths...))
			assert.Equal(t, tt.want(), got)
		})
	}
}

func TestJWTBuilder_Build(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name       string
		b          *JWTBuilder[T]
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
	}
	tests := []testCase[myClaims]{
		{
			name: "normal",
			b: NewJWTBuilder[myClaims](
				NewAuthHandler[myClaims](
					token.NewJWTToken[myClaims]("foo",
						token.WithNowFunc[myClaims](func() time.Time {
							return time.UnixMilli(1695571500000)
						})))),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.DkipgOka6QyyyvhW3IKLTnnWDQVuTBeGO5vb3Poj7ZY")
				return req
			},
			wantCode: http.StatusOK,
		},
		{
			name: "verification_failed",
			b:    NewJWTBuilder[myClaims](authHdl),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.DkipgOka6QyyyvhW3IKLTnnWDQVuTBeGO5vb3Poj7ZY")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "extract_token_failed",
			b: NewJWTBuilder[myClaims](
				NewAuthHandler[myClaims](
					token.NewJWTToken[myClaims]("foo",
						token.WithNowFunc[myClaims](func() time.Time {
							return time.UnixMilli(1695571500000)
						})))),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJiYXIiLCJzdWIiOiIxIiwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.DkipgOka6QyyyvhW3IKLTnnWDQVuTBeGO5vb3Poj7ZY")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "verification_failed",
			b: NewJWTBuilder[myClaims](authHdl,
				WithIgnorePaths[myClaims]("/login")),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/login", nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := gin.Default()
			server.Use(tt.b.Build())
			tt.b.registerRoutes(server)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tt.wantCode, recorder.Code)
		})
	}
}

func (b *JWTBuilder[T]) registerRoutes(server *gin.Engine) {
	server.GET("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	server.GET("/login", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}
