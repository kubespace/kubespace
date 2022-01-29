package serializers

type SecretsSerializers struct {
	Name string		`json:"name" form:"name"`
	Kind string		`json:"kind" form:"kind"`
	SecretType string `json:"secret_type" form:"secret_type"`
	User string `json:"user" form:"user"`
	Password string `json:"password" form:"password"`
	PrivateKey string `json:"private_key" form:"private_key"`
	AccessToken string `json:"access_token" form:"access_token"`
}
