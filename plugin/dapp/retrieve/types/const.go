// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

//retrieve
const (
	RetrieveBackup = iota + 1
	RetrievePreapre
	RetrievePerform
	RetrieveCancel
)

// retrieve op
const (
	RetrieveActionPrepare = 1
	RetrieveActionPerform = 2
	RetrieveActionBackup  = 3
	RetrieveActionCancel  = 4
)

// retrieve names
var (
	JRPCName  = "Retrieve"
	RetrieveX = "retrieve"

	ExecerRetrieve = []byte(RetrieveX)

	actionName = map[string]int32{
		"Prepare": RetrieveActionPrepare,
		"Perform": RetrieveActionPerform,
		"Backup":  RetrieveActionBackup,
		"Cancel":  RetrieveActionCancel,
	}

	ForkRetriveAssetX = "ForkRetriveAsset"
	ForkRetriveX      = "ForkRetrive"
	// ForkRetrivePerformRelayX 修复 ForkRetriveAsset 分支下 perform 成功后未清除找回关系导致的 perform 重放漏洞
	ForkRetrivePerformRelayX = "ForkRetrivePerformRelay"
)
