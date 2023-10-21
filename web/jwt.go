package web

import (
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
