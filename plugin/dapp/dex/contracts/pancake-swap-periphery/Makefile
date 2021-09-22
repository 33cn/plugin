##
##编译solidity，并产生bin文件，abi文件，和相应的go文件

SRC_CONTRACT := contracts
GO_OUT := src/pancakeRouter
PACKAGE := pancakeRouter

proj := "build"
.PHONY: default router clean factory

default: router

router:
	@abigen --sol $(SRC_CONTRACT)/PancakeRouter.sol --pkg $(PACKAGE) --out $(GO_OUT)/PancakeRouter.go

weth:
	@abigen --sol $(SRC_CONTRACT)/test/WETH9.sol --pkg $(PACKAGE) --out $(GO_OUT)/weth.go

factory:
	@abigen --sol pancake-swap-core/$(SRC_CONTRACT)/PancakeFactory.sol --pkg pancakeFactory --out src/pancakeFactory/pancakeFactory.go


clean:
	@rm -fr $(GO_OUT)/PancakeRouter.go
