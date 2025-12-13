package grpc

import (
	"context"

	"terminal-terrace/sse-wiki/internal/database"
	moduleModel "terminal-terrace/sse-wiki/internal/model/module"
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
	// 从 JWT 获取用户信息
	user := GetUserFromContext(ctx)

	tree, err := s.moduleService.GetModuleTree(uint(user.UserID), user.Role)
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
		Module: &pb.Module{
			Id:          uint32(mod.ID),
			Description: mod.Description,
			ParentId:    uint32(*mod.ParentID),
			OwnerId:     uint32(mod.OwnerID),
			CreatedAt:   mod.CreatedAt.String(),
			UpdatedAt:   mod.UpdatedAt.String(),
		},
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
	// 从 JWT 获取用户信息
	user := GetUserFromContext(ctx)

	createReq := module.CreateModuleRequest{
		Name:        req.Name,
		Description: req.Description,
	}
	if req.ParentId > 0 {
		parentID := uint(req.ParentId)
		createReq.ParentID = &parentID
	}

	mod, err := s.moduleService.CreateModule(createReq, uint(user.UserID), user.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateModuleResponse{
		Module: convertModule(mod),
	}, nil
}

// UpdateModule updates an existing module
func (s *ModuleServiceImpl) UpdateModule(ctx context.Context, req *pb.UpdateModuleRequest) (*pb.UpdateModuleResponse, error) {
	// 从 JWT 获取用户信息
	user := GetUserFromContext(ctx)

	updateReq := module.UpdateModuleRequest{
		Name:        req.Name,
		Description: req.Description,
	}
	if req.ParentId > 0 {
		parentID := uint(req.ParentId)
		updateReq.ParentID = &parentID
	}

	err := s.moduleService.UpdateModule(uint(req.Id), updateReq, uint(user.UserID), user.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateModuleResponse{}, nil
}

// DeleteModule deletes a module and its children
func (s *ModuleServiceImpl) DeleteModule(ctx context.Context, req *pb.DeleteModuleRequest) (*pb.DeleteModuleResponse, error) {
	// 从 JWT 获取用户信息
	user := GetUserFromContext(ctx)

	count, err := s.moduleService.DeleteModule(uint(req.Id), uint(user.UserID), user.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteModuleResponse{DeletedModules: int32(count)}, nil
}

// GetModerators returns the list of moderators for a module
func (s *ModuleServiceImpl) GetModerators(ctx context.Context, req *pb.GetModeratorsRequest) (*pb.GetModeratorsResponse, error) {
	// 从 JWT 获取用户信息
	user := GetUserFromContext(ctx)

	moderators, err := s.moduleService.GetModerators(uint(req.ModuleId), uint(user.UserID), user.Role)
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
	// 从 JWT 获取用户信息
	user := GetUserFromContext(ctx)

	addReq := module.AddModeratorRequest{
		UserID: uint(req.TargetUserId),
		Role:   req.Role,
	}

	err := s.moduleService.AddModerator(uint(req.ModuleId), addReq, uint(user.UserID), user.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddModeratorResponse{}, nil
}

// RemoveModerator removes a moderator from a module
func (s *ModuleServiceImpl) RemoveModerator(ctx context.Context, req *pb.RemoveModeratorRequest) (*pb.RemoveModeratorResponse, error) {
	// 从 JWT 获取用户信息
	user := GetUserFromContext(ctx)

	err := s.moduleService.RemoveModerator(uint(req.ModuleId), uint(req.TargetUserId), uint(user.UserID), user.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RemoveModeratorResponse{}, nil
}

// HandleLock handles edit lock acquire/release
func (s *ModuleServiceImpl) HandleLock(ctx context.Context, req *pb.HandleLockRequest) (*pb.HandleLockResponse, error) {
	// 从 JWT 获取用户信息
	user := GetUserFromContext(ctx)

	lockInfo := &pb.LockInfo{}

	if req.Action == "acquire" {
		result, err := s.lockService.AcquireLock(uint(user.UserID), user.Username)
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
		err := s.lockService.ReleaseLock(uint(user.UserID))
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
	case *moduleModel.Module:
		parentID := uint32(0)
		if m.ParentID != nil {
			parentID = uint32(*m.ParentID)
		}
		return &pb.Module{
			Id:          uint32(m.ID),
			Name:        m.ModuleName,
			Description: m.Description,
			ParentId:    parentID,
			OwnerId:     uint32(m.OwnerID),
			CreatedAt:   m.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   m.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	default:
		// For other types, we need to use reflection or type assertion
		// This is a simplified version
		return &pb.Module{}
	}
}
