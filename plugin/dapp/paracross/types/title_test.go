// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "testing"

func TestNormalizeParaTitle(t *testing.T) {
	cases := []struct {
		title string
		want  string
	}{
		{"", ""},
		{"user.p.superx.", "user.p.superx."},
		{"user.p.superx", "user.p.superx."},
		{"user.p.", "user.p."},
	}
	for _, c := range cases {
		if got := NormalizeParaTitle(c.title); got != c.want {
			t.Errorf("NormalizeParaTitle(%s)=%s,want %s", c.title, got, c.want)
		}
	}
}

func TestIsValidParaTitle(t *testing.T) {
	cases := []struct {
		title string
		want  bool
	}{
		{"user.p.test.", true},
		{"user.p.superx.", true},
		{"user.p.a1.", true},
		//攻击中使用的伪造前缀title，缺少结尾'.'
		{"user.p.super", false},
		{"user.p.superx", false},
		//空名称，会匹配所有平行链
		{"user.p.", false},
		{"u", false},
		{"", false},
		//大小写变体
		{"USER.P.I8847A", false},
		{"user.p.SuperX.", true}, //名称本身合法，但与真实链不同(大小写敏感)
		//多级名称，与chain33 GetParaExecTitleName的解析不一致
		{"user.p.a.b.", false},
		{"user.p..", false},
	}
	for _, c := range cases {
		if got := IsValidParaTitle(c.title); got != c.want {
			t.Errorf("IsValidParaTitle(%s)=%v,want %v", c.title, got, c.want)
		}
	}
}

func TestIsExactParaTitleExec(t *testing.T) {
	cases := []struct {
		title string
		exec  string
		want  bool
	}{
		{"user.p.superx.", "user.p.superx.paracross", true},
		//非规范title不做归一化: user.p.superx 与真实链 user.p.superx. 是两个不同的状态key，
		//若容忍归一化匹配，攻击者可用"真实链title去掉结尾'.'"建立自己的节点组并执行真实链跨链交易
		{"user.p.superx", "user.p.superx.paracross", false},
		{"user.p.superx", "user.p.super.paracross", false},
		{"user.p.superx.", "", false},
		{"user.p.superx.", "user.p.paracross", false},
		//伪造前缀title不能匹配真实链的execer
		{"user.p.super", "user.p.superx.paracross", false},
		{"user.p.super.", "user.p.superx.paracross", false},
		{"user.p.", "user.p.superx.paracross", false},
		{"u", "user.p.superx.paracross", false},
		{"user.p.superx.", "user.p.super.paracross", false},
		{"user.p.superx.", "user.p.superx2.paracross", false},
		{"user.p.superx.", "user.p.superx.coins", false},
		{"user.p.superx.", "paracross", false},
		{"user.p.SuperX.", "user.p.superx.paracross", false},
	}
	for _, c := range cases {
		if got := IsExactParaTitleExec(c.title, c.exec); got != c.want {
			t.Errorf("IsExactParaTitleExec(%s,%s)=%v,want %v", c.title, c.exec, got, c.want)
		}
	}
}
