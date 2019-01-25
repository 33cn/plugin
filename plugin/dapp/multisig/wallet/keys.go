// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import "fmt"

const (
	// MultisigAddr 记录本钱包owner地址拥有的多重签名地址，key:"multisig-addr-owneraddr, value [](multisigaddr,owneraddr,weight)
	MultisigAddr = "multisig-addr-"
)

func calcMultisigAddr(ownerAddr string) []byte {
	return []byte(fmt.Sprintf("%s%s", MultisigAddr, ownerAddr))
}

func calcPrefixMultisigAddr() []byte {
	return []byte(MultisigAddr)
}
