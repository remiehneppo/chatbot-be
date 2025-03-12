package types

type CreateUserRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	FullName        string `json:"full_name"`
	ManagementLevel int    `json:"management_level"`
	Role            string `json:"role"`
	WorkspaceRole   string `json:"workspace_role"`
	Workspace       string `json:"workspace"`
}

type BatchCreateUserRequest struct {
	Users []CreateUserRequest `json:"users"`
}

type UpdateUserRequest struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	FullName        string `json:"full_name"`
	ManagementLevel int    `json:"management_level"`
	Role            string `json:"role"`
	WorkspaceRole   string `json:"workspace_role"`
	Workspace       string `json:"workspace"`
}

type DeleteUserRequest struct {
	ID string `json:"id"`
}

type GetUserRequest struct {
	ID string `json:"id"`
}

type PaginateUserRequest struct {
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
