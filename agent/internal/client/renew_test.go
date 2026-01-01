package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"

	"pulse/agent/internal/storage"
	pb "pulse/backend/pkg/external/grpc/pb"
)

// MockAgentServiceClient mocks the protobuf client
type MockAgentServiceClient struct {
	mock.Mock
}

func (m *MockAgentServiceClient) Connect(ctx context.Context, opts ...grpc.CallOption) (pb.AgentService_ConnectClient, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(pb.AgentService_ConnectClient), args.Error(1)
}

func (m *MockAgentServiceClient) RenewCertificate(ctx context.Context, in *pb.RenewCertificateRequest, opts ...grpc.CallOption) (*pb.RenewCertificateResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.RenewCertificateResponse), args.Error(1)
}

func TestMaybeRenew_NotExpired(t *testing.T) {
	// Setup fresh cert
	certPEM, keyPEM, _ := generateCert(t, time.Now().Add(60*24*time.Hour))
	tlsCert, _ := tls.X509KeyPair(certPEM, keyPEM)

	r := Runner{}
	client := new(MockAgentServiceClient)

	renewed := r.maybeRenew(context.Background(), client, tlsCert)
	assert.False(t, renewed)
	client.AssertNotCalled(t, "RenewCertificate")
}

func TestMaybeRenew_Expired_Success(t *testing.T) {
	// Setup expiring cert (expires in 29 days)
	certPEM, keyPEM, _ := generateCert(t, time.Now().Add(29*24*time.Hour))
	tlsCert, _ := tls.X509KeyPair(certPEM, keyPEM)

	// Temp storage
	tmpDir := t.TempDir()
	store := storage.CertStore{
		CertPath: tmpDir + "/cert.pem",
		KeyPath:  tmpDir + "/key.pem",
	}

	r := Runner{Store: store}
	client := new(MockAgentServiceClient)

	// Mock renewal response
	newCertPEM, newKeyPEM, _ := generateCert(t, time.Now().Add(365*24*time.Hour))
	resp := &pb.RenewCertificateResponse{
		Certificate: string(newCertPEM) + "\n" + string(newKeyPEM),
	}
	client.On("RenewCertificate", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	renewed := r.maybeRenew(context.Background(), client, tlsCert)
	assert.True(t, renewed)

	// Verify stored certs
	storedCert, storedKey, err := store.Load()
	assert.NoError(t, err)
	assert.Equal(t, newCertPEM, storedCert)
	assert.Equal(t, newKeyPEM, storedKey)
}

func generateCert(t *testing.T, expiry time.Time) ([]byte, []byte, *x509.Certificate) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     expiry,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	cert, _ := x509.ParseCertificate(der)
	return certPEM, keyPEM, cert
}
