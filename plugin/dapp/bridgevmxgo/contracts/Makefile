##solc 使用版本为:Version: 0.5.16
##编译solidity，并产生bin文件，abi文件，和相应的go文件


SRC_CONTRACT0 := .
GO_OUT0 := ${SRC_CONTRACT0}/generated
PACKAGE := generated

proj := "build"
.PHONY: default build clean registry bridgeBank

default: depends build

build: depends
	@abigen --sol $(SRC_CONTRACT0)/BridgeRegistry.sol --pkg $(PACKAGE) --out $(GO_OUT0)/BridgeRegistry.go
	@abigen --sol $(SRC_CONTRACT0)/BridgeBank/BridgeBank.sol --pkg $(PACKAGE) --out $(GO_OUT0)/BridgeBank.go

clean:
	@rm -fr $(GO_OUT)/*

registry0:
	@abigen --sol $(SRC_CONTRACT0)/BridgeRegistry.sol --pkg $(PACKAGE) --out $(GO_OUT0)/BridgeRegistry.go

bridgeBank0:
	@abigen --sol $(SRC_CONTRACT0)/BridgeBank/BridgeBank.sol --pkg $(PACKAGE) --out $(GO_OUT0)/BridgeBank.go

asmExample:
	@solc --asm --bin -o  tmp/ valset-bin=./tmp/valset-bin Valset.sol

depends:
	if ! [ -d openzeppelin-solidity ]; then \
	echo "not exist openzeppelin-solidity and going to get" ; \
	go get github.com/OpenZeppelin/openzeppelin-contracts@v2.5 ; \
	mkdir openzeppelin-solidity ;\
	cp -r ${GOPATH}/pkg/mod/github.com/\!open\!zeppelin/openzeppelin-contracts@v2.5.0+incompatible/contracts openzeppelin-solidity ; \
	fi; \
