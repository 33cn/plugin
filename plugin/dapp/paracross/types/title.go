// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"strings"

	"github.com/33cn/chain33/types"
)

// 平行链title的规范形式是 user.p.<name>.，<name>非空且不含分隔符'.'，
// 与chain33中Chain33Config.isPara()以及GetParaExecTitleName的判断保持一致。

// NormalizeParaTitle 归一化平行链title，补齐结尾的'.'
// 例如 user.p.superx -> user.p.superx.，user.p.superx. 保持不变
func NormalizeParaTitle(title string) string {
	if title == "" || strings.HasSuffix(title, ".") {
		return title
	}
	return title + "."
}

// IsValidParaTitle 判断title是否为规范的平行链title，形如user.p.<name>.
// 注意平行链执行器名为 user.p.<name>.paracross，因此title必须以'.'结尾，
// 否则 user.p.super 会成为真实平行链 user.p.superx. 的前缀，进而前缀匹配到其跨链交易
func IsValidParaTitle(title string) bool {
	if !strings.HasPrefix(title, types.ParaKeyX) || !strings.HasSuffix(title, ".") {
		return false
	}
	//user.p. 这类没有名称的title会导致下面的切片越界，必须先做长度判断
	if len(title) <= len(types.ParaKeyX) {
		return false
	}
	name := title[len(types.ParaKeyX) : len(title)-1]
	return len(name) > 0 && !strings.Contains(name, ".")
}

// IsExactParaTitleExec 判断title是否精确对应execer，即exec == title+"paracross"。
// 这里**不做归一化**：user.p.superx 与 user.p.superx. 是节点组/commit中两个不同的状态key，
// 若容忍非规范title归一化后匹配，攻击者就可以用"真实链title去掉结尾'.'"的形式(如user.p.superx)
// 建立自己的节点组，再从rollup等执行入口拿走真实平行链(user.p.superx.)的跨链交易。
// title必须为规范形式(见IsValidParaTitle)，调用方需要归一化时请显式调用NormalizeParaTitle后再校验。
func IsExactParaTitleExec(title, exec string) bool {
	return IsValidParaTitle(title) && exec == title+ParaX
}
