package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
	"pulse/backend/pkg/external/logging"
)

type CAService struct {
	db           *gorm.DB
	rootCert     *x509.Certificate
	rootKey      *rsa.PrivateKey
	certValidity time.Duration
}

func NewCAService(db *gorm.DB) *CAService {
	// Load or generate root CA
	rootKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		logging.Logger.Fatal("Failed to generate root key", zap.Error(err))
	}

	var rootCert *x509.Certificate
	rootCert = &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Pulse CA"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	rootCertBytes, err := x509.CreateCertificate(rand.Reader, rootCert, rootCert, &rootKey.PublicKey, rootKey)
	if err != nil {
		logging.Logger.Fatal("Failed to create root cert", zap.Error(err))
	}

	parsedCert, err := x509.ParseCertificate(rootCertBytes)
	if err != nil {
		logging.Logger.Fatal("Failed to parse root cert", zap.Error(err))
	}

	return &CAService{
		db:           db,
		rootCert:     parsedCert,
		rootKey:      rootKey,
		certValidity: 365 * 24 * time.Hour,
	}
}

func (s *CAService) IssueCertificate(identity string) (string, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: identity},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(s.certValidity),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, s.rootCert, &priv.PublicKey, s.rootKey)
	if err != nil {
		return "", "", err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	// Store in DB
	certModel := models.Certificate{
		Identity:  identity,
		Status:    models.CertStatusActive,
		IssuedAt:  time.Now(),
		ExpiresAt: template.NotAfter,
	}
	if err := s.db.Create(&certModel).Error; err != nil {
		return "", "", err
	}

	metadata := map[string]interface{}{"identity": identity, "cert_id": certModel.ID}
	logging.Audit("Certificate issued", metadata)

	return string(certPEM), string(keyPEM), nil
}

func (s *CAService) RotateCertificate(identity string) error {
	logging.Logger.Info("Rotating certificate", zap.String("identity", identity))
	// TODO: Implement rotation logic
	return nil
}

func (s *CAService) RevokeCertificate(identity string) error {
	var cert models.Certificate
	if err := s.db.Where("identity = ? AND status = ?", identity, models.CertStatusActive).First(&cert).Error; err != nil {
		return err
	}

	cert.Status = models.CertStatusRevoked
	if err := s.db.Save(&cert).Error; err != nil {
		return err
	}

	metadata := map[string]interface{}{"identity": identity, "cert_id": cert.ID}
	logging.Audit("Certificate revoked", metadata)

	return nil
}
