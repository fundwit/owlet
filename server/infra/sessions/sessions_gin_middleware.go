package sessions

import (
	"owlet/server/infra/fail"
	"owlet/server/infra/meta"

	"github.com/gin-gonic/gin"
)

// SessionFilter using token from cookie to find the cached authentication info,
// then inject the valid authentication info into gin context.
func SessionFilter() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie(KeySecToken)
		if err != nil {
			panic(fail.ErrUnauthenticated)
		}
		securityContextValue, found := TokenCache.Get(token)
		if !found {
			panic(fail.ErrUnauthenticated)
		}
		secCtx, ok := securityContextValue.(*Session)
		if !ok {
			panic(fail.ErrUnauthenticated)
		}
		InjectSessionIntoGinContext(ctx, secCtx)
		ctx.Next()
	}
}

func SessionTokenAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie(KeySecToken)
		if err == nil && token == meta.Config.AdminSecret {
			ctx.Set(KeySecCtx, &Session{
				Token:    "hidden",
				Identity: Identity{ID: 1, Name: meta.Config.AdminName, Nickname: meta.Config.AdminName},
			})
		}
	}
}
