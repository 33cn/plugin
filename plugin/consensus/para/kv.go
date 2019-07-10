// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"fmt"

)



func calcTitleHeightKey(title string,height int64) []byte {
	return []byte(fmt.Sprintf("TH-%s-%d", title,height))
}

func calcTitleLastHeightKey(title string) []byte {
	return []byte(fmt.Sprintf("LH-%s", title))
}
