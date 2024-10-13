package otp

import (
	"errors"
	"testing"

	mld "github.com/MicahParks/magiclinksdev"
)

func TestGenerate(t *testing.T) {
	tc := []struct {
		name        string
		params      CreateParams
		expectedErr error
	}{
		{
			name:        "Empty",
			params:      CreateParams{},
			expectedErr: mld.ErrParams,
		},
		{
			name: "LengthZero",
			params: CreateParams{
				CharSetNumeric: true,
			},
			expectedErr: nil,
		},
		{
			name: "AlphaLower",
			params: CreateParams{
				CharSetAlphaLower: true,
				CharSetAlphaUpper: false,
				CharSetNumeric:    false,
				Length:            mld.DefaultOTPLength,
			},
		},
		{
			name: "AlphaUpper",
			params: CreateParams{
				CharSetAlphaLower: false,
				CharSetAlphaUpper: true,
				CharSetNumeric:    false,
				Length:            mld.DefaultOTPLength,
			},
		},
		{
			name: "Numeric",
			params: CreateParams{
				CharSetAlphaLower: false,
				CharSetAlphaUpper: false,
				CharSetNumeric:    true,
				Length:            mld.DefaultOTPLength,
			},
		},
		{
			name: "AlphaNumeric",
			params: CreateParams{
				CharSetAlphaLower: true,
				CharSetAlphaUpper: true,
				CharSetNumeric:    true,
				Length:            mld.DefaultOTPLength,
			},
		},
		{
			name: "ShortLength",
			params: CreateParams{
				CharSetAlphaLower: true,
				CharSetAlphaUpper: true,
				CharSetNumeric:    true,
				Length:            1,
			},
		},
		{
			name: "LongLength",
			params: CreateParams{
				CharSetAlphaLower: true,
				CharSetAlphaUpper: true,
				CharSetNumeric:    true,
				Length:            12,
			},
		},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			o, err := Generate(tt.params)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected %v, got %v", tt.expectedErr, err)
			}
			expectedSet := make([]rune, 0)
			if tt.params.CharSetAlphaLower {
				expectedSet = append(expectedSet, alphaLower...)
			}
			if tt.params.CharSetAlphaUpper {
				expectedSet = append(expectedSet, alphaUpper...)
			}
			if tt.params.CharSetNumeric {
				expectedSet = append(expectedSet, numeric...)
			}
			if uint(len(o)) != tt.params.Length {
				t.Fatalf("expected length %d, got %d", tt.params.Length, len(o))
			}
			for _, r := range o {
				found := false
				for _, e := range expectedSet {
					if r == e {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("rune %c not found in expected set", r)
				}
			}
		})
	}
}
