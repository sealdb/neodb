module github.com/sealdb/neodb

go 1.19

require github.com/shopspring/decimal v1.2.0 // indirect

require (
	github.com/ant0ine/go-json-rest v3.3.2+incompatible
	github.com/beefsack/go-rate v0.0.0-20220214233405-116f4ca011a0
	github.com/dvyukov/go-fuzz-corpus v0.0.0-20190920191254-c42c1b2914c7
	github.com/fortytw2/leaktest v1.3.0
	github.com/juju/errors v1.0.0
	github.com/lithammer/go-jump-consistent-hash v1.0.2
	github.com/mailru/easyjson v0.7.7
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.15.1
	github.com/prometheus/client_model v0.4.0
	github.com/sealdb/go-mysql v1.0.1
	github.com/sealdb/mysqlstack v1.0.1
	github.com/stretchr/testify v1.8.4
	golang.org/x/sync v0.2.0
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/siddontang/go v0.0.0-20180604090527-bdc77568d726 // indirect
	github.com/siddontang/go-log v0.0.0-20180807004314-8d05993dda07 // indirect
	golang.org/x/sys v0.6.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// replace github.com/sealdb/neodb/xbase => ./xbase
// replace github.com/siddontang/go-mysql => github.com/go-mysql-org/go-mysql v1.7.0
