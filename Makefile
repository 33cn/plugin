CHAIN33=gitlab.33.cn/chain33/chain33
MKPATH=$(abspath $(lastword $(MAKEFILE_LIST)))
MKDIR=$(dir $(MKPATH))
PKG_LIST := `go list ./... | grep -v "vendor" | grep -v "chain33/test" | grep -v "mocks" | grep -v "pbft"`

all: build

build:vendor
	go build -i -o chain33
	go build -i -o chain33-cli gitlab.33.cn/chain33/plugin/cli

update:
	rm -rf vendor/${CHAIN33}
	git clone --depth 1 -b master https://${CHAIN33}.git vendor/${CHAIN33}
	rm -rf vendor/${CHAIN33}/.git
	cp -R vendor/${CHAIN33}/vendor/* vendor/
	rm -rf vendor/${CHAIN33}/vendor
	govendor init
	go build -i -o tool gitlab.33.cn/chain33/plugin/vendor/gitlab.33.cn/chain33/chain33/cmd/tools
	./tool import --path "plugin" --packname "gitlab.33.cn/chain33/plugin/plugin" --conf ""

updatevendor:
	govendor add +e
	govendor fetch -v +m

dep:
	dep init -v

clean:
	@rm -rf chain33
	@rm -rf chain33-cli
	@rm -rf tool
	@rm -rf plugin/init.go
	@rm -rf plugin/consensus/init
	@rm -rf plugin/dapp/init
	@rm -rf plugin/crypto/init
	@rm -rf plugin/store/init

test: ## Run unittests
	@go test -race $(PKG_LIST)

