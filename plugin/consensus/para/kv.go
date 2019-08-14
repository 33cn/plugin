// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"fmt"

	"github.com/33cn/chain33/types"
)

func calcTitleHeightKey(title string, height int64) []byte {
	return []byte(fmt.Sprintf("%s-TH-%s-%d", types.ConsensusParaTxsPrefix, title, height))
}

//temp main height tx data for big size
func calcTitleMainHeightKey(title string, height int64) []byte {
	return []byte(fmt.Sprintf("%s-TMH-%s-%d", types.ConsensusParaTxsPrefix, title, height))
}

func calcTitleLastHeightKey(title string) []byte {
	return []byte(fmt.Sprintf("%s-TLH-%s", types.ConsensusParaTxsPrefix, title))
}

func calcTitleFirstHeightKey(title string) []byte {
	return []byte(fmt.Sprintf("%s-TFH-%s", types.ConsensusParaTxsPrefix, title))
}
