// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvmvccdb

import "errors"

var (
	//Err for StateHash lost
	ErrStateHashLost = errors.New("ErrStateHashLost")
)
