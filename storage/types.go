package storage

import (
	"time"
)

// JWKSetCustomKeyMeta is the custom metadata for a JWKSet key.
type JWKSetCustomKeyMeta struct {
	SigningDefault bool
}

// MagicLinkCustomCreateArgs is the custom arguments for creating a magic link.
type MagicLinkCustomCreateArgs struct {
	Expires time.Time
}

// MagicLinkCustomReadResponse is the custom response for reading a magic link.
type MagicLinkCustomReadResponse struct {
	Visited *time.Time
}
