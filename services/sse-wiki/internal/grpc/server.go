package grpc

import (
	"fmt"
	"net"

	articlepb "terminal-terrace/sse-wiki/protobuf/proto/article_service"
	discussionpb "terminal-terrace/sse-wiki/protobuf/proto/discussion_service"
	modulepb "terminal-terrace/sse-wiki/protobuf/proto/module_service"
	reviewpb "terminal-terrace/sse-wiki/protobuf/proto/review_service"

	"google.golang.org/grpc"
)

// Server wraps the gRPC server
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer creates a new gRPC server with all wiki services registered
func NewServer(port int, moduleService modulepb.ModuleServiceServer, articleService articlepb.ArticleServiceServer, reviewService reviewpb.ReviewServiceServer, discussionService discussionpb.DiscussionServiceServer) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	grpcServer := grpc.NewServer()

	// Register all services
	modulepb.RegisterModuleServiceServer(grpcServer, moduleService)
	articlepb.RegisterArticleServiceServer(grpcServer, articleService)
	reviewpb.RegisterReviewServiceServer(grpcServer, reviewService)
	discussionpb.RegisterDiscussionServiceServer(grpcServer, discussionService)
	

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
