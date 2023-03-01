package postgres

import (
	"errors"
	"testing"
)

func TestCompareSemVer(t *testing.T) {
	cases := []struct {
		fromConfig string
		inDatabase string
		name       string
		setupErr   bool
	}{
		{
			name:     "Empty",
			setupErr: true,
		},
		{
			fromConfig: "v1.0.0",
			name:       "EmptyConfig",
			setupErr:   true,
		},
		{
			inDatabase: "v1.0.0",
			name:       "EmptyDatabase",
			setupErr:   true,
		},
		{
			fromConfig: "v0.1.0",
			inDatabase: "v0.1.1",
			name:       "InvalidDev",
			setupErr:   true,
		},
		{
			fromConfig: "v0.1.0",
			inDatabase: "v0.1.0",
			name:       "ValidDev",
		},
		{
			fromConfig: "v1.0.0",
			inDatabase: "v1.0.0",
			name:       "Valid",
		},
		{
			fromConfig: "v1.0.0",
			inDatabase: "v1.0.1",
			name:       "ValidPatch",
		},
		{
			fromConfig: "v1.0.0",
			inDatabase: "v1.1.0",
			name:       "ValidMinor",
		},
		{
			fromConfig: "v1.1.0",
			inDatabase: "v1.0.0",
			name:       "InvalidMinor",
			setupErr:   true,
		},
		{
			fromConfig: "v1.0.0",
			inDatabase: "v2.0.0",
			name:       "InvalidDatabaseMajor",
			setupErr:   true,
		},
		{
			fromConfig: "v2.0.0",
			inDatabase: "v1.0.0",
			name:       "InvalidConfigMajor",
			setupErr:   true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := compareSemVer(tc.fromConfig, tc.inDatabase)
			if err != nil {
				if tc.setupErr && errors.Is(err, ErrPostgresSetupCheck) {
					return
				}
				t.Fatalf("Unexpected error: %v.", err)
			}
		})
	}
}
