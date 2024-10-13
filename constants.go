package magiclinksdev

import (
	"errors"
	"time"
)

const (
	// ContentTypeJSON is the content type for JSON.
	ContentTypeJSON = "application/json"
	// DefaultRelativePathRedirect is the default relative path for redirecting.
	DefaultRelativePathRedirect = "redirect"
	// HeaderContentType is the content type header.
	HeaderContentType = "Content-Type"
	// LogFmt is the log format.
	LogFmt = "%s\nError: %s"
	// LogErr is the log error.
	LogErr = "error"
	// LogRequestBody is key for logging the request body.
	LogRequestBody = "requestBody"
	// LogResponseBody is key for logging the response body.
	LogResponseBody = "responseBody"
	// Over250Years is the maximum duration for this project. Restriction derived from Golang's time.Duration.
	Over250Years = 250 * 366 * 24 * time.Hour
	// ResponseInternalServerError is the response for internal server errors.
	ResponseInternalServerError = "Internal server error."
	// ResponseTooManyRequests is the response for too many requests.
	ResponseTooManyRequests = "Too many requests."
	// ResponseUnauthorized is the response for unauthorized requests.
	ResponseUnauthorized = "Unauthorized."
)

var (
	ErrParams = errors.New("invalid parameters")
)

func Ptr[T any](v T) *T {
	return &v
}
