package serializers

type SecretsSerializers struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	SecretType  string `json:"secret_type" form:"secret_type"`
	User        string `json:"user" form:"user"`
	Password    string `json:"password" form:"password"`
	PrivateKey  string `json:"private_key" form:"private_key"`
	AccessToken string `json:"access_token" form:"access_token"`
}

type SettingsSerializers struct {
	Type    string      `json:"type" form:"type"`
	Scope   string      `json:"scope" form:"scope"`
	ScopeId string      `json:"scope_id" form:"scope_id"`
	Key     string      `json:"key" form:"key"`
	Value   interface{} `json:"value" form:"value"`
}
