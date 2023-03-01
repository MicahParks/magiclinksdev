package jsontype

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	// errStringUnmarshal is an error message for when an expected JSON string could not be unmarshalled.
	errStringUnmarshal = "failed to unmarshal as a string"

	// errUnmarshalPackage is prepended to all unmarshal errors to make troubleshooting easier.
	errUnmarshalPackage = "github.com/MicahParks/jsontype JSON unmarshal error"

	// errUnreachableFmt is an error message for when code should be unreachable due to an unsupported type.
	errUnreachableFmt = "%s: (should be unreachable code in github.com/MicahParks/jsontype) unsupported type: %T"
)

// Options is a set of options for a JSONType. It modifies the behavior of JSON marshal/unmarshal.
type Options struct {
	MailAddressAddressOnlyMarshal bool
	MailAddressLowerMarshal       bool
	MailAddressUpperMarshal       bool
	TimeFormatMarshal             string
	TimeFormatUnmarshal           string
}

// J is a set of common Go types that can be marshaled and unmarshalled with this package.
type J interface {
	*mail.Address | *regexp.Regexp | time.Duration | time.Time | *url.URL
}

// JSONType holds a generic J value. It can be used to marshal and unmarshal its value to and from JSON.
type JSONType[T J] struct {
	options Options
	v       T
}

// New creates a new JSONType.
func New[T J](v T) *JSONType[T] {
	return &JSONType[T]{
		v: v,
	}
}

// NewWithOptions creates a new JSONType with options.
func NewWithOptions[T J](v T, options Options) *JSONType[T] {
	return &JSONType[T]{
		options: options,
		v:       v,
	}
}

// Get returns the held value.
func (j *JSONType[T]) Get() T {
	if j == nil {
		var t T
		return t
	}
	return j.v
}

// MarshalJSON helps implement the json.Marshaler interface.
func (j *JSONType[T]) MarshalJSON() ([]byte, error) {
	var s string
	switch v := any(j.Get()).(type) {
	case *mail.Address:
		if j.options.MailAddressAddressOnlyMarshal {
			s = v.Address
		} else {
			s = v.String()
		}
		if j.options.MailAddressLowerMarshal {
			s = strings.ToLower(s)
		} else if j.options.MailAddressUpperMarshal {
			s = strings.ToUpper(s)
		}
	case *regexp.Regexp:
		s = v.String()
	case time.Duration:
		s = v.String()
	case time.Time:
		format := time.RFC3339
		if j.options.TimeFormatMarshal != "" {
			format = j.options.TimeFormatMarshal
		}
		s = v.Format(format)
	case *url.URL:
		s = v.String()
	default:
		return nil, fmt.Errorf("%s: (should be unreachable code in github.com/MicahParks/jsontype) unsupported type: %T", errUnmarshalPackage, j.v)
	}
	return json.Marshal(s)
}

// UnmarshalJSON helps implement the json.Unmarshaler interface.
func (j *JSONType[T]) UnmarshalJSON(bytes []byte) error {
	var v any
	switch any(j.v).(type) {
	case *mail.Address:
		var s string
		err := json.Unmarshal(bytes, &s)
		if err != nil {
			return fmt.Errorf("%s: %s: %w", errUnmarshalPackage, errStringUnmarshal, err)
		}
		addr, err := mail.ParseAddress(s)
		if err != nil {
			return fmt.Errorf("%s: failed to parse email address: %w", errUnmarshalPackage, err)
		}
		v = any(addr)
	case *regexp.Regexp:
		var s string
		err := json.Unmarshal(bytes, &s)
		if err != nil {
			return fmt.Errorf("%s: %s: %w", errStringUnmarshal, errUnmarshalPackage, err)
		}
		re, err := regexp.Compile(s)
		if err != nil {
			return fmt.Errorf("%s: failed to compile regexp: %w", errUnmarshalPackage, err)
		}
		v = any(re)
	case time.Duration:
		var s string
		err := json.Unmarshal(bytes, &s)
		if err != nil {
			return fmt.Errorf("%s: %s: %w", errUnmarshalPackage, errStringUnmarshal, err)
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("%s: failed to parse duration: %w", errUnmarshalPackage, err)
		}
		v = any(d)
	case time.Time:
		var s string
		err := json.Unmarshal(bytes, &s)
		if err != nil {
			return fmt.Errorf("%s: %s: %w", errUnmarshalPackage, errStringUnmarshal, err)
		}
		format := time.RFC3339
		if j.options.TimeFormatUnmarshal != "" {
			format = j.options.TimeFormatUnmarshal
		}
		t, err := time.Parse(format, s)
		if err != nil {
			return fmt.Errorf("%s: failed to parse time: %w", errUnmarshalPackage, err)
		}
		v = any(t)
	case *url.URL:
		var s string
		err := json.Unmarshal(bytes, &s)
		if err != nil {
			return fmt.Errorf("%s: %s: %w", errUnmarshalPackage, errStringUnmarshal, err)
		}
		u, err := url.Parse(s)
		if err != nil {
			return fmt.Errorf("%s: failed to parse url: %w", errUnmarshalPackage, err)
		}
		v = any(u)
	default:
		return fmt.Errorf(errUnreachableFmt, errUnmarshalPackage, j.v)
	}
	j.v = v.(T)
	return nil
}
