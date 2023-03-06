package types

type Ldap struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:64;not null" json:"name"`
	Enable      string `gorm:"size:16;not null" json:"enable"`
	Url         string `gorm:"size:255;not null" json:"url"`
	MaxConn     int    `gorm:"type:uint" json:"max_conn"`
	BaseDN      string `gorm:"size:255;not null" json:"base_dn"`
	AdminDN     string `gorm:"size:64;not null" json:"admin_dn"`
	AdminDNPass string `gorm:"size:64;not null" json:"admin_dn_pass"`
}

func (Ldap) TableName() string {
	return "ldap"
}
