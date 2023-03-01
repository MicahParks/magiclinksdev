[![Go Report Card](https://goreportcard.com/badge/github.com/MicahParks/jsontype)](https://goreportcard.com/report/github.com/MicahParks/jsontype) [![Go Reference](https://pkg.go.dev/badge/github.com/MicahParks/jsontype.svg)](https://pkg.go.dev/github.com/MicahParks/jsontype)
# jsontype
This package provides a generic [`json.Marshaler`](https://pkg.go.dev/encoding/json#Marshaler)
and [`json.Unmarshaler`](https://pkg.go.dev/encoding/json#Unmarshaler) wrapper for common Go types. This helps minimize
custom JSON string parsing and formatting code.

```go
import "github.com/MicahParks/jsontype"
```

# Supported Types
1. [`*mail.Address`](https://pkg.go.dev/net/mail#Address)
2. [`*regexp.Regexp`](https://pkg.go.dev/regexp#Regexp)
3. [`time.Time`](https://pkg.go.dev/time#Time)
4. [`time.Duration`](https://pkg.go.dev/time#Duration)
5. [`*url.URL`](https://pkg.go.dev/net/url#URL)

# Planned support
1. [`*big.Float`](https://pkg.go.dev/math/big#Float)
2. [`*big.Int`](https://pkg.go.dev/math/big#Int)
3. [`*big.Rat`](https://pkg.go.dev/math/big#Rat)
4. [`netip.Addr`](https://pkg.go.dev/net/netip#Addr)
5. [`netip.AddrPort`](https://pkg.go.dev/net/netip#AddrPort)
6. [`netip.Prefix`](https://pkg.go.dev/net/netip#Prefix)

# Usage
* All methods are safe for concurrent use by multiple goroutines.

## Define a data structure using one or more *jsontype.JSONType generic field.
```go
type myConfig struct {
	Ends            *jsontype.JSONType[time.Time]      `json:"ends"`
	GetInterval     *jsontype.JSONType[time.Duration]  `json:"getInterval"`
	NotificationMsg string                             `json:"notificationMsg"`
	Notify          *jsontype.JSONType[*mail.Address]  `json:"notify"`
	TargetPage      *jsontype.JSONType[*url.URL]       `json:"targetPage"`
	TargetRegExp    *jsontype.JSONType[*regexp.Regexp] `json:"targetRegExp"`
}
```

## Optionally set non-default behavior through options before unmarshalling
```go
// Set non-default unmarshal behavior.
endOpts := jsontype.Options{
	TimeFormatUnmarshal: time.RFC1123,
}
config.Ends = jsontype.NewWithOptions(time.Time{}, endOpts)
```

## Unmarshal JSON into the data structure
```go
// Unmarshal the configuration.
err := json.Unmarshal(json.RawMessage(exampleConfig), &config)
if err != nil {
	logger.Fatalf("failed to unmarshal JSON: %s", err)
}
```

## Use the fields on the data structure by accessing the `.Get()` method
```go
// Access fields on the unmarshalled configuration.
logger.Printf("Ends: %s", config.Ends.Get().String())
logger.Printf("Get interval: %s", config.GetInterval.Get().String())
```

## Optionally set non-default behavior through options before marshalling
```go
// Set non-default marshal behavior.
emailOpts := jsontype.Options{
	MailAddressAddressOnlyMarshal: true,
	MailAddressLowerMarshal:       true,
}
config.Notify = jsontype.NewWithOptions(config.Notify.Get(), emailOpts)
```

## Marshal the data structure into JSON
```go
// Marshal the configuration back to JSON.
remarshaled, err := json.MarshalIndent(config, "", "  ")
if err != nil {
	logger.Fatalf("failed to re-marshal configuration: %s", err)
}
```

# Examples
Please see the `examples` directory.

```go
package main

import (
	"encoding/json"
	"log"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/MicahParks/jsontype"
)

const exampleConfig = `{
  "ends": "Wed, 04 Oct 2022 00:00:00 MST",
  "getInterval": "1h30m",
  "notificationMsg": "Your item is on sale!",
  "notify": "EXAMPLE@example.com",
  "targetPage": "https://www.example.com",
  "targetRegExp": "example"
}`

type myConfig struct {
	Ends            *jsontype.JSONType[time.Time]      `json:"ends"`
	GetInterval     *jsontype.JSONType[time.Duration]  `json:"getInterval"`
	NotificationMsg string                             `json:"notificationMsg"`
	Notify          *jsontype.JSONType[*mail.Address]  `json:"notify"`
	TargetPage      *jsontype.JSONType[*url.URL]       `json:"targetPage"`
	TargetRegExp    *jsontype.JSONType[*regexp.Regexp] `json:"targetRegExp"`
}

func main() {
	logger := log.New(os.Stdout, "", 0)
	var config myConfig

	// Set non-default unmarshal behavior.
	endOpts := jsontype.Options{
		TimeFormatUnmarshal: time.RFC1123,
	}
	config.Ends = jsontype.NewWithOptions(time.Time{}, endOpts)

	// Unmarshal the configuration.
	err := json.Unmarshal(json.RawMessage(exampleConfig), &config)
	if err != nil {
		logger.Fatalf("failed to unmarshal JSON: %s", err)
	}

	// Access fields on the unmarshalled configuration.
	logger.Printf("Ends: %s", config.Ends.Get().String())
	logger.Printf("Get interval: %s", config.GetInterval.Get().String())

	// Set non-default marshal behavior.
	emailOpts := jsontype.Options{
		MailAddressAddressOnlyMarshal: true,
		MailAddressLowerMarshal:       true,
	}
	config.Notify = jsontype.NewWithOptions(config.Notify.Get(), emailOpts)

	// Marshal the configuration back to JSON.
	remarshaled, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		logger.Fatalf("failed to re-marshal configuration: %s", err)
	}
	logger.Println(string(remarshaled))
}
```

Output:
```
Ends: 2022-10-04 00:00:00 -0500 -0500
Get interval: 1h30m0s
{
  "ends": "2022-10-04T00:00:00-05:00",
  "getInterval": "1h30m0s",
  "notificationMsg": "Your item is on sale!",
  "notify": "example@example.com",
  "targetPage": "https://www.example.com",
  "targetRegExp": "example"
}
```

# Testing
```
$ go test -cover -race
PASS
coverage: 90.1% of statements
ok      github.com/MicahParks/jsontype  0.021s
```
