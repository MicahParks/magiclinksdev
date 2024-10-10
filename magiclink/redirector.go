package magiclink

import (
	"context"
	"net/http"
)

// RedirectorParams are passed to a Redirector when performing a redirect.
type RedirectorParams struct {
	ReadAndExpireLink func(ctx context.Context, secret string) (jwtB64 string, response ReadResult, err error)
	Request           *http.Request
	Secret            string
	Writer            http.ResponseWriter
}

// Redirector is a custom implementation of redirecting a user to a magic link target.
type Redirector interface {
	Redirect(args RedirectorParams)
}
