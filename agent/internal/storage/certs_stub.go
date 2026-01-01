//go:build !linux

package storage

import (
    "errors"
    "os"
    "path/filepath"
)

type CertStore struct {
    CertPath string
    KeyPath  string
}

func (s CertStore) EnsurePaths() error {
    dir := filepath.Dir(s.CertPath)
    if err := os.MkdirAll(dir, 0700); err != nil {
        return err
    }
    kdir := filepath.Dir(s.KeyPath)
    if err := os.MkdirAll(kdir, 0700); err != nil {
        return err
    }
    return nil
}

func (s CertStore) Save(certPEM, keyPEM []byte) error {
    if err := s.EnsurePaths(); err != nil {
        return err
    }
    if err := os.WriteFile(s.CertPath, certPEM, 0600); err != nil {
        return err
    }
    if err := os.WriteFile(s.KeyPath, keyPEM, 0600); err != nil {
        _ = os.Remove(s.CertPath)
        return err
    }
    return nil
}

func (s CertStore) Load() ([]byte, []byte, error) {
    cert, cerr := os.ReadFile(s.CertPath)
    key, kerr := os.ReadFile(s.KeyPath)
    if cerr != nil || kerr != nil {
        return nil, nil, errors.New("missing certificate files")
    }
    return cert, key, nil
}

