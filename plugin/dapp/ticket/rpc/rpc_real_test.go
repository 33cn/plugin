// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

//only load all plugin and system
import (
	"testing"

	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/util/testnode"
	_ "github.com/33cn/plugin/plugin"
)

func TestNewTicket(t *testing.T) {
	cfg, sub := testnode.GetDefaultConfig()
	cfg.Consensus.Name = "ticket"
	mock33 := testnode.NewWithConfig(cfg, sub, nil)
	defer mock33.Close()
	mock33.WaitHeight(5)
}
