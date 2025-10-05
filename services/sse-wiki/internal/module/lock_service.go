package module

import (
	"context"
	"strconv"
	pkgDatabase "terminal-terrace/database"
	"terminal-terrace/response"
	"time"
)

const (
	LockKey        = "navigation:edit_lock"
	LockExpiration = 15 * time.Minute
)

type LockService struct {
	redis *pkgDatabase.RedisClient
}

func NewLockService(redis *pkgDatabase.RedisClient) *LockService {
	return &LockService{redis: redis}
}

// AcquireLock 获取编辑锁
func (s *LockService) AcquireLock(userID uint, username string) (*LockResponse, error) {
	ctx := context.Background()

	// 检查锁状态
	lockData, err := s.redis.HGetAll(ctx, LockKey).Result()
	if err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("检查锁状态失败"),
		)
	}

	// 如果锁存在且不是当前用户持有
	if len(lockData) > 0 {
		lockedUserIDStr := lockData["user_id"]
		if lockedUserIDStr != strconv.Itoa(int(userID)) {
			// 锁被其他用户占用
			return &LockResponse{
				Success: false,
				LockedBy: &UserInfo{
					ID:       parseUserID(lockedUserIDStr),
					Username: lockData["username"],
				},
				LockedAt: lockData["locked_at"],
			}, nil
		}
	}

	// 设置锁
	lockInfo := map[string]interface{}{
		"user_id":   userID,
		"username":  username,
		"locked_at": time.Now().Format(time.RFC3339),
	}

	if err := s.redis.HSet(ctx, LockKey, lockInfo).Err(); err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取锁失败"),
		)
	}

	// 设置过期时间
	if err := s.redis.Expire(ctx, LockKey, LockExpiration).Err(); err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("设置锁过期时间失败"),
		)
	}

	return &LockResponse{
		Success: true,
	}, nil
}

// ReleaseLock 释放编辑锁
func (s *LockService) ReleaseLock(userID uint) error {
	ctx := context.Background()

	// 检查是否是当前用户持有的锁
	lockData, err := s.redis.HGetAll(ctx, LockKey).Result()
	if err != nil {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("检查锁状态失败"),
		)
	}

	if len(lockData) > 0 {
		lockedUserIDStr := lockData["user_id"]
		if lockedUserIDStr != strconv.Itoa(int(userID)) {
			return response.NewBusinessError(
				response.WithErrorCode(response.Forbidden),
				response.WithErrorMessage("不能释放他人持有的锁"),
			)
		}
	}

	// 删除锁
	if err := s.redis.Del(ctx, LockKey).Err(); err != nil {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("释放锁失败"),
		)
	}

	return nil
}

// parseUserID 解析用户ID
func parseUserID(idStr string) uint {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0
	}
	return uint(id)
}
