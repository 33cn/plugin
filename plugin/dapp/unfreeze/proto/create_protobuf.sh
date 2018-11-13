// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#!/bin/sh
protoc --go_out=plugins=grpc:../types ./*.proto --proto_path=. --proto_path="$GOPATH/src/gitlab.33.cn/chain33/chain33/types/proto/"
