package handle_test

import (
	"testing"

	"github.com/MicahParks/magiclinksdev/handle"
)

func TestMiddlewareHookFunc(t *testing.T) {
	const pathChange = "test"
	hook := handle.MiddlewareHookFunc(func(options handle.MiddlewareOptions) handle.MiddlewareOptions {
		options.Path = pathChange
		return options
	})

	options := hook.Hook(handle.MiddlewareOptions{})
	if options.Path != pathChange {
		t.Fatalf("Failed to set option path.")
	}
}
