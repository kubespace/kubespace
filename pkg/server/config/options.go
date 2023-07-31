package config

type ServerOptions struct {
	InsecurePort         int
	Port                 int
	RedisAddress         string
	RedisDB              int
	RedisPassword        string
	CertFilePath         string
	KeyFilePath          string
	MysqlHost            string
	MysqlUser            string
	MysqlPassword        string
	MysqlDbName          string
	ListWatcherResyncSec int
	AgentVersion         string
	AgentRepository      string
	ReleaseVersion       string
}
