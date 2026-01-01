package utils

import (
    "os"
    "testing"
)

type fakeLinux struct{}

func (fakeLinux) IsLinux() bool { return true }

func TestInvalidCertPermissionRefuses(t *testing.T) {
    f, err := os.CreateTemp("", "cert")
    if err != nil {
        t.Fatalf("%v", err)
    }
    defer os.Remove(f.Name())
    _ = os.Chmod(f.Name(), 0o777)
    err = EnsureSecureFile(f.Name(), fakeLinux{})
    if err == nil {
        t.Fatalf("expected error")
    }
}

