package token

import "testing"

func TestConfigIsIdentifier(t *testing.T) {
	if got := SearchKeyword([]byte("config")); got != IDENTIFIER {
		t.Fatalf("expected config to be an identifier, got %s", got)
	}
}
