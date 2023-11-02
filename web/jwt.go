package web

import (
	jwt2 "github.com/iris-contrib/middleware/jwt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
	"nw-guardian/internal/auth"
	"os"
)

// GetClaims Define a function to verify the session token
func GetClaims(ctx iris.Context) *auth.Session {
	claims := jwt.Get(ctx).(*auth.Session)
	return claims
}

func VerifySession() iris.Handler {
	secret := os.Getenv("KEY")
	verifier := jwt.NewVerifier(jwt.HS256, []byte(secret))
	verifier.Extractors = []jwt.TokenExtractor{jwt.FromHeader} // extract token only from Authorization: Bearer $token

	return verifier.Verify(func() interface{} {
		return new(auth.Session)
	})
}

func GetWebSocketJWT() *jwt2.Middleware {
	secret := os.Getenv("KEY")
	j := jwt2.New(jwt2.Config{
		// Extract by the "token" url,
		// so the client should dial with ws://localhost:8080/echo?token=$token
		Extractor: jwt2.FromParameter("token"),

		ValidationKeyGetter: func(token *jwt2.Token) (interface{}, error) {
			return []byte(secret), nil
		},

		// When set, the middleware verifies that tokens are signed
		// with the specific signing algorithm
		// If the signing method is not constant the
		// `Config.ValidationKeyGetter` callback field can be used
		// to implement additional checks
		// Important to avoid security issues described here:
		// https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
		SigningMethod: jwt2.SigningMethodHS256,
	})

	return j
}
