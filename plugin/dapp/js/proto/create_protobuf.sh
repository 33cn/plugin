#!/bin/sh
protoc --go_out=plugins=grpc:../types/jsproto/ ./*.proto
