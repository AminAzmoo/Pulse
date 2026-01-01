package storage

import "testing"

func TestLoadMissingCerts(t *testing.T) {
    s := CertStore{CertPath: "testdata/missing.crt", KeyPath: "testdata/missing.key"}
    _, _, err := s.Load()
    if err == nil {
        t.Fatal("expected error for missing certs")
    }
}

