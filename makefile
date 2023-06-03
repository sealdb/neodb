ROOT:=$(shell pwd)
NAME:=github.com/sealdb/neodb
#export GOPATH := $(shell pwd):$(GOPATH)
GIT_TAG=$(shell git describe --tags --always || echo "git not found")
BUILD_TIME=$(shell date "+%Y-%m-%d_%H:%M:%S")
LDFLAGS="-X github.com/sealdb/neodb/version.gitTag=${GIT_TAG} -X github.com/sealdb/neodb/version.buildTime=${BUILD_TIME}"

initmod:
	rm -f go.mod go.sum
	go mod init ${NAME}
	go mod tidy
	go mod vendor

updatemod:
	go mod tidy
	go mod vendor

build:
	@echo "--> Building..."
	@mkdir -p bin/
	go build -v -o bin/neodb --ldflags $(LDFLAGS) neodb/neodb.go
	@chmod 755 bin/*

clean:
	@echo "--> Cleaning..."
	@go clean
	@rm -f bin/*

fmt:
	go fmt ./...
	go vet ./...

test:
	@echo "--> Testing..."
	@$(MAKE) testxbase
	@$(MAKE) testxcontext
	@$(MAKE) testconfig
	@$(MAKE) testrouter
	@$(MAKE) testoptimizer
	@$(MAKE) testplanner
	@$(MAKE) testexecutor
	@$(MAKE) testbackend
	@$(MAKE) testproxy
	@$(MAKE) testaudit
	@$(MAKE) testsyncer
	@$(MAKE) testctl
	@$(MAKE) testmonitor
	@$(MAKE) testplugins
	@$(MAKE) testfuzz

testxbase:
	go test -v -race github.com/sealdb/neodb/xbase
	go test -v -race github.com/sealdb/neodb/xbase/stats
	go test -v -race github.com/sealdb/neodb/xbase/sync2
testxcontext:
	go test -v github.com/sealdb/neodb/xcontext
testconfig:
	go test -v github.com/sealdb/neodb/config
testrouter:
	go test -v github.com/sealdb/neodb/router
testoptimizer:
	go test -v github.com/sealdb/neodb/optimizer
testplanner:
	go test -v github.com/sealdb/neodb/planner/...
testexecutor:
	go test -v github.com/sealdb/neodb/executor/...
testbackend:
	go test -v -race github.com/sealdb/neodb/backend
testproxy:
	go test -v -race github.com/sealdb/neodb/proxy
testaudit:
	go test -v -race github.com/sealdb/neodb/audit
testsyncer:
	go test -v -race github.com/sealdb/neodb/syncer
testctl:
	go test -v -race github.com/sealdb/neodb/ctl/v1
testpoc:
	go test -v poc
testmonitor:
	go test -v github.com/sealdb/neodb/monitor
testplugins:
	go test -v github.com/sealdb/neodb/plugins
	go test -v github.com/sealdb/neodb/plugins/autoincrement
	go test -v github.com/sealdb/neodb/plugins/privilege
	go test -v github.com/sealdb/neodb/plugins/shiftmanager
testmysqlstack:
	cd mysqlstack&&make test

testfuzz:
	go test -v -race github.com/sealdb/neodb/fuzz/sqlparser
testshift:
	cd tools/shift&&make test

# code coverage
allpkgs =	./xbase\
			./ctl/v1/\
			./xcontext\
			./config\
			./router\
			./optimizer\
			./planner/...\
			./executor/...\
			./backend\
			./proxy\
			./audit\
			./syncer\
			./monitor\
			./plugins/...
coverage:
	go build -v -o bin/gotestcover tools/gotestcover/gotestcover.go
	bin/gotestcover -coverprofile=coverage.out -v $(allpkgs)
	go tool cover -html=coverage.out

check:
	go get -v gopkg.in/alecthomas/gometalinter.v2
	bin/gometalinter.v2 -j 4 --disable-all \
	--enable=gofmt \
	--enable=golint \
	--enable=vet \
	--enable=gosimple \
	--enable=unconvert \
	--deadline=10m $(allpkgs) 2>&1 | tee /dev/stderr

.PHONY: build clean install fmt test coverage check
