package register

import (
	"testing"

	"terminal-terrace/response"

	"github.com/stretchr/testify/assert"
)

func TestRegisterService_validateRequest(t *testing.T) {
	service := &RegisterService{}

	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效的注册请求",
			req: RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "Test123456",
				ConfirmPassword: "Test123456",
			},
			wantErr: false,
		},
		{
			name: "用户名为空",
			req: RegisterRequest{
				Username:        "",
				Email:           "test@example.com",
				Password:        "Test123456",
				ConfirmPassword: "Test123456",
			},
			wantErr: true,
			errMsg:  "用户名不能为空",
		},
		{
			name: "用户名太短",
			req: RegisterRequest{
				Username:        "ab",
				Email:           "test@example.com",
				Password:        "Test123456",
				ConfirmPassword: "Test123456",
			},
			wantErr: true,
			errMsg:  "用户名长度必须在3-50个字符之间",
		},
		{
			name: "用户名太长",
			req: RegisterRequest{
				Username:        "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
				Email:           "test@example.com",
				Password:        "Test123456",
				ConfirmPassword: "Test123456",
			},
			wantErr: true,
			errMsg:  "用户名长度必须在3-50个字符之间",
		},
		{
			name: "用户名包含非法字符",
			req: RegisterRequest{
				Username:        "test@user",
				Email:           "test@example.com",
				Password:        "Test123456",
				ConfirmPassword: "Test123456",
			},
			wantErr: true,
			errMsg:  "用户名只能包含字母、数字和下划线",
		},
		{
			name: "邮箱为空",
			req: RegisterRequest{
				Username:        "testuser",
				Email:           "",
				Password:        "Test123456",
				ConfirmPassword: "Test123456",
			},
			wantErr: true,
			errMsg:  "邮箱不能为空",
		},
		{
			name: "邮箱格式不正确",
			req: RegisterRequest{
				Username:        "testuser",
				Email:           "invalid-email",
				Password:        "Test123456",
				ConfirmPassword: "Test123456",
			},
			wantErr: true,
			errMsg:  "邮箱格式不正确",
		},
		{
			name: "密码为空",
			req: RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "",
				ConfirmPassword: "",
			},
			wantErr: true,
			errMsg:  "密码不能为空",
		},
		{
			name: "密码太短",
			req: RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "Tt1",
				ConfirmPassword: "Tt1",
			},
			wantErr: true,
			errMsg:  "密码长度必须在6-100个字符之间",
		},
		{
			name: "两次密码不一致",
			req: RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "Test123456",
				ConfirmPassword: "Test654321",
			},
			wantErr: true,
			errMsg:  "两次密码输入不一致",
		},
		{
			name: "密码强度不足-没有大写字母",
			req: RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "test123456",
				ConfirmPassword: "test123456",
			},
			wantErr: true,
			errMsg:  "密码强度不足，需包含大小写字母、数字",
		},
		{
			name: "密码强度不足-没有小写字母",
			req: RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "TEST123456",
				ConfirmPassword: "TEST123456",
			},
			wantErr: true,
			errMsg:  "密码强度不足，需包含大小写字母、数字",
		},
		{
			name: "密码强度不足-没有数字",
			req: RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "TestPassword",
				ConfirmPassword: "TestPassword",
			},
			wantErr: true,
			errMsg:  "密码强度不足，需包含大小写字母、数字",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateRequest(tt.req)

			if tt.wantErr {
				assert.NotNil(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Msg)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestRegisterService_isStrongPassword(t *testing.T) {
	service := &RegisterService{}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "强密码-包含大小写字母和数字",
			password: "Test123456",
			want:     true,
		},
		{
			name:     "强密码-包含特殊字符",
			password: "Test@123456",
			want:     true,
		},
		{
			name:     "弱密码-只有小写字母和数字",
			password: "test123456",
			want:     false,
		},
		{
			name:     "弱密码-只有大写字母和数字",
			password: "TEST123456",
			want:     false,
		},
		{
			name:     "弱密码-只有大小写字母",
			password: "TestPassword",
			want:     false,
		},
		{
			name:     "弱密码-只有小写字母",
			password: "testpassword",
			want:     false,
		},
		{
			name:     "弱密码-只有数字",
			password: "123456789",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.isStrongPassword(tt.password)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegisterRequest_Validation(t *testing.T) {
	// 测试结构体的边界情况
	tests := []struct {
		name        string
		username    string
		email       string
		password    string
		wantErrCode response.ResponseCode
	}{
		{
			name:        "用户名包含下划线是合法的",
			username:    "test_user_123",
			email:       "test@example.com",
			password:    "Test123456",
			wantErrCode: 0,
		},
		{
			name:        "用户名全是数字是合法的",
			username:    "123456",
			email:       "test@example.com",
			password:    "Test123456",
			wantErrCode: 0,
		},
		{
			name:        "邮箱带加号是合法的",
			username:    "testuser",
			email:       "test+123@example.com",
			password:    "Test123456",
			wantErrCode: 0,
		},
		{
			name:        "邮箱带下划线是合法的",
			username:    "testuser",
			email:       "test_user@example.com",
			password:    "Test123456",
			wantErrCode: 0,
		},
	}

	service := &RegisterService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := RegisterRequest{
				Username:        tt.username,
				Email:           tt.email,
				Password:        tt.password,
				ConfirmPassword: tt.password,
			}

			err := service.validateRequest(req)

			if tt.wantErrCode == 0 {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tt.wantErrCode, err.Code)
			}
		})
	}
}
