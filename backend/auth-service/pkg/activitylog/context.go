package activitylog

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Actor is the authenticated caller of a request.
type Actor struct {
	UserID   string
	Username string
}

// RequestMeta is the client address and user agent of a request.
type RequestMeta struct {
	IPAddress string
	UserAgent string
}

type actorKey struct{}
type metaKey struct{}

// WithActor returns ctx carrying the authenticated caller.
func WithActor(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, a)
}

// ActorFrom returns the caller stored by WithActor.
func ActorFrom(ctx context.Context) (Actor, bool) {
	if ctx == nil {
		return Actor{}, false
	}
	a, ok := ctx.Value(actorKey{}).(Actor)
	return a, ok && a.UserID != ""
}

// WithRequestMeta returns ctx carrying the client address and user agent.
func WithRequestMeta(ctx context.Context, m RequestMeta) context.Context {
	return context.WithValue(ctx, metaKey{}, m)
}

// RequestMetaFrom returns the values stored by WithRequestMeta.
func RequestMetaFrom(ctx context.Context) RequestMeta {
	if ctx == nil {
		return RequestMeta{}
	}
	m, _ := ctx.Value(metaKey{}).(RequestMeta)
	return m
}

// ContextMiddleware stores the client IP and user agent in the request
// context, so write points deep in a service can attribute an event without
// threading them through every signature. It uses gin's ClientIP (trusted
// proxy aware), not a raw X-Forwarded-For header.
func ContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(WithRequestMeta(c.Request.Context(), RequestMeta{
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}))
		c.Next()
	}
}
