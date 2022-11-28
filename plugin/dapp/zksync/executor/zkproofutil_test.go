package executor

import (
	"testing"

	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/stretchr/testify/assert"
)

func TestParseRollbackOps(t *testing.T) {
	special := &zt.ZkDepositWitnessInfo{
		AccountID: 5,
		TokenID:   1,
		Amount:    "100",
		BlockInfo: &zt.OpBlockInfo{},
	}
	opDeposit1 := &zt.ZkOperation{
		Ty: zt.TyDepositAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}},
	}
	special = &zt.ZkDepositWitnessInfo{
		AccountID: 5,
		TokenID:   1,
		Amount:    "200",
		BlockInfo: &zt.OpBlockInfo{},
	}
	opDeposit2 := &zt.ZkOperation{
		Ty: zt.TyDepositAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}},
	}

	special3 := &zt.ZkDepositWitnessInfo{
		AccountID: 6,
		TokenID:   1,
		Amount:    "100",
		BlockInfo: &zt.OpBlockInfo{},
	}
	opDeposit3 := &zt.ZkOperation{
		Ty: zt.TyDepositAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special3}},
	}

	special2 := &zt.ZkWithdrawWitnessInfo{
		AccountID: 7,
		TokenID:   1,
		Amount:    "100",
		Fee: &zt.ZkFee{
			Fee: "10",
		},
		BlockInfo: &zt.OpBlockInfo{},
	}
	opWithdraw1 := &zt.ZkOperation{
		Ty: zt.TyWithdrawAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Withdraw{Withdraw: special2}},
	}

	ops := []*zt.ZkOperation{opDeposit1, opDeposit2, opDeposit3, opWithdraw1}
	depositAccts, withdAccts, depositMap, withdrawMap, err := parseRollbackOps(ops)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(depositAccts)) //包含sysFeeId
	assert.Equal(t, 1, len(withdAccts))
	assert.Equal(t, "10", depositMap[zt.SystemFeeAccountId].Tokens[0].Balance)
	assert.Equal(t, "300", depositMap[special.AccountID].Tokens[0].Balance)
	assert.Equal(t, special.TokenID, depositMap[special.AccountID].Tokens[0].TokenId)
	assert.Equal(t, "110", withdrawMap[special2.AccountID].Tokens[0].Balance)

}

//测试deposit withdraw merge
func TestParseRollbackOps2(t *testing.T) {
	special := &zt.ZkDepositWitnessInfo{
		AccountID: 5,
		TokenID:   1,
		Amount:    "100",
		BlockInfo: &zt.OpBlockInfo{},
	}
	opDeposit1 := &zt.ZkOperation{
		Ty: zt.TyDepositAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}},
	}
	special = &zt.ZkDepositWitnessInfo{
		AccountID: 5,
		TokenID:   1,
		Amount:    "200",
		BlockInfo: &zt.OpBlockInfo{},
	}
	opDeposit2 := &zt.ZkOperation{
		Ty: zt.TyDepositAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}},
	}

	special2 := &zt.ZkWithdrawWitnessInfo{
		AccountID: 5,
		TokenID:   1,
		Amount:    "100",
		Fee: &zt.ZkFee{
			Fee: "10",
		},
		BlockInfo: &zt.OpBlockInfo{},
	}
	opWithdraw1 := &zt.ZkOperation{
		Ty: zt.TyWithdrawAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Withdraw{Withdraw: special2}},
	}

	ops := []*zt.ZkOperation{opDeposit1, opDeposit2, opWithdraw1}
	depositAccts, withdAccts, depositMap, withdrawMap, err := parseRollbackOps(ops)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(depositAccts)) //包含sysFeeId
	assert.Equal(t, 1, len(withdAccts))
	assert.Equal(t, "10", depositMap[zt.SystemFeeAccountId].Tokens[0].Balance)
	assert.Equal(t, "0", withdrawMap[special2.AccountID].Tokens[0].Balance)
	assert.Equal(t, "190", depositMap[special.AccountID].Tokens[0].Balance)

}

