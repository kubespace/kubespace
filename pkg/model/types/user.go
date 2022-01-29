package types

const (
	ADMIN = "admin"
)

type UserStore struct {
	Common
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Status    string `json:"status"`
	IsSuper   string `json:"is_super"`
	Roles     string `json:"roles"`
	LastLogin string `json:"last_login"`
}

type User struct {
	Common
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	Status    string   `json:"status"`
	IsSuper   bool     `json:"is_super"`
	Roles     []string `json:"roles"`
	LastLogin string   `json:"last_login"`
}
