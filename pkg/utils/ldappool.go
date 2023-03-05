package utils

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"k8s.io/klog/v2"
	"sync"
)

type LDAPPool struct {
	config *LdapConfig
	pool   chan *ldap.Conn // 连接池对象
	lock   sync.Mutex      // 互斥锁，用于保护连接池
}

type LdapError struct {
	Message string
}

func (e *LdapError) Error() string {
	return fmt.Sprintf("Error: %s", e.Message)
}

var ldapPools = make(map[string]*LDAPPool)

const MaxLdapPoolSize int = 3

func getLDAPPool(ldconfig *LdapConfig) (*LDAPPool, error) {

	// 如果连接池已经存在，则直接返回连接池实例
	if pool, ok := ldapPools[ldconfig.Url]; ok {
		return pool, nil
	}

	pool := &LDAPPool{
		config: ldconfig,
		pool:   make(chan *ldap.Conn, MaxLdapPoolSize),
	}

	for i := 0; i < MaxLdapPoolSize; i++ {
		conn, err := ldap.Dial("tcp", ldconfig.Url)
		if err != nil {
			return nil, err
		}

		if err := conn.Bind(ldconfig.GetUserDN(), ldconfig.Password); err != nil {
			conn.Close()
			return nil, err
		}

		pool.pool <- conn
	}

	ldapPools[ldconfig.Url] = pool

	return pool, nil
}

func WithLDAPConn(config *LdapConfig, params interface{}, fn func(conn *ldap.Conn, params interface{}) error) error {
	pool, err := getLDAPPool(config)
	if err != nil {
		return err
	}

	conn := pool.getConn()
	defer pool.putConn(conn)

	return fn(conn, params)
}

func AuthenticationFunc(conn *ldap.Conn, params interface{}) error {

	p, ok := params.(*LdapConfig)
	if !ok {
		klog.Fatal("params trans error")
		return &LdapError{Message: "params trans error"}
	}

	searchFilter := "(uid=" + p.User + ")"
	searchRequest := ldap.NewSearchRequest(
		p.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		[]string{"dn"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		klog.Fatal(err)
		return err
	}

	if len(sr.Entries) != 1 {
		klog.Fatal("User does not exist or too many entries returned")
		return &LdapError{Message: "User does not exist or too many entries returned"}
	}

	userDN := sr.Entries[0].DN

	err = conn.Bind(userDN, p.Password)
	if err != nil {
		klog.Fatal(err)
		return err
	}

	klog.Infof("Authentication successful!")
	return nil
}

func SearchLdapUsersFunc(conn *ldap.Conn, params interface{}) error {

	return nil
}

type LdapConfig struct {
	Url      string
	User     string
	Password string
	BaseDN   string
}

func (lc *LdapConfig) GetUserDN() string {
	return "cn=" + lc.User + "," + lc.BaseDN
}

func (p *LDAPPool) getConn() *ldap.Conn {
	p.lock.Lock()
	defer p.lock.Unlock()

	select {
	case conn := <-p.pool:
		if !conn.IsClosing() {
			return conn
		}
		conn.Close()
		return p.createConn()
	default:
		return p.createConn()
	}
}

func (p *LDAPPool) putConn(conn *ldap.Conn) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.pool) >= cap(p.pool) {
		conn.Close()
		return
	}

	p.pool <- conn
}

func (p *LDAPPool) createConn() *ldap.Conn {
	conn, err := ldap.Dial("tcp", p.config.Url)
	if err != nil {
		return nil
	}

	if err := conn.Bind(p.config.GetUserDN(), p.config.Password); err != nil {
		conn.Close()
		return nil
	}

	return conn
}
