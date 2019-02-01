#!/bin/sh
# win系统下
protoc --go_out=plugins=grpc:..\\types ./*.proto --proto_path=. --proto_path="../../../../vendor/github.com/33cn/chain33/types/proto/"
# ubuntu系统下
# protoc --go_out=plugins=grpc:../types ./*.proto --proto_path=. --proto_path="../../../../vendor/github.com/33cn/chain33/types/proto/"
