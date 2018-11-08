all: vendor build

build:
	go build -i -o tool gitlab.33.cn/chain33/chain33/cmd/tools
	./tool import --path "plugin" --packname "gitlab.33.cn/chain33/plugin/plugin" --conf ""
	go build -i -o chain33
	go build -i -o chain33-cli gitlab.33.cn/chain33/plugin/cli
updatevendor:
	govendor fetch +m
	govendor add +e

vendor:
	go get -v -u gitlab.33.cn/chain33/chain33
	govendor init
	govendor add +e
	govendor fetch +m
clean:
	@rm -rf chain33
	@rm -rf chain33-cli
	@rm -rf tool
	@rm -rf plugin/init.go
	@rm -rf plugin/consensus/init
	@rm -rf plugin/dapp/init
	@rm -rf plugin/crypto/init
	@rm -rf plugin/store/init
