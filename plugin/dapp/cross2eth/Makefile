
# golang1.16+
# 1. make help
# 2. make dep
# 3. make build
# ...

SRC_EBCLI := github.com/33cn/plugin/plugin/dapp/cross2eth/ebcli
SRC_EBRELAYER := github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer
CLI_A := build/ebcli_A
CLI_B := build/ebcli_B
CLI_C := build/ebcli_C
CLI_D := build/ebcli_D
EBRELAER := build/ebrelayer   ##通过配置文件启动不同的ebrelayer
LDFLAGS := ' -w -s'
BUILDTIME:=$(shell date +"%Y-%m-%d %H:%M:%S %A")
VERSION=$(shell git describe --tags || git rev-parse --short=8 HEAD)
GitCommit=$(shell git rev-parse HEAD)
BUILD_FLAGS := -ldflags '-X "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/version.GitCommit=$(GitCommit)" \
                         -X "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/version.BuildTime=$(BUILDTIME)" \
                         -X "github.com/33cn/plugin/version.Version=$(VERSION)" '

BUILD_FLAGSB := -ldflags '-X "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/version.GitCommit=$(GitCommit)" \
                         -X "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/version.BuildTime=$(BUILDTIME)" \
                         -X "github.com/33cn/plugin/version.Version=$(VERSION)" \
                         -X $(SRC_EBCLI)/buildflags.RPCAddr=http://localhost:9902" '

BUILD_FLAGSC := -ldflags '-X "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/version.GitCommit=$(GitCommit)" \
                         -X "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/version.BuildTime=$(BUILDTIME)" \
                         -X "github.com/33cn/plugin/version.Version=$(VERSION)" \
                         -X $(SRC_EBCLI)/buildflags.RPCAddr=http://localhost:9903" '

BUILD_FLAGSD := -ldflags '-X "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/version.GitCommit=$(GitCommit)" \
                         -X "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/version.BuildTime=$(BUILDTIME)" \
                         -X "github.com/33cn/plugin/version.Version=$(VERSION)" \
                         -X $(SRC_EBCLI)/buildflags.RPCAddr=http://localhost:9904" '

proj := "build"
.PHONY: default dep all build release cli linter race test fmt vet bench msan coverage coverhtml docker docker-compose protobuf clean help autotest

default: build

build:
	@go build $(BUILD_FLAGS) -v -i -o $(EBRELAER) $(SRC_EBRELAYER)
	@go build $(BUILD_FLAGS) -v -i -o $(CLI_A) $(SRC_EBCLI)
	@go build $(BUILD_FLAGSB) -v -i -o $(CLI_B) $(SRC_EBCLI)
	@go build $(BUILD_FLAGSC) -v -i -o $(CLI_C) $(SRC_EBCLI)
	@go build $(BUILD_FLAGSD) -v -i -o $(CLI_D) $(SRC_EBCLI)
	@cp ebrelayer/relayer.toml build/

rebuild:
	make -C ebrelayer/ethcontract
	make build

cli:
	@go build $(BUILD_FLAGS) -v -i -o $(CLI_A) $(SRC_EBCLI)
	@go build $(BUILD_FLAGSB) -v -i -o $(CLI_B) $(SRC_EBCLI)
	@go build $(BUILD_FLAGSC) -v -i -o $(CLI_C) $(SRC_EBCLI)
	@go build $(BUILD_FLAGSD) -v -i -o $(CLI_D) $(SRC_EBCLI)

build_ci: depends ## Build the binary file for CI
	@go build -v -i -o $(CLI) $(SRC_EBCLI)
	@go build $(BUILD_FLAGS) -v -o $(APP)
	@cp chain33.toml build/

para:
	@go build -v -o build/$(NAME) -ldflags "-X $(SRC_EBCLI)/buildflags.ParaName=user.p.$(NAME). -X $(SRC_EBCLI)/buildflags.RPCAddr=http://localhost:8901" $(SRC_EBCLI)

vet:
	@go vet ${PKG_LIST_VET}


race: ## Run data race detector
	@go test -race -short $(PKG_LIST)

test: ## Run unittests
	@go test -race $(PKG_LIST)

fmt: fmt_proto fmt_shell ## go fmt
	@go fmt ./...
	@find . -name '*.go' -not -path "./vendor/*" | xargs goimports -l -w

.PHONY: fmt_proto fmt_shell
fmt_proto: ## go fmt protobuf file
	#@find . -name '*.proto' -not -path "./vendor/*" | xargs clang-format -i

fmt_shell: ## check shell file
	@find . -name '*.sh' -not -path "./vendor/*" | xargs shfmt -w -s -i 4 -ci -bn

fmt_go: fmt_shell ## go fmt
	@go fmt ./...
	@find . -name '*.go' -not -path "./vendor/*" | xargs goimports -l -w

docker: ## build docker image for chain33 run
	@sudo docker build . -f ./build/Dockerfile-run -t chain33:latest

docker-compose: ## build docker-compose for chain33 run
	@cd build && if ! [ -d ci ]; then \
	 make -C ../ ; \
	 fi; \
	 cp chain33* Dockerfile  docker-compose* ci/ && cd ci/ && ./docker-compose-pre.sh run $(proj) $(dapp)  && cd ../..

docker-compose-down: ## build docker-compose for chain33 run
	@cd build && if [ -d ci ]; then \
	 cp chain33* Dockerfile  docker-compose* ci/ && cd ci/ && ./docker-compose-pre.sh down $(proj) $(dapp) && cd .. ; \
	 fi; \
	 cd ..

clean: ## remove all the bins
	@rm -rf $(CLI_A)
	@rm -rf $(CLI_B)
	@rm -rf $(CLI_C)
	@rm -rf $(CLI_D)
	@rm -rf $(EBRELAER)
	@rm -rf ./build


proto:protobuf

protobuf: ## Generate protbuf file of types package
#	@cd ${CHAIN33_PATH}/types/proto && ./create_protobuf.sh && cd ../..
	@find ./plugin/dapp -maxdepth 2 -type d  -name proto -exec make -C {} \;


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

