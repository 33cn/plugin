##
##编译solidity，并产生bin文件，abi文件，和相应的go文件

SRC_CONTRACT := contracts
GO_OUT0 := src/cakeToken/
GO_OUT1 := src/syrupBar/
GO_OUT2 := src/sousChef/
GO_OUT3 := src/masterChef/

PACKAGE0 := cakeToken
PACKAGE1 := syrupBar
PACKAGE2 := sousChef
PACKAGE3 := masterChef

proj := "build"
.PHONY: default build clean

default: sol build

sol:
	@abigen --sol $(SRC_CONTRACT)/CakeToken.sol --pkg $(PACKAGE0) --out $(GO_OUT0)/cakeToken.go
	@abigen --sol $(SRC_CONTRACT)/SyrupBar.sol --pkg $(PACKAGE1) --out $(GO_OUT1)/syrupBar.go
	@abigen --sol $(SRC_CONTRACT)/SousChef.sol --pkg $(PACKAGE2) --out $(GO_OUT2)/sousChef.go
	@abigen --sol $(SRC_CONTRACT)/MasterChef.sol --pkg $(PACKAGE3) --out $(GO_OUT3)/masterChef.go

build:
	cd bsctest && go build && cd ..


clean:
	@rm -fr $(GO_OUT)/*.go
