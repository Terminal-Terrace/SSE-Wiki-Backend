package register

type RegisterRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Email           string `json:"email"`
	Code            string `json:"code"`
	State           string `json:"state"`
}

type RegisterResponse struct {
	RefreshToken string `json:"refresh_token"`
	RedirectUrl  string `json:"redirect_url"`
}