// proxy exit
func TestParseRollbackOps3(t *testing.T) {
	special := &zt.ZkDepositWitnessInfo{
		AccountID: 5,
		TokenID:   1,
		Amount:    "100",
		BlockInfo: &zt.OpBlockInfo{},
	}
	opDeposit1 := &zt.ZkOperation{
		Ty: zt.TyDepositAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}},
	}
	special = &zt.ZkDepositWitnessInfo{
		AccountID: 5,
		TokenID:   1,
		Amount:    "200",
		BlockInfo: &zt.OpBlockInfo{},
	}
	opDeposit2 := &zt.ZkOperation{
		Ty: zt.TyDepositAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}},
	}

	special2 := &zt.ZkWithdrawWitnessInfo{
		AccountID: 7,
		TokenID:   1,
		Amount:    "100",
		Fee: &zt.ZkFee{
			Fee: "10",
		},
		BlockInfo: &zt.OpBlockInfo{},
	}
	opWithdraw1 := &zt.ZkOperation{
		Ty: zt.TyWithdrawAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Withdraw{Withdraw: special2}},
	}

	special3 := &zt.ZkProxyExitWitnessInfo{
		ProxyID:  8,
		TargetID: 7,
		TokenID:  1,
		Amount:   "100",
		Fee: &zt.ZkFee{
			Fee: "10",
		},
		BlockInfo: &zt.OpBlockInfo{},
	}
	opWithdraw2 := &zt.ZkOperation{
		Ty: zt.TyProxyExitAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_ProxyExit{ProxyExit: special3}},
	}

	ops := []*zt.ZkOperation{opDeposit1, opDeposit2, opWithdraw1, opWithdraw2}
	depositAccts, withdAccts, depositMap, withdrawMap, err := parseRollbackOps(ops)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(depositAccts)) //包含sysFeeId
	assert.Equal(t, 2, len(withdAccts))
	assert.Equal(t, "20", depositMap[zt.SystemFeeAccountId].Tokens[0].Balance)
	assert.Equal(t, "300", depositMap[special.AccountID].Tokens[0].Balance)
	assert.Equal(t, special.TokenID, depositMap[special.AccountID].Tokens[0].TokenId)
	assert.Equal(t, "210", withdrawMap[special2.AccountID].Tokens[0].Balance)
	assert.Equal(t, "10", withdrawMap[special3.ProxyID].Tokens[0].Balance)

}

func TestTransferPubDataToOps(t *testing.T) {
	//pubDatas :=[]string{
	//	"105312291815676758621043358086071746925458994755948066983944066622",
	//	"11551031525469018416798716462513755140273357930241311630414796971950",
	//	"13389833727841474876374951051773572305605675901678281473929158590464",
	//	"105312291840196687274897579819805299359863941693847892938875019668",
	//	"14749129358606502571982394423181528725152298876320481584554474268308",
	//	"4603405951260395322289355837836733611107199270909109925598967365632",
	//	"0",
	//	"0",
	//
	//}

	//pubDatas :=[]string{
	//	"315936875103751274720588951531580869730210795562584809349622792192",
	//	"0",
	//	"1158435208378648982342660591707908748612050541684879444608322371584",
	//	"947810625115094394896679136088222526798239328566498828731702837248",
	//	"0",
	//	"1158435208378648982330956533419261168913596741043317166254816493568",
	//	"0",
	//	"0",
	//
	//}
	pubDatas := []string{
		"315936875103751274709170969989933190681758961993115476335953707008",
		"0",
		"1158435208378648982330992685017922849415955089961610717309601579008",
		"526561458440865648133551979207405568953260984143369919068960359424",
		"31479141322252523513778259526959860474083891618159814059367596032",
		"1158435208378648982330989897424773033088062397996826636264413331456",
		"0",
		"0",
	}

	ops := transferPubDataToOps(pubDatas)
	for _, o := range ops {
		t.Log(o)
	}

}
