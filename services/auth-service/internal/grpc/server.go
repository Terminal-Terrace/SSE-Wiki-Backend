package grpc

import (
	"fmt"
	"net"

	pb "terminal-terrace/auth-service/internal/pb/authservice"

	"google.golang.org/grpc"
)

// Server wraps the gRPC server
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer creates a new gRPC server
func NewServer(port int, authService pb.AuthServiceServer) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authService)

	return &Server{
		grpcServer: grpcServer,
		listener:   listener,
	}, nil
}

// Start starts the gRPC server (blocking)
func (s *Server) Start() error {
	return s.grpcServer.Serve(s.listener)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.listener.Addr().String()
}
