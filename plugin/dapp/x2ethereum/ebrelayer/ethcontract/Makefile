##
##编译solidity，并产生bin文件，abi文件，和相应的go文件


SRC_CONTRACT := sol_contracts
#SRC_CONTRACT := github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/ethcontract/sol_contracts
GO_OUT := generated
PACKAGE := generated

proj := "build"
.PHONY: default build clean registry bridgeBank setup

default: depends build

build: depends
	@abigen --sol $(SRC_CONTRACT)/BridgeRegistry.sol --pkg $(PACKAGE) --out $(GO_OUT)/BridgeRegistry.go
	@abigen --sol $(SRC_CONTRACT)/BridgeBank/BridgeBank.sol --pkg $(PACKAGE) --out $(GO_OUT)/BridgeBank.go
#@abigen --sol $(SRC_CONTRACT)/CosmosBridge.sol --pkg $(PACKAGE) --out $(GO_OUT)/CosmosBridge.go
#@abigen --sol $(SRC_CONTRACT)/Oracle.sol --pkg $(PACKAGE) --out $(GO_OUT)/Oracle.go
#@abigen --sol $(SRC_CONTRACT)/Valset.sol --pkg $(PACKAGE) --out $(GO_OUT)/Valset.go
clean:
	@rm -fr $(GO_OUT)/*

registry:
	@abigen --sol $(SRC_CONTRACT)/BridgeRegistry.sol --pkg $(PACKAGE) --out $(GO_OUT)/BridgeRegistry.go

bridgeBank:
	@abigen --sol $(SRC_CONTRACT)/BridgeBank/BridgeBank.sol --pkg $(PACKAGE) --out $(GO_OUT)/BridgeBank.go

depends:
	if ! [ -d openzeppelin-solidity ]; then \
	echo "not exist openzeppelin-solidity and going to get" ; \
	go get github.com/OpenZeppelin/openzeppelin-contracts@v2.5 ; \
	mkdir openzeppelin-solidity ;\
	cp -r ${GOPATH}/pkg/mod/github.com/\!open\!zeppelin/openzeppelin-contracts@v2.5.0+incompatible/contracts openzeppelin-solidity ; \
	fi; \
