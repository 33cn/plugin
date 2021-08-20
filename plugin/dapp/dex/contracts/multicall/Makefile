SRC_CONTRACT := src
GO_OUT := multicall
PACKAGE := multicall

proj := "build"
.PHONY: default multicall

default: multicall

multicall:
	@abigen --sol $(SRC_CONTRACT)/Multicall.sol --pkg $(PACKAGE) --out $(GO_OUT)/multicall.go
