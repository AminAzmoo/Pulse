package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"pulse/backend/internal/ca"
	"pulse/backend/internal/services"
	"pulse/backend/pkg/external/grpc/pb"
	"pulse/backend/pkg/external/logging"
)

const (
	GRPCPort      = "50051"
	MinTLSVersion = tls.VersionTLS13
)

func NewGRPCServer(caSvc *ca.CAService, identitySvc *services.IdentityService, taskSvc *services.TaskService) (*grpc.Server, error) {
	certPath := os.Getenv("SERVER_CERT_PATH")
	keyPath := os.Getenv("SERVER_KEY_PATH")
	caCertPath := os.Getenv("CA_CERT_PATH")

	if certPath == "" || keyPath == "" || caCertPath == "" {
		return nil, fmt.Errorf("missing TLS config env vars")
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		MinVersion:   MinTLSVersion,
	}

	creds := credentials.NewTLS(tlsConfig)

	server := grpc.NewServer(grpc.Creds(creds))

	pb.RegisterAgentServiceServer(server, &AgentService{IdentitySvc: identitySvc, TaskSvc: taskSvc, CASvc: caSvc})

	reflection.Register(server)

	logging.Logger.Info("gRPC server initialized with mTLS")

	return server, nil
}

func StartGRPCServer(server *grpc.Server) error {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = GRPCPort
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	logging.Logger.Info("Starting gRPC server", zap.String("port", port))
	return server.Serve(lis)
}
