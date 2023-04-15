package magiclink

import (
	"context"
	"net/http"
)

type RedirectorArgs[CustomCreateArgs, CustomReadResponse, CustomKeyMeta any] struct {
	ReadAndExpireLink func(ctx context.Context, secret string) (jwtB64 string, response ReadResponse[CustomCreateArgs, CustomReadResponse], err error)
	Request           *http.Request
	Secret            string
	Writer            http.ResponseWriter
}

type Redirector[CustomCreateArgs, CustomReadResponse, CustomKeyMeta any] interface {
	Redirect(args RedirectorArgs[CustomCreateArgs, CustomReadResponse, CustomKeyMeta])
}
