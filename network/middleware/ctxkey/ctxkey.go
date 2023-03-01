package ctxkey

import (
	"errors"
)

// ErrCtxKey is the error returned when a context key is not as expected.
var ErrCtxKey = errors.New("context key was not as expected")

const (
	// RequestUUID is the context key for the request's UUID value.
	RequestUUID ContextKey = iota
	// ServiceAccount is the context key for the request's service account value.
	ServiceAccount
	// Sugared is the context key for the request's zap sugared logger value.
	Sugared
	// Tx is the context key for the request's transaction value.
	Tx
)

// ContextKey is a type for context value keys.
type ContextKey int
