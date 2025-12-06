package grpc

import (
	"context"

	"terminal-terrace/sse-wiki/internal/database"
	"terminal-terrace/sse-wiki/internal/module"
	pb "terminal-terrace/sse-wiki/protobuf/proto/module_service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ModuleServiceImpl implements the ModuleService gRPC interface
type ModuleServiceImpl struct {
	pb.UnimplementedModuleServiceServer
	moduleService *module.ModuleService
	lockService   *module.LockService
}

// NewModuleServiceImpl creates a new ModuleService implementation
func NewModuleServiceImpl() *ModuleServiceImpl {
	return &ModuleServiceImpl{
		moduleService: module.NewModuleService(database.PostgresDB),
		lockService:   module.NewLockService(database.RedisDB),
	}
}

// GetModuleTree returns the complete module tree
func (s *ModuleServiceImpl) GetModuleTree(ctx context.Context, req *pb.GetModuleTreeRequest) (*pb.GetModuleTreeResponse, error) {
	tree, err := s.moduleService.GetModuleTree(uint(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert to proto format
	pbTree := make([]*pb.ModuleTreeNode, len(tree))
	for i, node := range tree {
		pbTree[i] = convertModuleTreeNode(node)
	}

	return &pb.GetModuleTreeResponse{Tree: pbTree}, nil
}

// GetModule returns a single module by ID
func (s *ModuleServiceImpl) GetModule(ctx context.Context, req *pb.GetModuleRequest) (*pb.GetModuleResponse, error) {
	mod, err := s.moduleService.GetModule(uint(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.GetModuleResponse{
		Module: convertModule(mod),
	}, nil
}

// GetBreadcrumbs returns breadcrumb navigation for a module
func (s *ModuleServiceImpl) GetBreadcrumbs(ctx context.Context, req *pb.GetBreadcrumbsRequest) (*pb.GetBreadcrumbsResponse, error) {
	breadcrumbs, err := s.moduleService.GetBreadcrumbs(uint(req.ModuleId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbBreadcrumbs := make([]*pb.BreadcrumbNode, len(breadcrumbs))
	for i, bc := range breadcrumbs {
		pbBreadcrumbs[i] = &pb.BreadcrumbNode{
			Id:   uint32(bc.ID),
			Name: bc.Name,
		}
	}

	return &pb.GetBreadcrumbsResponse{Breadcrumbs: pbBreadcrumbs}, nil
}

// CreateModule creates a new module
func (s *ModuleServiceImpl) CreateModule(ctx context.Context, req *pb.CreateModuleRequest) (*pb.CreateModuleResponse, error) {
	createReq := module.CreateModuleRequest{
		Name:        req.Name,
		Description: req.Description,
	}
	if req.ParentId > 0 {
		parentID := uint(req.ParentId)
		createReq.ParentID = &parentID
	}

	mod, err := s.moduleService.CreateModule(createReq, uint(req.UserId), req.UserRole)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateModuleResponse{
		Module: convertModule(mod),
	}, nil
}

// UpdateModule updates an existing module
func (s *ModuleServiceImpl) UpdateModule(ctx context.Context, req *pb.UpdateModuleRequest) (*pb.UpdateModuleResponse, error) {
	updateReq := module.UpdateModuleRequest{
		Name:        req.Name,
		Description: req.Description,
	}
	if req.ParentId > 0 {
		parentID := uint(req.ParentId)
		updateReq.ParentID = &parentID
	}

	err := s.moduleService.UpdateModule(uint(req.Id), updateReq, uint(req.UserId), req.UserRole)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateModuleResponse{}, nil
}

// DeleteModule deletes a module and its children
func (s *ModuleServiceImpl) DeleteModule(ctx context.Context, req *pb.DeleteModuleRequest) (*pb.DeleteModuleResponse, error) {
	count, err := s.moduleService.DeleteModule(uint(req.Id), uint(req.UserId), req.UserRole)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteModuleResponse{DeletedModules: int32(count)}, nil
}

// GetModerators returns the list of moderators for a module
func (s *ModuleServiceImpl) GetModerators(ctx context.Context, req *pb.GetModeratorsRequest) (*pb.GetModeratorsResponse, error) {
	moderators, err := s.moduleService.GetModerators(uint(req.ModuleId), uint(req.UserId), req.UserRole)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbModerators := make([]*pb.ModeratorInfo, len(moderators))
	for i, m := range moderators {
		pbModerators[i] = &pb.ModeratorInfo{
			UserId:    uint32(m.UserID),
			Username:  m.Username,
			Role:      m.Role,
			CreatedAt: m.CreatedAt,
		}
	}

	return &pb.GetModeratorsResponse{Moderators: pbModerators}, nil
}

// AddModerator adds a moderator to a module
func (s *ModuleServiceImpl) AddModerator(ctx context.Context, req *pb.AddModeratorRequest) (*pb.AddModeratorResponse, error) {
	addReq := module.AddModeratorRequest{
		UserID: uint(req.TargetUserId),
		Role:   req.Role,
	}

	err := s.moduleService.AddModerator(uint(req.ModuleId), addReq, uint(req.UserId), req.UserRole)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddModeratorResponse{}, nil
}

// RemoveModerator removes a moderator from a module
func (s *ModuleServiceImpl) RemoveModerator(ctx context.Context, req *pb.RemoveModeratorRequest) (*pb.RemoveModeratorResponse, error) {
	err := s.moduleService.RemoveModerator(uint(req.ModuleId), uint(req.TargetUserId), uint(req.UserId), req.UserRole)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RemoveModeratorResponse{}, nil
}

// HandleLock handles edit lock acquire/release
func (s *ModuleServiceImpl) HandleLock(ctx context.Context, req *pb.HandleLockRequest) (*pb.HandleLockResponse, error) {
	lockInfo := &pb.LockInfo{}

	if req.Action == "acquire" {
		// For acquire, we need username - for now use a placeholder
		// In production, this should come from the user service
		result, err := s.lockService.AcquireLock(uint(req.UserId), "User")
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		lockInfo.Success = result.Success
		if result.LockedBy != nil {
			lockInfo.LockedBy = &pb.UserInfo{
				Id:       uint32(result.LockedBy.ID),
				Username: result.LockedBy.Username,
			}
		}
		if result.LockedAt != "" {
			lockInfo.LockedAt = result.LockedAt
		}
	} else if req.Action == "release" {
		err := s.lockService.ReleaseLock(uint(req.UserId))
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		lockInfo.Success = true
	}

	return &pb.HandleLockResponse{LockInfo: lockInfo}, nil
}

// Helper functions to convert between internal types and proto types

func convertModuleTreeNode(node module.ModuleTreeNode) *pb.ModuleTreeNode {
	pbNode := &pb.ModuleTreeNode{
		Id:          uint32(node.ID),
		Name:        node.Name,
		Description: node.Description,
		OwnerId:     uint32(node.OwnerID),
		IsModerator: node.IsModerator,
	}

	if len(node.Children) > 0 {
		pbNode.Children = make([]*pb.ModuleTreeNode, len(node.Children))
		for i, child := range node.Children {
			pbNode.Children[i] = convertModuleTreeNode(child)
		}
	}

	return pbNode
}

func convertModule(mod interface{}) *pb.Module {
	// Handle different module types that might be returned
	switch m := mod.(type) {
	case *module.ModuleTreeNode:
		return &pb.Module{
			Id:          uint32(m.ID),
			Name:        m.Name,
			Description: m.Description,
			OwnerId:     uint32(m.OwnerID),
		}
	default:
		// For other types, we need to use reflection or type assertion
		// This is a simplified version
		return &pb.Module{}
	}
}
