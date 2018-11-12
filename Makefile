
# golang1.9 or latest
# 1. make help
# 2. make dep
# 3. make build
# ...

CLI := build/chain33-cli
SRC_CLI := github.com/33cn/plugin/cli
APP := build/chain33
CHAIN33=github.com/33cn/chain33
CHAIN33_PATH=vendor/${CHAIN33}
AUTO_TEST := build/tools/autotest/autotest
SRC_AUTO_TEST := ${CHAIN33}/cmd/autotest
LDFLAGS := -ldflags "-w -s"
PKG_LIST := `go list ./... | grep -v "vendor" | grep -v "chain33/test" | grep -v "mocks" | grep -v "pbft"`
PKG_LIST_Q := `go list ./... | grep -v "vendor" | grep -v "chain33/test" | grep -v "mocks" | grep -v "blockchain" | grep -v "pbft"`
BUILD_FLAGS = -ldflags "-X github.com/33cn/chain33/common/version.GitCommit=`git rev-parse --short=8 HEAD`"
MKPATH=$(abspath $(lastword $(MAKEFILE_LIST)))
MKDIR=$(dir $(MKPATH))
DAPP := ""
PROJ := "build"
.PHONY: default dep all build release cli linter race test fmt vet bench msan coverage coverhtml docker docker-compose protobuf clean help autotest

default: build depends

build:
	@go build $(BUILD_FLAGS) -v -i -o $(APP)
	@go build -v -i -o $(CLI) $(SRC_CLI)
	@cp chain33.toml build/

build_ci: depends ## Build the binary file for CI
	@go build -v -i -o $(CLI) $(SRC_CLI)
	@go build $(BUILD_FLAGS) -v -o $(APP)
	@cp chain33.toml build/

para:
	@go build -v -o build/$(NAME) -ldflags "-X $(SRC_CLI)/buildflags.ParaName=user.p.$(NAME). -X $(SRC_CLI)/buildflags.RPCAddr=http://localhost:8901" $(SRC_CLI)


