package sessions

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

const TokenExpiration = 24 * time.Hour

var TokenCache = cache.New(TokenExpiration, 1*time.Minute)

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

const KeySecCtx = "SecCtx"
const KeySecToken = "sec_token"

func ExtractSessionFromGinContext(ctx *gin.Context) *Session {
	value, found := ctx.Get(KeySecCtx)
	if !found {
		return &GuestSession
	}
	s0, ok := value.(*Session)
	if !ok || s0.Token == "" {
		return &GuestSession
	}
	s := s0.Clone()
	s.Context = ctx.Request.Context() // trace context
	return &s
}

func InjectSessionIntoGinContext(ctx *gin.Context, s *Session) {
	if s != nil && s.Token != "" {
		ctx.Set(KeySecCtx, s)
	}
}
