package storage

type CertStorer interface {
	Save(certPEM, keyPEM []byte) error
	Load() ([]byte, []byte, error)
}
