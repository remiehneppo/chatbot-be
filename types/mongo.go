package types

const (
	USER_ROLE_ADMIN = "admin"
	USER_ROLE_USER  = "user"
)
const (
	USER_WORKSPACE_ROLE_EXECUTIVE = "executive"
	USER_WORKSPACE_ROLE_HEAD      = "head"
	USER_WORKSPACE_ROLE_DHEAD     = "dhead"
	USER_WORKSPACE_ROLE_ASSISTANT = "assistant"
	USER_WORKSPACE_ROLE_STAFF     = "staff"
)

const (
	USER_MANAGEMENT_LEVEL_EXECUTIVE = 5
	USER_MANAGEMENT_LEVEL_HEAD      = 4
	USER_MANAGEMENT_LEVEL_DHEAD     = 3
	USER_MANAGEMENT_LEVEL_ASSISTANT = 2
	USER_MANAGEMENT_LEVEL_STAFF     = 2
)

const (
	TASK_STATUS_OPEN      = "open"
	TASK_STATUS_CLOSE     = "close"
	TASK_STATUS_CANCEL    = "cancel"
	TASK_STATUS_DOING     = "doing"
	TASK_STATUS_COMPLETED = "completed"
	TASK_STATUS_REVIEW    = "review"
)

const (
	DepartmentTechnical      = "DepartmentTechnical"
	DepartmentProductionPlan = "DepartmentProductionPlan"
	DepartmentQuality        = "DepartmentQuality"
	DepartmentMaterial       = "DepartmentMaterial"
)

type User struct {
	ID              string `json:"id" bson:"_id,omitempty"`
	Username        string `json:"username" bson:"username"`
	Password        string `json:"password" bson:"password"`
	FullName        string `json:"full_name" bson:"full_name"`
	ManagementLevel int    `json:"management_level" bson:"management_level"`
	Role            string `json:"role" bson:"role"`
	WorkspaceRole   string `json:"workspace_role" bson:"workspace_role"`
	Workspace       string `json:"workspace" bson:"workspace"`
	CreateAt        int64  `json:"created_at" bson:"created_at"`
	UpdateAt        int64  `json:"updated_at" bson:"updated_at"`
}

type Workspace struct {
	ID   string `json:"id" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name"`
}

type Task struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	Workspace   string `json:"workspace" bson:"workspace"`
	Creator     string `json:"creator" bson:"creator"`
	Deadline    int64  `json:"deadline" bson:"deadline"`
	Assignee    string `json:"assignee" bson:"assignee"`
	Reporter    string `json:"reporter" bson:"reporter"`
	Status      string `json:"status" bson:"status"`
	CreateAt    int64  `json:"created_at" bson:"created_at"`
	UpdateAt    int64  `json:"updated_at" bson:"updated_at"`
	Report      string `json:"report" bson:"report"`
}
