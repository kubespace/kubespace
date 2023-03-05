package utils

import (
	"testing"
)

func TestLdapPool(t *testing.T) {

	config := &LdapConfig{
		Url:      "192.168.102.54:389",
		User:     "admin",
		Password: "admin",
		BaseDN:   "dc=zgj,dc=com",
	}

	userConfig := &LdapConfig{
		User:     "zhangsan",
		Password: "123456",
		BaseDN:   config.BaseDN,
	}

	// ldap认证功能
	err := WithLDAPConn(config, userConfig, AuthenticationFunc)
	if err != nil {
		t.Errorf("err(%v)", err)
	}

	// ldap 查询功能
	err = WithLDAPConn(config, userConfig, SearchLdapUsersFunc)
	if err != nil {
		t.Errorf("err(%v)", err)
	}

}
