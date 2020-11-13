// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/stretchr/testify/assert"
)

func TestVerifyCommitValues(t *testing.T) {
	input1 := &mixTy.TransferInputPublicInput{
		CommitValueX: "8728367628344135467582547753719073727968275979035063555332785894244029982715",
		CommitValueY: "8834462946188529904793384347374734779374831553974460136522409595751449858199",
	}

	input2 := &mixTy.TransferInputPublicInput{
		CommitValueX: "9560056125663567360314373555170485462871740364163814576088225107862234393497",
		CommitValueY: "13024071698463677601393829581435828705327146000694268918451707151508990195684",
	}

	var inputs []*mixTy.TransferInputPublicInput
	inputs = append(inputs, input1)
	inputs = append(inputs, input2)

	output1 := &mixTy.TransferOutputPublicInput{
		CommitValueX: "8728367628344135467582547753719073727968275979035063555332785894244029982715",
		CommitValueY: "8834462946188529904793384347374734779374831553974460136522409595751449858199",
	}
	var outputs []*mixTy.TransferOutputPublicInput
	outputs = append(outputs, output1)

	ret := verifyCommitValues(inputs, outputs)
	assert.Equal(t, true, ret)

}
