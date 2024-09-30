package storage

import (
	"time"
)

// ReadSigningKeyOptions are the options for the ReadSigningKey method.
type ReadSigningKeyOptions struct {
	JWTAlg string
}

// MagicLinkCustomReadResponse is the custom response for reading a magic link.
type MagicLinkCustomReadResponse struct {
	Visited *time.Time
}
