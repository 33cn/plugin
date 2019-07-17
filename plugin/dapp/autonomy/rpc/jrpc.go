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

// 提案董事会相关
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

// VoteProposalBoardTx  投票提案董事会成员的RPC接口
func (c *Jrpc) VoteProposalBoardTx(parm *auty.VoteProposalBoard, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.voteProposalBoard(context.Background(), parm)
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

// 提案项目相关
// PropProjectTx 提案项目RPC接口
func (c *Jrpc) PropProjectTx(parm *auty.ProposalProject, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.propProject(context.Background(), parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// RevokeProposalProjectTx  撤销提案项目的RPC接口
func (c *Jrpc) RevokeProposalProjectTx(parm *auty.RevokeProposalProject, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.revokeProposalProject(context.Background(), parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// VoteProposalProjectTx  董事会投票提案项目的RPC接口
func (c *Jrpc) VoteProposalProjectTx(parm *auty.VoteProposalProject, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.voteProposalProject(context.Background(), parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// PubVoteProposalProjectTx  全体持票人投票提案项目的RPC接口
func (c *Jrpc) PubVoteProposalProjectTx(parm *auty.PubVoteProposalProject, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.pubVoteProposalProject(context.Background(), parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// TerminateProposalProjectTx  终止提案项目的RPC接口
func (c *Jrpc) TerminateProposalProjectTx(parm *auty.TerminateProposalProject, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.terminateProposalProject(context.Background(), parm)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}

// 提案规则相关
// PropRuleTx 提案规则RPC接口
func (c *Jrpc) PropRuleTx(parm *auty.ProposalRule, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.propRule(context.Background(), parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// RevokeProposalRuleTx  撤销提案规则RPC接口
func (c *Jrpc) RevokeProposalRuleTx(parm *auty.RevokeProposalRule, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.revokeProposalRule(context.Background(), parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// VoteProposalRuleTx  投票提案规则RPC接口
func (c *Jrpc) VoteProposalRuleTx(parm *auty.VoteProposalRule, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.voteProposalRule(context.Background(), parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// TerminateProposalRuleTx  终止提案规则RPC接口
func (c *Jrpc) TerminateProposalRuleTx(parm *auty.TerminateProposalRule, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.terminateProposalRule(context.Background(), parm)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}