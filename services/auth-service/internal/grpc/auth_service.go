package grpc

import (
	"context"

	"terminal-terrace/auth-service/internal/code"
	"terminal-terrace/auth-service/internal/database"
	"terminal-terrace/auth-service/internal/login"
	pb "terminal-terrace/auth-service/protobuf/proto/authservice"
	"terminal-terrace/auth-service/internal/pkg"
	"terminal-terrace/auth-service/internal/refresh"
	"terminal-terrace/auth-service/internal/register"
	"terminal-terrace/email"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServiceImpl implements the AuthService gRPC interface
// It acts as a thin adapter layer, delegating to existing service implementations
type AuthServiceImpl struct {
	pb.UnimplementedAuthServiceServer
	codeService     *code.CodeService
	refreshService  *refresh.RefreshTokenService
	registerService *register.RegisterService
}

// NewAuthServiceImpl creates a new AuthService implementation
func NewAuthServiceImpl(mailer *email.Client) *AuthServiceImpl {
	refreshTokenRepo := refresh.NewRefreshTokenRepository(database.RedisDB)
	return &AuthServiceImpl{
		codeService:     code.NewCodeService(mailer),
		refreshService:  refresh.NewRefreshTokenService(refreshTokenRepo),
		registerService: register.NewRegisterService(refreshTokenRepo),
	}
}

// Prelogin generates a state for CSRF protection
// Reuses existing pkg functions
func (s *AuthServiceImpl) Prelogin(ctx context.Context, req *pb.PreloginRequest) (*pb.PreloginResponse, error) {
	if req.RedirectUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "redirect_url is required")
	}

	// Generate state (reuse existing pkg function)
	state, err := pkg.GenerateState()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate state")
	}

	// Save state with redirect URL to Redis (reuse existing pkg function)
	if err := pkg.SaveStateWithRedirect(state, req.RedirectUrl); err != nil {
		return nil, status.Error(codes.Internal, "failed to save state")
	}

	return &pb.PreloginResponse{State: state}, nil
}

// SendCode sends a verification code to the email
// Delegates to existing code.CodeService
func (s *AuthServiceImpl) SendCode(ctx context.Context, req *pb.CodeRequest) (*pb.CodeResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	// Map proto CodeType to internal CodeType
	var codeType code.CodeType
	switch req.Type {
	case pb.CodeType_REGISTRATION:
		codeType = code.CodeTypeRegister
	case pb.CodeType_PASSWORD_RESET:
		codeType = code.CodeTypeResetPassword
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid code type")
	}

	// Delegate to existing code service
	bizErr := s.codeService.SendCode(code.SendCodeRequest{
		Email: req.Email,
		Type:  codeType,
	})
	if bizErr != nil {
		return nil, status.Error(codes.Internal, bizErr.Msg)
	}

	return &pb.CodeResponse{}, nil
}

// Login authenticates a user
// Delegates to existing login.LoginService implementations
func (s *AuthServiceImpl) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Map proto LoginType to internal type string
	var loginType string
	switch req.Type {
	case pb.LoginType_STANDARD:
		loginType = "sse-wiki"
	case pb.LoginType_GITHUB:
		loginType = "github"
	case pb.LoginType_SSE_MARKET:
		loginType = "sse-market"
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid login type")
	}

	// Delegate to existing login service
	// The login package uses a registry pattern with loginServices map
	result, bizErr := login.DoLogin(login.LoginRequest{
		Type:     loginType,
		State:    req.State,
		Username: req.Username,
		Password: req.Password,
		Code:     req.Code,
	})
	if bizErr != nil {
		return nil, status.Error(codes.Unauthenticated, bizErr.Msg)
	}

	return &pb.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		RedirectUrl:  result.RedirectUrl,
	}, nil
}

// Logout logs out a user
// Cookie clearing is done in Node.js Gateway, this just returns success
func (s *AuthServiceImpl) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// The actual cookie clearing is done in the Node.js Gateway
	// Here we could optionally invalidate the refresh token in Redis if needed
	return &pb.LogoutResponse{}, nil
}

// GetUserInfo returns user information based on user_id
// Note: In REST API, user info comes from JWT middleware context
// For gRPC, we need to look up by user_id
func (s *AuthServiceImpl) GetUserInfo(ctx context.Context, req *pb.InfoRequest) (*pb.InfoResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Get user from database using existing model
	user, bizErr := pkg.GetUserByID(req.UserId)
	if bizErr != nil {
		return nil, status.Error(codes.NotFound, bizErr.Msg)
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	return &pb.InfoResponse{
		User: &pb.AuthUser{
			UserId:   req.UserId,
			Username: username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

// RefreshToken refreshes the access token
// Delegates to existing refresh.RefreshTokenService
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	// Delegate to existing refresh service
	result, bizErr := s.refreshService.RefreshToken(refresh.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if bizErr != nil {
		return nil, status.Error(codes.Unauthenticated, bizErr.Msg)
	}

	return &pb.RefreshResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.NewRefreshToken,
	}, nil
}

// Register registers a new user
// Delegates to existing register.RegisterService
func (s *AuthServiceImpl) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" || req.Email == "" || req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "all fields are required")
	}

	// Delegate to existing register service
	// Note: Proto doesn't have confirm_password, we use password for both
	result, bizErr := s.registerService.Register(register.RegisterRequest{
		Username:        req.Username,
		Password:        req.Password,
		ConfirmPassword: req.Password,
		Email:           req.Email,
		Code:            req.Code,
		State:           "",
	})
	if bizErr != nil {
		return nil, status.Error(codes.Internal, bizErr.Msg)
	}

	return &pb.RegisterResponse{
		RefreshToken: result.RefreshToken,
		RedirectUrl:  result.RedirectUrl,
	}, nil
}
