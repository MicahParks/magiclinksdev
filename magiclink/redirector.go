package magiclink

import (
	"net/http"
	"net/url"
)

type RedirectArgs[CustomCreateArgs, CustomReadResponse any] struct {
	ReadResponse ReadResponse[CustomCreateArgs, CustomReadResponse]
	RedirectURL  *url.URL // TODO Change type to a type that marks magic link as used when read.
	Request      *http.Request
	Writer       http.ResponseWriter
}

type Redirector[CustomCreateArgs, CustomReadResponse any] interface {
	Redirect(args RedirectArgs[CustomCreateArgs, CustomReadResponse])
}
