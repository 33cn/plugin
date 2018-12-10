/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */
package types

var (
	// 本游戏合约管理员地址
	f3dManagerAddr = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"

	//本游戏合约平台开发者分成地址
	f3dDeveloperAddr = "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"

	// 超级大奖分成百分比
	f3dBonusWinner = float32(0.4)

	// 参与者分红百分比
	f3dBonusKey = float32(0.3)

	// 滚动到下期奖金池百分比
	f3dBonusPool = float32(0.2)

	// 平台运营及开发者费用百分比
	f3dBonusDeveloper = float32(0.1)

	// 本游戏一轮运行的最长周期（单位：秒）
	f3dTimeLife = int64(3600)

	// 一把钥匙延长的游戏时间（单位：秒）
	f3dTimeKey = int64(30)

	// 一次购买钥匙最多延长的游戏时间（单位：秒）
	f3dTimeMaxkey = int64(300)

	// 钥匙涨价幅度（下一个人购买钥匙时在上一把钥匙基础上浮动幅度百分比），范围1-100
	f3dKeyPriceIncr = float32(0.1)

	// start Key price  o.1 token
	f3dKeyPriceStart = float32(0.1)
)

func SetConfig() {

}

func GetF3dManagerAddr() string {
	return f3dManagerAddr
}

func GetF3dDeveloperAddr() string {
	return f3dDeveloperAddr
}

func GetF3dBonusWinner() float32 {
	return f3dBonusWinner
}

func GetF3dBonusKey() float32 {
	return f3dBonusKey
}

func GetF3dBonusPool() float32 {
	return f3dBonusPool
}

func GetF3dBonusDeveloper() float32 {
	return f3dBonusDeveloper
}

func GetF3dTimeLife() int64 {
	return f3dTimeLife
}

func GetF3dTimeKey() int64 {
	return f3dTimeKey
}

func GetF3dTimeMaxkey() int64 {
	return f3dTimeMaxkey
}

func GetF3dKeyPriceIncr() float32 {
	return f3dKeyPriceIncr
}

func GetF3dKeyPriceStart() float32 {
	return f3dKeyPriceStart
}
