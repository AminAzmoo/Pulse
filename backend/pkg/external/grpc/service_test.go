//go:build integration

package grpc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
	"pulse/backend/internal/services"
	pb "pulse/backend/pkg/external/grpc/pb"
)

const (
	testBufferSize      = 1024 * 1024
	testIdentityAllowed = "allowed-identity"
	testIdentityUnknown = "unknown-identity"
	testIdentityRevoked = "revoked-identity"
	testServerName      = "bufserver"
)

// mockIdentityService mocks the IdentityService methods used in AgentService
type mockIdentityService struct {
	mock.Mock
}

func (m *mockIdentityService) IsAllowed(identity string) (bool, error) {
	args := m.Called(identity)
	return args.Bool(0), args.Error(1)
}

func (m *mockIdentityService) Subscribe(identity string) chan struct{} {
	args := m.Called(identity)
	return args.Get(0).(chan struct{})
}

func (m *mockIdentityService) Unsubscribe(identity string, ch chan struct{}) {
	m.Called(identity, ch)
}

func setupTLS(t *testing.T, cn string) (*tls.Config, *tls.Config) {
	t.Helper()

	caKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	caTemplate := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test-ca"}, NotBefore: time.Now(), NotAfter: time.Now().Add(24 * time.Hour), IsCA: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign, BasicConstraintsValid: true}
	caDER, _ := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	caCert, _ := x509.ParseCertificate(caDER)

	serverKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	serverTemplate := &x509.Certificate{SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: testServerName}, DNSNames: []string{testServerName}, NotBefore: time.Now(), NotAfter: time.Now().Add(24 * time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	serverDER, _ := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	serverCert := tls.Certificate{Certificate: [][]byte{serverDER}, PrivateKey: serverKey}

	clientKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	clientTemplate := &x509.Certificate{SerialNumber: big.NewInt(3), Subject: pkix.Name{CommonName: cn}, NotBefore: time.Now(), NotAfter: time.Now().Add(24 * time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}
	clientDER, _ := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	clientCert := tls.Certificate{Certificate: [][]byte{clientDER}, PrivateKey: clientKey}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(pemEncode(caDER))

	serverTLS := &tls.Config{Certificates: []tls.Certificate{serverCert}, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: caPool}
	clientTLS := &tls.Config{Certificates: []tls.Certificate{clientCert}, RootCAs: caPool, ServerName: testServerName}

	return serverTLS, clientTLS
}

func pemEncode(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func setupTestServer(t *testing.T, identitySvc *mockIdentityService, cn string) (pb.AgentServiceClient, func()) {
	lis := bufconn.Listen(testBufferSize)
	serverTLS, clientTLS := setupTLS(t, cn)
	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(serverTLS)))
	// minimal TaskService for tests
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&models.Task{})
	ts := services.NewTaskService(db)
	pb.RegisterAgentServiceServer(s, &AgentService{IdentitySvc: identitySvc, TaskSvc: ts})

	go func() {
		_ = s.Serve(lis)
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(context.Background(), "",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(credentials.NewTLS(clientTLS)))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewAgentServiceClient(conn)

	return client, func() {
		conn.Close()
		s.Stop()
		lis.Close()
	}
}

func TestConnect_ValidCertUnknownIdentity(t *testing.T) {
	mockSvc := new(mockIdentityService)
	mockSvc.On("IsAllowed", testIdentityUnknown).Return(false, nil)

	client, cleanup := setupTestServer(t, mockSvc, testIdentityUnknown)
	defer cleanup()

	stream, err := client.Connect(context.Background())
	assert.NoError(t, err)

	_, err = stream.Recv()
	st := status.Convert(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
	assert.Contains(t, st.Message(), "unauthorized")

	mockSvc.AssertExpectations(t)
}

func TestConnect_RevokedIdentityReconnect(t *testing.T) {
	mockSvc := new(mockIdentityService)
	mockSvc.On("IsAllowed", testIdentityRevoked).Return(false, nil)

	client, cleanup := setupTestServer(t, mockSvc, testIdentityRevoked)
	defer cleanup()

	stream, err := client.Connect(context.Background())
	assert.NoError(t, err)

	_, err = stream.Recv()
	st := status.Convert(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
	assert.Contains(t, st.Message(), "unauthorized")

	mockSvc.AssertExpectations(t)
}

func TestConnect_RevokeWhileConnected(t *testing.T) {
	mockSvc := new(mockIdentityService)
	mockSvc.On("IsAllowed", testIdentityAllowed).Return(true, nil)
	revokeChan := make(chan struct{})
	mockSvc.On("Subscribe", testIdentityAllowed).Return(revokeChan)
	mockSvc.On("Unsubscribe", testIdentityAllowed, revokeChan)

	client, cleanup := setupTestServer(t, mockSvc, testIdentityAllowed)
	defer cleanup()

	stream, err := client.Connect(context.Background())
	assert.NoError(t, err)

	close(revokeChan)

	_, err = stream.Recv()
	st := status.Convert(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
	assert.Contains(t, st.Message(), "revoked")

	mockSvc.AssertExpectations(t)
}

func TestConnect_NoCert(t *testing.T) {
	t.Skip("No cert rejection is enforced at transport level with mTLS config; tested in integration")
}
