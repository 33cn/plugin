#!/bin/bash
# proto生成命令，将pb.go文件生成到types/目录下, chain33_path支持引用chain33框架的proto文件
chain33_path=$(go list -f '{{.Dir}}' "github.com/33cn/chain33")
protoc --go_out=plugins=grpc:../types ./*.proto --proto_path=. --proto_path="${chain33_path}/types/proto/"
