package utils

import (
    "errors"
    "os"
    "runtime"
)

type OSAdapter interface {
    IsLinux() bool
}

type DefaultOSAdapter struct{}

func (DefaultOSAdapter) IsLinux() bool { return runtime.GOOS == "linux" }

func EnsureSecureDir(path string, osd OSAdapter) error {
    if path == "" {
        return errors.New("empty path")
    }
    if _, err := os.Stat(path); os.IsNotExist(err) {
        if err := os.MkdirAll(path, 0o700); err != nil {
            return err
        }
    }
    if osd.IsLinux() {
        if err := os.Chmod(path, 0o700); err != nil {
            return err
        }
        st, err := os.Stat(path)
        if err != nil {
            return err
        }
        if st.Mode().Perm()&0o077 != 0 {
            return errors.New("insecure directory permissions")
        }
    }
    return nil
}

func EnsureSecureFile(path string, osd OSAdapter) error {
    if path == "" {
        return errors.New("empty path")
    }
    if _, err := os.Stat(path); err == nil {
        if osd.IsLinux() {
            if err := os.Chmod(path, 0o600); err != nil {
                return err
            }
            st, err := os.Stat(path)
            if err != nil {
                return err
            }
            if st.Mode().Perm()&0o077 != 0 {
                return errors.New("insecure file permissions")
            }
        }
    }
    return nil
}

