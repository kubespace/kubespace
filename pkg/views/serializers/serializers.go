package serializers

type UserCreateSerializers struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

type UserSerializers struct {
	UserName string   `json:"username"`
	Password string   `json:"password"`
	Email    string   `json:"email"`
	Status   string   `json:"status"`
	Roles    []string `json:"roles"`
}

type ClusterCreateSerializers struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

type DeleteClusterSerializers struct {
	Id string `json:"name"`
}

type DeleteUserSerializers struct {
	Name string `json:"name"`
}

type DeleteRoleSerializers struct {
	Name string `json:"name"`
}

type ApplyYamlSerializers struct {
	YamlStr string `json:"yaml"`
}

type UserRoleSerializers struct {
	UserId  uint   `json:"user_id" form:"user_id"`
	Scope   string `json:"scope" form:"scope"`
	ScopeId uint   `json:"scope_id" form:"scope_id"`
	Role    string `json:"role" form:"from"`
}

type UserRoleUpdateSerializers struct {
	UserIds []uint `json:"user_ids" form:"user_ids"`
	Scope   string `json:"scope" form:"scope"`
	ScopeId uint   `json:"scope_id" form:"scope_id"`
	Role    string `json:"role" form:"from"`
}
