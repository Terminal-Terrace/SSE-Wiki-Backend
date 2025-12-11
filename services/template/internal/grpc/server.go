package grpc

import (
	"fmt"
	"net"

	// TODO: 导入你的 protobuf 生成的包
	// pb "terminal-terrace/template/protobuf/proto/template_service"

	"google.golang.org/grpc"
)

// Server wraps the gRPC server
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer creates a new gRPC server with all services registered
// TODO: 添加你的 service 参数，例如：
// func NewServer(port int, exampleService pb.ExampleServiceServer) (*Server, error)
func NewServer(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	grpcServer := grpc.NewServer()

	// TODO: 注册你的 gRPC services
	// pb.RegisterExampleServiceServer(grpcServer, exampleService)

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
