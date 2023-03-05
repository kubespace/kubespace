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

type ImageRegistrySerializers struct {
	Registry string `json:"registry" form:"registry"`
	User     string `json:"user" form:"user"`
	Password string `json:"password" form:"password"`
}

type LdapSerializers struct {
	Name        string `json:"name" form:"name"`
	Enable      string `json:"enable" form:"enable"`
	Url         string `json:"url" form:"url"`
	BaseDN      string `json:"baseDN" form:"baseDN"`
	AdminDN     string `json:"adminDN" form:"adminDN"`
	AdminDNPass string `json:"adminDNPass" form:"adminDNPass"`
}
