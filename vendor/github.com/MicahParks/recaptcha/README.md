[![Go Report Card](https://goreportcard.com/badge/github.com/MicahParks/recaptcha)](https://goreportcard.com/report/github.com/MicahParks/recaptcha) [![Go Reference](https://pkg.go.dev/badge/github.com/MicahParks/recaptcha.svg)](https://pkg.go.dev/github.com/MicahParks/recaptcha)

# recaptcha

The purpose of this package is to provide a simple interface to the Google reCAPTCHA V3 service for verifying requests
server side in Golang.

# Basic usage
For complete examples, please see the `examples` directory.

```go
import "github.com/MicahParks/recaptcha"
```

## Step 1: Create the verifier
```go
// Create the verifier.
verifier := recaptcha.NewVerifierV3("mySecret", recaptcha.VerifierV3Options{})
```

## Step 2: Verify the request with Google
```go
// Verify the request with Google.
response, err := verifier.Verify(ctx, frontendToken, remoteAddr)
if err != nil {
    // Handle the error.
}
```

## Step 3: Check the response
```go
// Check the reCAPTCHA response.
err = response.Check(recaptcha.V3ResponseCheckOptions{
    Action:   []string{"submit"},
    Hostname: []string{"example.com"},
    Score:    0.5,
})
if err != nil {
    // Fail the request.
}
```

# Test coverage
Test coverage is currently `>90%`.

# References
* [reCAPTCHA V3 docs](https://developers.google.com/recaptcha/docs/v3)
* [reCAPTCHA V3 server side docs](https://developers.google.com/recaptcha/docs/verify)
