/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/sealdb/neodb/xbase"

	"github.com/pkg/errors"
)

const (
	NormalBackend = 0
	AttachBackend = 1
)

// ProxyConfig tuple.
type ProxyConfig struct {
	IPS                 []string `json:"allowip"`
	MetaDir             string   `json:"meta-dir"`
	Endpoint            string   `json:"endpoint"`
	TwopcEnable         bool     `json:"twopc-enable"`
	LoadBalance         int      `json:"load-balance"`           // 0 -- disable balance, 1 -- enable balance to replica
	LowerCaseTableNames int      `json:"lower-case-table-names"` // 0 -- case sensitive, 1 -- case insensitive

	MaxConnections   int    `json:"max-connections"`
	MaxResultSize    int    `json:"max-result-size"`
	MaxJoinRows      int    `json:"max-join-rows"`
	DDLTimeout       int    `json:"ddl-timeout"`
	QueryTimeout     int    `json:"query-timeout"`
	PeerAddress      string `json:"peer-address,omitempty"`
	LongQueryTime    int    `json:"long-query-time"`
	StreamBufferSize int    `json:"stream-buffer-size"`
	IdleTxnTimeout   uint32 `json:"kill-idle-transaction"` //is consistent with the official 8.0 kill_idle_transaction

	//If autocommit-false-is-txn=true (false by default), a client connection with cmd: set autocommit=0
	//is treated as start a transaction, e.g. begin, start transaction.
	AutocommitFalseIsTxn bool `json:"autocommit-false-is-txn"`
}

// DefaultProxyConfig returns default proxy config.
func DefaultProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		MetaDir:             "./neodb-meta",
		Endpoint:            "127.0.0.1:3308",
		LoadBalance:         0,
		LowerCaseTableNames: 0,
		MaxConnections:      1024,
		MaxResultSize:       1024 * 1024 * 1024, // 1GB
		MaxJoinRows:         32768,
		DDLTimeout:          10 * 3600 * 1000, // 10hours
		QueryTimeout:        5 * 60 * 1000,    // 5minutes
		PeerAddress:         "127.0.0.1:8080",
		LongQueryTime:       5,                // 5 seconds
		StreamBufferSize:    1024 * 1024 * 32, // 32MB
		IdleTxnTimeout:      60,               // 60 seconds
	}
}

// UnmarshalJSON interface on ProxyConfig.
func (c *ProxyConfig) UnmarshalJSON(b []byte) error {
	type confAlias *ProxyConfig
	conf := confAlias(DefaultProxyConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = ProxyConfig(*conf)
	return nil
}

// AuditConfig tuple.
type AuditConfig struct {
	Mode        string `json:"mode"`
	LogDir      string `json:"audit-dir"`
	MaxSize     int    `json:"max-size"`
	ExpireHours int    `json:"expire-hours"`
}

// DefaultAuditConfig returns default audit config.
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		Mode:        "N",
		LogDir:      "/tmp/auditlog",
		MaxSize:     1024 * 1024 * 256, // 256MB
		ExpireHours: 1,                 // 1hours
	}
}

// UnmarshalJSON interface on AuditConfig.
func (c *AuditConfig) UnmarshalJSON(b []byte) error {
	type confAlias *AuditConfig
	conf := confAlias(DefaultAuditConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = AuditConfig(*conf)
	return nil
}

// LogConfig tuple.
type LogConfig struct {
	Level string `json:"level"`
}

// DefaultLogConfig returns default log config.
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level: "ERROR",
	}
}

// UnmarshalJSON interface on LogConfig.
func (c *LogConfig) UnmarshalJSON(b []byte) error {
	type confAlias *LogConfig
	conf := confAlias(DefaultLogConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = LogConfig(*conf)
	return nil
}

// MonitorConfig tuple
type MonitorConfig struct {
	MonitorAddress string `json:"monitor-address"`
}

// DefaultMonitorConfig returns default monitor config.
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		MonitorAddress: "0.0.0.0:13380",
	}
}

