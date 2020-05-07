#!/bin/sh
protoc --go_out=../../types ./*.proto --proto_path=.
