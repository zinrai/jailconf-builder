package main

import (
	"testing"
)

func TestExtractVersionFromURL(t *testing.T) {
	testCases := []struct {
		name          string
		url           string
		expectedVer   string
		expectedError bool
	}{
		{
			name:          "Valid URL",
			url:           "https://download.freebsd.org/ftp/releases/amd64/amd64/14.1-RELEASE/base.txz",
			expectedVer:   "14.1-RELEASE",
			expectedError: false,
		},
		{
			name:          "Valid URL with different version",
			url:           "http://download.freebsd.org/ftp/releases/amd64/amd64/13.2-RELEASE/base.txz",
			expectedVer:   "13.2-RELEASE",
			expectedError: false,
		},
		{
			name:          "Invalid URL - wrong format",
			url:           "https://download.freebsd.org/ftp/releases/amd64/amd64/14.1/base.txz",
			expectedVer:   "",
			expectedError: true,
		},
		{
			name:          "Invalid URL - not enough parts",
			url:           "https://download.freebsd.org/ftp/releases/base.txz",
			expectedVer:   "",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			version, err := extractVersionFromURL(tc.url)

			if tc.expectedError && err == nil {
				t.Errorf("Expected an error, but got none")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if version != tc.expectedVer {
				t.Errorf("Expected version %s, but got %s", tc.expectedVer, version)
			}
		})
	}
}
