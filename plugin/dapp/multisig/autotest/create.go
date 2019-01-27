// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"encoding/json"

	"github.com/33cn/chain33/cmd/autotest/types"
)

type createMultisigCase struct {
	types.BaseCase
	Creator string   `toml:"creator"`
	Owners  []string `toml:"owners"`
}

type createMultisigPack struct {
	types.BaseCasePack
	info *multisigInfo
}

type multisigInfo struct {
	account string
}

// SendCommand send command
func (t *createMultisigCase) SendCommand(packID string) (types.PackFunc, error) {

	return types.DefaultSend(t, &createMultisigPack{}, packID)
}

// GetCheckHandlerMap defines get check handle for map
func (pack *createMultisigPack) GetCheckHandlerMap() interface{} {

	funcMap := make(types.CheckHandlerMap, 1)
	funcMap["create"] = pack.checkCreate

	return funcMap
}

// GetDependData defines get depend data function
func (pack *createMultisigPack) GetDependData() interface{} {

	return pack.info
}

func (pack *createMultisigPack) checkCreate(txInfo types.CheckHandlerParamType) bool {

	var createLog map[string]interface{}

	err := json.Unmarshal(txInfo.Receipt.Logs[1].Log, &createLog)

	if err != nil {
		pack.FLog.Error("checkMultisigCreate", "id", pack.PackID, "unmarshalErr", err)
		return false
	}
	interCase := pack.TCase.(*createMultisigCase)
	info := &multisigInfo{}
	info.account = createLog["multiSigAddr"].(string)
	pack.info = info

	creator := createLog["createAddr"].(string)
	owners := createLog["owners"].([]interface{})

	if creator != interCase.Creator {
		pack.FLog.Error("WrongMultiSignCreator", "id", pack.PackID, "creator", creator, "expect", interCase.Creator)
		return false
	}

	for i, owner := range owners {
		addr := owner.(map[string]interface{})["ownerAddr"].(string)

		if addr != interCase.Owners[i] {
			pack.FLog.Error("WrongMultiSignOwner", "id", pack.PackID, "owner", addr, "expect", interCase.Owners[i])
			return false
		}
	}

	return true
}
