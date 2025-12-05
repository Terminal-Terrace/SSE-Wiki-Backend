package pkg

import (
	"strconv"

	"terminal-terrace/auth-service/internal/database"
	"terminal-terrace/auth-service/internal/model/user"
	"terminal-terrace/response"

	"gorm.io/gorm"
)

// GetUserByID retrieves a user by their ID
func GetUserByID(userIDStr string) (*user.User, *response.BusinessError) {
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.InvalidParameter),
			response.WithErrorMessage("无效的用户ID格式"),
		)
	}

	var foundUser user.User
	result := database.PostgresDB.First(&foundUser, uint(userID))
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, response.NewBusinessError(
				response.WithErrorCode(response.NotFound),
				response.WithErrorMessage("用户不存在"),
			)
		}
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("查询用户失败"),
		)
	}

	return &foundUser, nil
}
