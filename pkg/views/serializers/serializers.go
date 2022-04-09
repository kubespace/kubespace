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
