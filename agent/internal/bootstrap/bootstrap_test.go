package bootstrap

import "testing"

func TestMissingTokenFails(t *testing.T) {
    _, _, err := ExchangeToken("https://cp.example", "v1/agent/bootstrap", "", nil)
    if err == nil {
        t.Fatal("expected error for missing token")
    }
}

