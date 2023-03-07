package ldap

import (
	"fmt"
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
	//err, _ := WithLDAPConn(config, userConfig, AuthenticationFunc)
	//if err != nil {
	//	t.Errorf("err(%v)", err)
	//}

	// ldap 查询功能
	err1, re := WithLDAPConn(config, userConfig, SearchLdapUsersFunc)
	if err1 != nil {
		t.Errorf("err(%v)", err1)
	}

	fmt.Println(re)

}