autotest:## build autotest binary
	@go build -v -i -o $(AUTO_TEST) $(SRC_AUTO_TEST)
	@cp cmd/autotest/*.toml build/tools/autotest/
	@if [ -n "$(dapp)" ]; then \
		cd build/tools/autotest && bash ./local-autotest.sh $(dapp) && cd ../../../; \
	fi

update:
	rm -rf ${CHAIN33_PATH}
	git clone --depth 1 -b master https://${CHAIN33}.git ${CHAIN33_PATH}
	rm -rf vendor/${CHAIN33}/.git
	rm -rf vendor/${CHAIN33}/vendor/github.com/apache/thrift/tutorial/erl/
	cp -Rf vendor/${CHAIN33}/vendor/* vendor/
	rm -rf vendor/${CHAIN33}/vendor
	govendor init
	go build -i -o tool github.com/33cn/plugin/vendor/github.com/33cn/chain33/cmd/tools
	./tool import --path "plugin" --packname "github.com/33cn/plugin/plugin" --conf ""

updatevendor:
	govendor add +e
	govendor fetch -v +m

dep:
	dep init -v


linter: ## Use gometalinter check code, ignore some unserious warning
	@res=$$(gometalinter.v2 -t --sort=linter --enable-gc --deadline=2m --disable-all \
	--enable=gofmt \
	--enable=gosimple \
	--enable=deadcode \
	--enable=unconvert \
	--enable=interfacer \
	--enable=varcheck \
	--enable=structcheck \
	--enable=goimports \
	--vendor ./...) \
#	--enable=vet \
#	--enable=staticcheck \
#	--enable=gocyclo \
#	--enable=staticcheck \
#	--enable=golint \
#	--enable=unused \
#	--enable=gotype \
#	--enable=gotypex \
	if [ -n "$$res" ]; then \
		echo "$${res}"; \
		exit 1; \
		fi;
	@find . -name '*.sh' -not -path "./vendor/*" | xargs shellcheck

race: ## Run data race detector
	@go test -race -short $(PKG_LIST)

test: ## Run unittests
	@go test -race $(PKG_LIST)


fmt: fmt_proto fmt_shell ## go fmt
	@go fmt ./...
	@find . -name '*.go' -not -path "./vendor/*" | xargs goimports -l -w

.PHONY: fmt_proto fmt_shell
fmt_proto: ## go fmt protobuf file
	@find . -name '*.proto' -not -path "./vendor/*" | xargs clang-format -i

fmt_shell: ## check shell file
	@find . -name '*.sh' -not -path "./vendor/*" | xargs shfmt -w -s -i 4 -ci -bn

fmt_go: fmt_shell ## go fmt
	@go fmt ./...
	@find . -name '*.go' -not -path "./vendor/*" | xargs goimports -l -w


coverage: ## Generate global code coverage report
	@./build/tools/coverage.sh;

coverhtml: ## Generate global code coverage report in HTML
	@./build/tools/coverage.sh html;

docker: ## build docker image for chain33 run
	@sudo docker build . -f ./build/Dockerfile-run -t chain33:latest

docker-compose: ## build docker-compose for chain33 run
	@cd build && if ! [ -d ci ]; then \
	 make -C ../ ; \
	 fi; \
	 cp chain33* Dockerfile  docker-compose* ci/ && cd ci/ && ./docker-compose-pre.sh run $(PROJ) $(DAPP)  && cd ../..

docker-compose-down: ## build docker-compose for chain33 run
	@cd build && if [ -d ci ]; then \
	 cp chain33* Dockerfile  docker-compose* ci/ && cd ci/ && ./docker-compose-pre.sh down $(PROJ) $(DAPP) && cd .. ; \
	 fi; \
	 cd ..

fork-test: ## build fork-test for chain33 run
	@cd build && cp chain33* Dockerfile system-fork-test.sh docker-compose* ci/ && cd ci/ && ./docker-compose-pre.sh forktest $(PROJ) $(DAPP) && cd ../..


clean: ## Remove previous build
	@rm -rf $(shell find . -name 'datadir' -not -path "./vendor/*")
	@rm -rf build/chain33*
	@rm -rf build/relayd*
	@rm -rf build/*.log
	@rm -rf build/logs
	@rm -rf build/tools/autotest/autotest
	@rm -rf build/ci
	@rm -rf tool
	@go clean

proto:protobuf

protobuf: ## Generate protbuf file of types package
#	@cd ${CHAIN33_PATH}/types/proto && ./create_protobuf.sh && cd ../..
	@find ./plugin/dapp -maxdepth 2 -type d  -name proto -exec make -C {} \;

depends: ## Generate depends file of types package
	@find ./plugin/dapp -maxdepth 2 -type d  -name cmd -exec make -C {} OUT="$(MKDIR)build/ci" FLAG= \;
	@find ./vendor/github.com/33cn/chain33/system/dapp -maxdepth 2 -type d  -name cmd -exec make -C {} OUT="$(MKDIR)build/ci" FLAG= \;

help: ## Display this help screen
	@printf "Help doc:\nUsage: make [command]\n"
	@printf "[command]\n"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

cleandata:
	rm -rf build/datadir/addrbook
	rm -rf build/datadir/blockchain.db
	rm -rf build/datadir/mavltree
	rm -rf build/chain33.log

.PHONY: checkgofmt
checkgofmt: ## get all go files and run go fmt on them
	@files=$$(find . -name '*.go' -not -path "./vendor/*" | xargs gofmt -l -s); if [ -n "$$files" ]; then \
		  echo "Error: 'make fmt' needs to be run on:"; \
		  echo "${files}"; \
		  exit 1; \
		  fi;
	@files=$$(find . -name '*.go' -not -path "./vendor/*" | xargs goimports -l -w); if [ -n "$$files" ]; then \
		  echo "Error: 'make fmt' needs to be run on:"; \
		  echo "${files}"; \
		  exit 1; \
		  fi;

.PHONY: mock
mock:
	@cd client && mockery -name=QueueProtocolAPI && mv mocks/QueueProtocolAPI.go mocks/api.go && cd -
	@cd queue && mockery -name=Client && mv mocks/Client.go mocks/client.go && cd -
	@cd common/db && mockery -name=KV && mv mocks/KV.go mocks/kv.go && cd -
	@cd common/db && mockery -name=KVDB && mv mocks/KVDB.go mocks/kvdb.go && cd -
	@cd types/ && mockery -name=Chain33Client && mv mocks/Chain33Client.go mocks/chain33client.go && cd -


.PHONY: auto_ci_before auto_ci_after auto_ci
auto_ci_before: clean fmt protobuf
	@echo "auto_ci"
	@go version
	@protoc --version
	@mockery -version
	@docker version
	@docker-compose version
	@git version
	@git status

.PHONY: auto_ci_after
auto_ci_after: clean fmt protobuf
	@git add *.go *.sh *.proto
	@git status
	@files=$$(git status -suno);if [ -n "$$files" ]; then \
		  git status; \
		  git commit -a -m "auto ci [ci-skip]"; \
		  git push origin HEAD:$(branch); \
		  fi;

.PHONY: auto_ci
auto_fmt := find . -name '*.go' -not -path './vendor/*' | xargs goimports -l -w
auto_ci: clean fmt_proto fmt_shell protobuf
	@-find . -name '*.go' -not -path './vendor/*' | xargs gofmt -l -w -s
	@-${auto_fmt}
	@-find . -name '*.go' -not -path './vendor/*' | xargs gofmt -l -w -s
	@${auto_fmt}
	@git add -u
	@git status
	@files=$$(git status -suno); if [ -n "$$files" ]; then \
		  git add -u; \
		  git status; \
		  git commit -a -m "auto ci"; \
		  git push origin HEAD:$(branch); \
		  exit 1; \
		  fi;


