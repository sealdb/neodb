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
	"io/ioutil"
	_ "log"
	"os"
	"path"
	"testing"

	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

var (
	neoDBTestJSON = "neodb.test.config.json"
)

func TestWriteConfig(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	tmpDir := getTmpDir("", "neodb_config_", log)
	defer os.RemoveAll(tmpDir)

	conf := &Config{
		Proxy:   MockProxyConfig,
		Log:     MockLogConfig,
		Audit:   DefaultAuditConfig(),
		Router:  DefaultRouterConfig(),
		Monitor: DefaultMonitorConfig(),
		Scatter: DefaultScatterConfig(),
	}

	path := path.Join(tmpDir, neoDBTestJSON)
	err := WriteConfig(path, conf)
	assert.Nil(t, err)

	want, err := LoadConfig(path)
	assert.Nil(t, err)
	assert.Equal(t, want, conf)
}

func TestLoadConfig(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	tmpDir := getTmpDir("", "neodb_config_", log)
	defer os.RemoveAll(tmpDir)

	path := path.Join(tmpDir, neoDBTestJSON)
	{
		_, err := LoadConfig(path)
		assert.NotNil(t, err)
	}

	{
		mockProxyConfig := &ProxyConfig{
			TwopcEnable:    true,
			Endpoint:       ":5566",
			MaxConnections: 1024,
			MetaDir:        "/tmp/neodbmeta",
			PeerAddress:    ":8080",
		}
		conf := &Config{
			Proxy:   mockProxyConfig,
			Audit:   DefaultAuditConfig(),
			Router:  DefaultRouterConfig(),
			Monitor: DefaultMonitorConfig(),
			Log:     MockLogConfig,
			Scatter: DefaultScatterConfig(),
		}

		err := WriteConfig(path, conf)
		assert.Nil(t, err)
		want, err := LoadConfig(path)
		assert.Nil(t, err)
		assert.Equal(t, want, conf)
	}

	{
		mockProxyConfig := &ProxyConfig{
			Endpoint:       ":5566",
			MaxConnections: 1024,
			MetaDir:        "/tmp/neodbmeta",
			PeerAddress:    ":8080",
		}

		conf := &Config{
			Proxy: mockProxyConfig,
			Log:   MockLogConfig,
		}
		err := WriteConfig(path, conf)
		assert.Nil(t, err)
		{
			want := &Config{
				Proxy:   MockProxyConfig,
				Log:     MockLogConfig,
				Audit:   DefaultAuditConfig(),
				Router:  DefaultRouterConfig(),
				Monitor: DefaultMonitorConfig(),
				Scatter: DefaultScatterConfig(),
			}
			got, err := LoadConfig(path)
			assert.Nil(t, err)
			assert.Equal(t, want, got)
		}
	}

	{
		want := &Config{
			Proxy:   MockProxyConfig,
			Log:     MockLogConfig,
			Audit:   DefaultAuditConfig(),
			Router:  DefaultRouterConfig(),
			Monitor: DefaultMonitorConfig(),
			Scatter: DefaultScatterConfig(),
		}

		err := WriteConfig(path, want)
		assert.Nil(t, err)
		got, err := LoadConfig(path)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	}
}

func TestWriteLoadConfig(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	tmpDir := getTmpDir("", "neodb_config_", log)
	defer os.RemoveAll(tmpDir)

	path := path.Join(tmpDir, neoDBTestJSON)

	conf := &Config{
		Proxy: MockProxyConfig,
		Log:   MockLogConfig,
	}

	err := WriteConfig(path, conf)
	assert.Nil(t, err)

	{
		conf, err := LoadConfig(path)
		assert.Nil(t, err)
		want := &Config{
			Proxy:   MockProxyConfig,
			Log:     MockLogConfig,
			Audit:   DefaultAuditConfig(),
			Router:  DefaultRouterConfig(),
			Monitor: DefaultMonitorConfig(),
			Scatter: DefaultScatterConfig(),
		}
		got := conf
		assert.Equal(t, want, got)
	}
}

