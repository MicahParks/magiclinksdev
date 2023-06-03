package postgres

import (
	"errors"
	"testing"
)

func TestCompareSemVer(t *testing.T) {
	cases := []struct {
		programSemVer  string
		databaseSemVer string
		name           string
		setupErr       bool
	}{
		{
			name:     "Empty",
			setupErr: true,
		},
		{
			programSemVer: "v1.0.0",
			name:          "EmptyConfig",
			setupErr:      true,
		},
		{
			databaseSemVer: "v1.0.0",
			name:           "EmptyDatabase",
			setupErr:       true,
		},
		{
			programSemVer:  "v0.1.0",
			databaseSemVer: "v0.1.1",
			name:           "InvalidDev",
			setupErr:       true,
		},
		{
			programSemVer:  "v0.1.0",
			databaseSemVer: "v0.1.0",
			name:           "ValidDev",
		},
		{
			programSemVer:  "v1.0.0",
			databaseSemVer: "v1.0.0",
			name:           "Valid",
		},
		{
			programSemVer:  "v1.0.0",
			databaseSemVer: "v1.0.1",
			name:           "ValidPatch",
		},
		{
			programSemVer:  "v1.0.0",
			databaseSemVer: "v1.1.0",
			name:           "ValidMinor",
		},
		{
			programSemVer:  "v1.1.0",
			databaseSemVer: "v1.0.0",
			name:           "InvalidMinor",
			setupErr:       true,
		},
		{
			programSemVer:  "v1.0.0",
			databaseSemVer: "v2.0.0",
			name:           "InvalidDatabaseMajor",
			setupErr:       true,
		},
		{
			programSemVer:  "v2.0.0",
			databaseSemVer: "v1.0.0",
			name:           "InvalidConfigMajor",
			setupErr:       true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := compareSemVer(tc.programSemVer, tc.databaseSemVer)
			if err != nil {
				if tc.setupErr && errors.Is(err, ErrPostgresSetupCheck) {
					return
				}
				t.Fatalf("Unexpected error: %v.", err)
			}
		})
	}
}
