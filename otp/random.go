package otp

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var (
	alphaLower = []rune("abcdefghijklmnopqrstuvwxyz")
	alphaUpper = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	numeric    = []rune("0123456789")
)

var (
	ErrInvalidArgs = errors.New("invalid OTP arguments")
)

type OTPArgs struct { // TODO Rename?
	CharSetAlphaLower bool
	CharSetAlphaUpper bool
	CharSetNumeric    bool
	Length            int64
}

func generateOTP(args OTPArgs) (string, error) {
	charSet := make([]rune, 0)
	if args.CharSetAlphaLower {
		charSet = append(charSet, alphaLower...)
	}
	if args.CharSetAlphaUpper {
		charSet = append(charSet, alphaUpper...)
	}
	if args.CharSetNumeric {
		charSet = append(charSet, numeric...)
	}
	if len(charSet) == 0 {
		return "", fmt.Errorf("must include at least one character set: %w", ErrInvalidArgs)
	}
	o := strings.Builder{}
	for range args.Length {
		i, err := rand.Int(rand.Reader, big.NewInt(int64(len(charSet))))
		if err != nil {
			return "", fmt.Errorf("failed to read random number for OTP: %w", err)
		}
		_, err = o.WriteRune(charSet[i.Int64()])
		if err != nil {
			return "", fmt.Errorf("failed to write rune to OTP string builder: %w", err)
		}
	}
	return o.String(), nil
}
