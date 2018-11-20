// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"encoding/hex"

	"github.com/33cn/plugin/plugin/dapp/retrieve/types"
)

//CreateRawRetrieveBackupTx 构造backup
func (c *Jrpc) CreateRawRetrieveBackupTx(in *RetrieveBackupTx, result *interface{}) error {
	head := &types.BackupRetrieve{
		BackupAddress:  in.BackupAddr,
		DefaultAddress: in.DefaultAddr,
		DelayPeriod:    in.DelayPeriod,
	}

	reply, err := c.cli.Backup(context.Background(), head)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}

//CreateRawRetrievePrepareTx 构造prepare
func (c *Jrpc) CreateRawRetrievePrepareTx(in *RetrievePrepareTx, result *interface{}) error {
	head := &types.PrepareRetrieve{
		BackupAddress:  in.BackupAddr,
		DefaultAddress: in.DefaultAddr,
	}

	reply, err := c.cli.Prepare(context.Background(), head)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}

//CreateRawRetrievePerformTx 构造perform
func (c *Jrpc) CreateRawRetrievePerformTx(in *RetrievePerformTx, result *interface{}) error {
	head := &types.PerformRetrieve{
		BackupAddress:  in.BackupAddr,
		DefaultAddress: in.DefaultAddr,
	}
	reply, err := c.cli.Perform(context.Background(), head)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}

//CreateRawRetrieveCancelTx 构造cancel
func (c *Jrpc) CreateRawRetrieveCancelTx(in *RetrieveCancelTx, result *interface{}) error {
	head := &types.CancelRetrieve{
		BackupAddress:  in.BackupAddr,
		DefaultAddress: in.DefaultAddr,
	}

	reply, err := c.cli.Cancel(context.Background(), head)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}