func TestReadBackendsConfig(t *testing.T) {
	data := `{
       "backends": [
               {
                       "name": "backend1",
                       "address": "127.0.0.1:3304",
                       "user": "root",
                       "password": "",
                       "max-connections": 1024,
                       "role": 0
               }
       ]
}`

	backend, err := ReadBackendsConfig(data)
	assert.Nil(t, err)
	want := &BackendsConfig{Backends: MockBackends}
	got := backend
	assert.Equal(t, want, got)
}

func TestReadTableConfig(t *testing.T) {
	data := `{
	"name": "A",
	"shardtype": "",
	"shardkey": "id",
	"partitions": [
		{
			"table": "A1",
			"segment": "0-2",
			"backend": "backend1"
		},
		{
			"table": "A2",
			"segment": "2-4",
			"backend": "backend1"
		},
		{
			"table": "A3",
			"segment": "4-8",
			"backend": "backend2"
		},
		{
			"table": "A4",
			"segment": "8-16",
			"backend": "backend2"
		}
	],
	"auto-increment":{
		"column":"autoincrement"
	}
}`

	table, err := ReadTableConfig(data)
	assert.Nil(t, err)
	want := MockTablesConfig[0]
	got := table
	assert.Equal(t, want, got)
}

func TestReadTableConfigList(t *testing.T) {
	data := `{
	"name": "L",
	"shardtype": "",
	"shardkey": "id",
	"partitions": [
		{
			"table": "L1",
			"segment": "",
			"backend": "backend2",
			"listvalue": "2"
		}
	]
}`

	table, err := ReadTableConfig(data)
	assert.Nil(t, err)
	want := MockTablesConfig[2]
	got := table
	assert.Equal(t, want, got)
}

func TestRouterConfigUnmarshalJSON(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	tmpDir := getTmpDir("", "neodb_config_", log)
	defer os.RemoveAll(tmpDir)

	path := path.Join(tmpDir, neoDBTestJSON)

	// All nil.
	{
		os.Remove(path)
		data := `{}`
		err := ioutil.WriteFile(path, []byte(data), 0644)
		assert.Nil(t, err)
		got, err := LoadConfig(path)
		assert.Nil(t, err)
		want := &Config{
			Proxy:   DefaultProxyConfig(),
			Router:  DefaultRouterConfig(),
			Audit:   DefaultAuditConfig(),
			Log:     DefaultLogConfig(),
			Monitor: DefaultMonitorConfig(),
			Scatter: DefaultScatterConfig(),
		}
		assert.Equal(t, want, got)
	}

	// Default UnmarshalJSON.
	{
		os.Remove(path)
		data := `{
	"proxy": {
		"endpoint": ":5566",
		"twopc-enable": false,
		"max-connections": 1024
	},
	"audit": {
		"mode": "N",
		"expire-hours": 1
	},
	"router": {
		"blocks-readonly": 64
	},
	"log": {
		"level": "ERROR"
	},
	"xacheck": {
		"xa-check-interval": 1,
		"xa-check-dir":      "/tmp/xacheck"
	}
}`
		err := ioutil.WriteFile(path, []byte(data), 0644)
		assert.Nil(t, err)
		got, err := LoadConfig(path)
		assert.Nil(t, err)

		proxy := DefaultProxyConfig()
		proxy.Endpoint = ":5566"
		want := &Config{
			Proxy:   proxy,
			Router:  DefaultRouterConfig(),
			Audit:   DefaultAuditConfig(),
			Log:     DefaultLogConfig(),
			Monitor: DefaultMonitorConfig(),
			Scatter: DefaultScatterConfig(),
		}
		assert.Equal(t, want, got)
	}
}

func TestReadBackendsConfigAttach(t *testing.T) {
	data := `{
       "backends": [
               {
                       "name": "backend1",
                       "address": "127.0.0.1:3304",
                       "user": "root",
                       "password": "",
                       "max-connections": 1024,
                       "role": 1
               }
       ]
}`

	backends, err := ReadBackendsConfig(data)
	assert.Nil(t, err)
	want := &BackendsConfig{Backends: MockBackends}
	got := backends
	assert.NotEqual(t, want.Backends[0].Role, got.Backends[0].Role)
}
