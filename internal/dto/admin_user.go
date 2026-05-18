package dto

type ListUsersResponse struct {
	Items    []UserPayload `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=64"`
	Password    string `json:"password" binding:"required,min=6,max=128"`
	DisplayName string `json:"display_name" binding:"max=128"`
	Email       string `json:"email" binding:"max=128"`
	Role        string `json:"role" binding:"required,oneof=admin user"`
}

type UpdateUserRequest struct {
	DisplayName string `json:"display_name" binding:"max=128"`
	Email       string `json:"email" binding:"max=128"`
	Role        string `json:"role" binding:"omitempty,oneof=admin user"`
	Status      string `json:"status" binding:"omitempty,oneof=active disabled"`
}

type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6,max=128"`
}