// UnmarshalJSON interface on MonitorConfig.
func (c *MonitorConfig) UnmarshalJSON(b []byte) error {
	type confAlias *MonitorConfig
	conf := confAlias(DefaultMonitorConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = MonitorConfig(*conf)
	return nil
}

// BackendConfig tuple.
type BackendConfig struct {
	Name           string `json:"name"`
	Address        string `json:"address"`
	Replica        string `json:"replica-address"`
	User           string `json:"user"`
	Password       string `json:"password"`
	DBName         string `json:"database"`
	Charset        string `json:"charset"`
	MaxConnections int    `json:"max-connections"`
	Role           int    `json:"role"`
}

// BackendsConfig tuple.
type BackendsConfig struct {
	Backends []*BackendConfig `json:"backends"`
}

// PartitionConfig tuple.
type PartitionConfig struct {
	Table     string `json:"table"`
	Segment   string `json:"segment"`
	Backend   string `json:"backend"`
	ListValue string `json:"listvalue"`
}

// AutoIncrement tuple.
type AutoIncrement struct {
	Column string `json:"column"`
}

// TableConfig tuple.
type TableConfig struct {
	Name          string             `json:"name"`
	Slots         int                `json:"slots-readonly"`
	Blocks        int                `json:"blocks-readonly"`
	ShardType     string             `json:"shardtype"`
	ShardKey      string             `json:"shardkey"`
	Partitions    []*PartitionConfig `json:"partitions"`
	AutoIncrement *AutoIncrement     `json:"auto-increment,omitempty"`
}

// SchemaConfig tuple.
type SchemaConfig struct {
	DB     string         `json:"database"`
	Tables []*TableConfig `json:"tables"`
}

// RouterConfig tuple.
type RouterConfig struct {
	Slots  int `json:"slots-readonly"`
	Blocks int `json:"blocks-readonly"`
}

// DefaultRouterConfig returns the default router config.
func DefaultRouterConfig() *RouterConfig {
	return &RouterConfig{
		Slots:  4096,
		Blocks: 64,
	}
}

// UnmarshalJSON interface on RouterConfig.
func (c *RouterConfig) UnmarshalJSON(b []byte) error {
	type confAlias *RouterConfig
	conf := confAlias(DefaultRouterConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = RouterConfig(*conf)
	return nil
}

// ScatterConfig tuple.
type ScatterConfig struct {
	XaCheckInterval int    `json:"xa-check-interval"`
	XaCheckDir      string `json:"xa-check-dir"`
	XaCheckRetrys   int    `json:"xa-check-retrys"`
}

// DefaultScatterConfig returns default ScatterConfig config.
func DefaultScatterConfig() *ScatterConfig {
	return &ScatterConfig{
		XaCheckInterval: 10,
		XaCheckDir:      "./xacheck", //In the production environment, don't set the tmp dir
		XaCheckRetrys:   10,
	}
}

// UnmarshalJSON interface on XaCheckConfig.
func (c *ScatterConfig) UnmarshalJSON(b []byte) error {
	type confAlias *ScatterConfig
	conf := confAlias(DefaultScatterConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = ScatterConfig(*conf)
	return nil
}

// Config tuple.
type Config struct {
	Proxy   *ProxyConfig   `json:"proxy"`
	Audit   *AuditConfig   `json:"audit"`
	Router  *RouterConfig  `json:"router"`
	Log     *LogConfig     `json:"log"`
	Monitor *MonitorConfig `json:"monitor"`
	Scatter *ScatterConfig `json:"scatter"`
}

func checkConfig(conf *Config) {
	if conf.Proxy == nil {
		conf.Proxy = DefaultProxyConfig()
	}

	if conf.Audit == nil {
		conf.Audit = DefaultAuditConfig()
	}

	if conf.Router == nil {
		conf.Router = DefaultRouterConfig()
	}

	if conf.Log == nil {
		conf.Log = DefaultLogConfig()
	}

	if conf.Monitor == nil {
		conf.Monitor = DefaultMonitorConfig()
	}

	if conf.Scatter == nil {
		conf.Scatter = DefaultScatterConfig()
	}
}

// LoadConfig used to load the config from file.
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	conf := &Config{}
	if err := json.Unmarshal([]byte(data), conf); err != nil {
		return nil, errors.WithStack(err)
	}
	checkConfig(conf)
	return conf, nil
}

// ReadTableConfig used to read the table config from the data.
func ReadTableConfig(data string) (*TableConfig, error) {
	conf := &TableConfig{}
	if err := json.Unmarshal([]byte(data), conf); err != nil {
		return nil, errors.WithStack(err)
	}
	return conf, nil
}

// ReadBackendsConfig used to read the backend config from the data.
func ReadBackendsConfig(data string) (*BackendsConfig, error) {
	conf := &BackendsConfig{}
	if err := json.Unmarshal([]byte(data), conf); err != nil {
		return nil, errors.WithStack(err)
	}
	return conf, nil
}

// WriteConfig used to write the conf to file.
func WriteConfig(path string, conf interface{}) error {
	b, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return errors.WithStack(err)
	}
	return xbase.WriteFile(path, b)
}
