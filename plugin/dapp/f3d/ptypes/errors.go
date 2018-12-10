/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package types
import "fmt"

// some errors definition
var (
	ErrF3dStartRound = fmt.Errorf("%s", "There's still one round left,you cann't start next round f3d!")
	ErrF3dManageAddr   = fmt.Errorf("%s", "You don't have permission to start f3d game.")
	ErrGameCloseAddr    = fmt.Errorf("%s", "The game time has not yet expired,You don't have permission to call yet.")
	ErrGameTimeOut      = fmt.Errorf("%s", "The game has expired.,You don't have permission to call.")
	ErrGameMatchStatus  = fmt.Errorf("%s", "can't join the game, the game has matched or finished!")
	ErrGameMatch        = fmt.Errorf("%s", "can't join the game, You can't match the game you created!")
	ErrGameCancleStatus = fmt.Errorf("%s", "can't cancle the game, the game has matched!")
	ErrGameCloseStatus  = fmt.Errorf("%s", "can't close the game again, the game has  finished!")
)