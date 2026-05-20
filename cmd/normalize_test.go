package cmd

import "testing"

func TestNormalizeBackendRegion(t *testing.T) {
	tests := map[string]string{
		"":                          "dev",
		"dev":                       "dev",
		"na1":                       "https://na1.staticbackend.dev",
		"na1\n":                     "https://na1.staticbackend.dev",
		"na1.staticbackend.dev":     "https://na1.staticbackend.dev",
		"https://example.com":       "https://example.com",
		"http://localhost:8099":     "http://localhost:8099",
		`"https://example.com"`:     "https://example.com",
		"'https://example.com'\r\n": "https://example.com",
	}

	for input, want := range tests {
		if got := normalizeBackendRegion(input); got != want {
			t.Fatalf("normalizeBackendRegion(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCleanConfigValue(t *testing.T) {
	if got := cleanConfigValue(" 'token-value'\r\n"); got != "token-value" {
		t.Fatalf("cleanConfigValue returned %q", got)
	}
}
