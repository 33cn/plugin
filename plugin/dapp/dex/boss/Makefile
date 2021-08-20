proj := "build"
.PHONY: default build remote winset

SRC_CLI := 33cn/plugin/plugin/dapp/dex/boss
OUT := build

default: build

build:
	@go build -o ${OUT}/boss -ldflags "-X ${SRC_CLI}/buildFlags.RPCAddr4Chain33=http://localhost:8801 -X ${SRC_CLI}/buildFlags.RPCAddr4Ethereum=wss://ws-testnet.huobichain.com"

remote:
	@go build -v -o ${OUT}/remote -ldflags "-X ${SRC_CLI}/buildFlags.RPCAddr4Chain33=http://183.129.226.74:8901 -X ${SRC_CLI}/buildFlags.RPCAddr4Ethereum=wss://ws-testnet.huobichain.com"

winset:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ${OUT}/bosswin.exe -ldflags "-X ${SRC_CLI}/buildFlags.RPCAddr4Chain33=http://183.129.226.74:8901 -X ${SRC_CLI}/buildFlags.RPCAddr4Ethereum=https://http-testnet.huobichain.com"

clean:
	@rm ${OUT}/*

