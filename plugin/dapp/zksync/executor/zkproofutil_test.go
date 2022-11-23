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
	}
	opDeposit1 := &zt.ZkOperation{
		Ty: zt.TyDepositAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}},
	}
	special = &zt.ZkDepositWitnessInfo{
		AccountID: 5,
		TokenID:   1,
		Amount:    "200",
	}
	opDeposit2 := &zt.ZkOperation{
		Ty: zt.TyDepositAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}},
	}

	special3 := &zt.ZkDepositWitnessInfo{
		AccountID: 6,
		TokenID:   1,
		Amount:    "100",
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
	}
	opDeposit1 := &zt.ZkOperation{
		Ty: zt.TyDepositAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}},
	}
	special = &zt.ZkDepositWitnessInfo{
		AccountID: 5,
		TokenID:   1,
		Amount:    "200",
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
	}
	opDeposit1 := &zt.ZkOperation{
		Ty: zt.TyDepositAction,
		Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}},
	}
	special = &zt.ZkDepositWitnessInfo{
		AccountID: 5,
		TokenID:   1,
		Amount:    "200",
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
