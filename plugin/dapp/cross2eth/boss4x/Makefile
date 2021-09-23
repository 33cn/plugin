

proj := "build"
.PHONY: default build remote winset

SRC_CLI := 33cn/plugin/plugin/dapp/cross2eth/boss4x
OUT := build

default: build

build:
	@go build -o ${OUT}/boss4x

remote:
	@go build -v -o ${OUT}/boss4x_remote -ldflags "-X ${SRC_CLI}/buildFlags.RPCAddr4Chain33=http://183.129.226.74:8901"

winset:
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ${OUT}/boss4x.exe -ldflags "-X ${SRC_CLI}/buildFlags.RPCAddr4Chain33=http://183.129.226.74:8901"

clean:
	@rm ${OUT}/*
