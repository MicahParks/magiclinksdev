package magiclink

import (
	"context"
	"net/http"
)

// RedirectorArgs are passed to a Redirector when performing a redirect.
type RedirectorArgs[CustomReadResponse any] struct {
	ReadAndExpireLink func(ctx context.Context, secret string) (jwtB64 string, response ReadResponse[CustomReadResponse], err error)
	Request           *http.Request
	Secret            string
	Writer            http.ResponseWriter
}

// Redirector is a custom implementation of redirecting a user to a magic link target.
type Redirector[CustomReadResponse any] interface {
	Redirect(args RedirectorArgs[CustomReadResponse])
}
