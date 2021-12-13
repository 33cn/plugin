##
##编译solidity，并产生bin文件，abi文件，和相应的go文件

SRC_BEP := bep20
SRC_ERC20 := erc20
SRC_USDT := usdt
SRC_CONTRACT0 := contracts4chain33
SRC_CONTRACT1 := contracts4eth
SRC_MULTISIGN := gnosis/safe-contracts/contracts

GO_OUT0 := ${SRC_CONTRACT0}/generated
GO_OUT1 := ${SRC_CONTRACT1}/generated
GO_OUT_MULTISIGN := gnosis/generated
GO_OUT_BEP20 := bep20/generated
GO_OUT_ERC20 := erc20/generated
GO_OUT_USDT := usdt/generated
PACKAGE := generated

proj := "build"
.PHONY: default build clean registry bridgeBank setup

default: depends build

build: depends
	@abigen --sol $(SRC_CONTRACT0)/BridgeRegistry.sol --pkg $(PACKAGE) --out $(GO_OUT0)/BridgeRegistry.go
	@abigen --sol $(SRC_CONTRACT0)/BridgeBank/BridgeBank.sol --pkg $(PACKAGE) --out $(GO_OUT0)/BridgeBank.go
	@abigen --sol $(SRC_CONTRACT1)/BridgeRegistry.sol --pkg $(PACKAGE) --out $(GO_OUT1)/BridgeRegistry.go
	@abigen --sol $(SRC_CONTRACT1)/BridgeBank/BridgeBank.sol --pkg $(PACKAGE) --out $(GO_OUT1)/BridgeBank.go
	@abigen --sol $(SRC_BEP)/BEP20.sol --pkg $(PACKAGE) --out $(GO_OUT_BEP20)/bep20.go

multisign:
	@abigen --sol $(SRC_MULTISIGN)/GnosisSafe.sol --pkg $(PACKAGE) --out $(GO_OUT_MULTISIGN)/gnosisSafe.go

bep20Bin:
	@abigen --sol $(SRC_BEP)/BEP20.sol --pkg $(PACKAGE) --out $(GO_OUT_BEP20)/bep20.go

erc20Bin:
	@abigen --sol $(SRC_ERC20)/ERC20.sol --pkg $(PACKAGE) --out $(GO_OUT_ERC20)/erc20.go

usdtBin:
	@abigen --alias _totalSupply=_totalSupply4usdt --sol $(SRC_USDT)/USDT.sol --pkg $(PACKAGE) --out $(GO_OUT_USDT)/usdt.go


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
