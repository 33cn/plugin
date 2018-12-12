/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package types

import "fmt"

// some errors definition
var (
	ErrF3dStartRound     = fmt.Errorf("%s", "There's still one round left,you cann't start next round f3d!")
	ErrF3dManageAddr     = fmt.Errorf("%s", "You don't have permission to start f3d game.")
	ErrF3dManageBuyKey   = fmt.Errorf("%s", "You are manager,you don't have permission to buy key")
	ErrF3dBuyKey         = fmt.Errorf("%s", "the f3d is not start a new round!")
	ErrF3dBuyKeyTimeOut  = fmt.Errorf("%s", "The rest of the time is over,you can't buy any more keys!")
	ErrF3dDrawRound      = fmt.Errorf("%s", "There's not f3d round to draw!")
	ErrF3dDrawRemainTime = fmt.Errorf("%s", "There is time reamining,you can't draw the round game!")
	ErrF3dDrawRepeat     = fmt.Errorf("%s", "You can't repeat draw!")
)
