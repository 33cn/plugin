// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"encoding/hex"

	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

// PropBoardTx 提案董事会成员RPC接口
func (c *Jrpc) PropBoardTx(parm *auty.ProposalBoard, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.propBoard(context.Background(), parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// RevokeProposalBoardTx  撤销提案董事会成员的RPC接口
func (c *Jrpc) RevokeProposalBoardTx(parm *auty.RevokeProposalBoard, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.revokeProposalBoard(context.Background(), parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// TerminateProposalBoardTx  终止提案董事会成员的RPC接口
func (c *Jrpc) TerminateProposalBoardTx(parm *auty.TerminateProposalBoard, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.terminateProposalBoard(context.Background(), parm)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}