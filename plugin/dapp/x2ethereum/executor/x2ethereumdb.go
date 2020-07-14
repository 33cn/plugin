package executor

import (
	"strconv"
	"strings"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	x2eTy "github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
)

//Oracle ...
type Oracle struct {
	db                 dbm.KV
	consensusThreshold int64
}

//NewOracle ...
func NewOracle(db dbm.KV, consensusThreshold int64) *Oracle {
	if consensusThreshold <= 0 || consensusThreshold > 100 {
		return nil
	}
	return &Oracle{
		consensusThreshold: consensusThreshold,
		db:                 db,
	}
}

//ProcessSuccessfulClaimForLock 处理经过审核的关于Lock的claim
func (o *Oracle) ProcessSuccessfulClaimForLock(claim, execAddr string, accDB *account.DB) (*types.Receipt, error) {
	var receipt *types.Receipt
	oracleClaim, err := CreateOracleClaimFromOracleString(claim)
	if err != nil {
		elog.Error("CreateEthClaimFromOracleString", "CreateOracleClaimFromOracleString error", err)
		return nil, err
	}

	receiverAddress := oracleClaim.Chain33Receiver

	if oracleClaim.ClaimType == int64(x2eTy.LockClaimType) {
		//铸币到相关的tokenSymbolBank账户下
		amount, _ := strconv.ParseInt(x2eTy.TrimZeroAndDot(oracleClaim.Amount), 10, 64)

		receipt, err = accDB.Mint(receiverAddress, amount)
		if err != nil {
			return nil, err
		}

		return receipt, nil
	}
	return nil, x2eTy.ErrInvalidClaimType
}

//ProcessSuccessfulClaimForBurn 处理经过审核的关于Burn的claim
func (o *Oracle) ProcessSuccessfulClaimForBurn(claim, execAddr, tokenSymbol string, accDB *account.DB) (*types.Receipt, error) {
	oracleClaim, err := CreateOracleClaimFromOracleString(claim)
	if err != nil {
		elog.Error("CreateEthClaimFromOracleString", "CreateOracleClaimFromOracleString error", err)
		return nil, err
	}

	senderAddr := oracleClaim.Chain33Receiver

	if oracleClaim.ClaimType == int64(x2eTy.BurnClaimType) {
		amount, _ := strconv.ParseInt(x2eTy.TrimZeroAndDot(oracleClaim.Amount), 10, 64)
		receipt, err := accDB.ExecTransfer(address.ExecAddress(tokenSymbol), senderAddr, execAddr, amount)
		if err != nil {
			return nil, err
		}
		return receipt, nil
	}
	return nil, x2eTy.ErrInvalidClaimType
}

// ProcessBurn processes the burn of bridged coins from the given sender
func (o *Oracle) ProcessBurn(address, amount string, accDB *account.DB) (*types.Receipt, error) {
	var a int64
	a, _ = strconv.ParseInt(x2eTy.TrimZeroAndDot(amount), 10, 64)

	receipt, err := accDB.Burn(address, a)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

// ProcessLock processes the lockup of cosmos coins from the given sender
// accDB = mavl-coins-bty-addr
func (o *Oracle) ProcessLock(address, to, execAddr, amount string, accDB *account.DB) (*types.Receipt, error) {
	// 转到 mavl-coins-bty-execAddr:addr
	a, _ := strconv.ParseInt(x2eTy.TrimZeroAndDot(amount), 10, 64)
	receipt, err := accDB.ExecTransfer(address, to, execAddr, a)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

//ProcessAddValidator ...
// 对于相同的地址该如何处理?
// 现有方案是相同地址就报错
func (o *Oracle) ProcessAddValidator(address string, power int64) (*types.Receipt, error) {
	receipt := new(types.Receipt)

	validatorMaps, err := o.GetValidatorArray()
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}

	if validatorMaps == nil {
		validatorMaps = new(x2eTy.ValidatorList)
	}

	elog.Info("ProcessLogInValidator", "pre validatorMaps", validatorMaps, "Add Address", address, "Add power", power)
	var totalPower int64
	for _, p := range validatorMaps.Validators {
		if p.Address != address {
			totalPower += p.Power
		} else {
			return nil, x2eTy.ErrAddressExists
		}
	}

	vs := append(validatorMaps.Validators, &x2eTy.MsgValidator{
		Address: address,
		Power:   power,
	})

	validatorMaps.Validators = vs

	v := types.Encode(validatorMaps)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: x2eTy.CalValidatorMapsPrefix(), Value: v})
	totalPower += power

	totalP := x2eTy.ReceiptQueryTotalPower{
		TotalPower: totalPower,
	}
	totalPBytes := types.Encode(&totalP)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: x2eTy.CalLastTotalPowerPrefix(), Value: totalPBytes})
	return receipt, nil
}

//ProcessRemoveValidator ...
func (o *Oracle) ProcessRemoveValidator(address string) (*types.Receipt, error) {
	var exist bool
	receipt := new(types.Receipt)

	validatorMaps, err := o.GetValidatorArray()
	if err != nil {
		return nil, err
	}

	elog.Info("ProcessLogOutValidator", "pre validatorMaps", validatorMaps, "Delete Address", address)
	var totalPower int64
	validatorRes := new(x2eTy.ValidatorList)
	for _, p := range validatorMaps.Validators {
		if address != p.Address {
			v := append(validatorRes.Validators, p)
			validatorRes.Validators = v
			totalPower += p.Power
		} else {
			exist = true
			continue
		}
	}

	if !exist {
		return nil, x2eTy.ErrAddressNotExist
	}

	v := types.Encode(validatorRes)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: x2eTy.CalValidatorMapsPrefix(), Value: v})
	totalP := x2eTy.ReceiptQueryTotalPower{
		TotalPower: totalPower,
	}
	totalPBytes := types.Encode(&totalP)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: x2eTy.CalLastTotalPowerPrefix(), Value: totalPBytes})
	return receipt, nil
}

//ProcessModifyValidator 这里的power指的是修改后的power
func (o *Oracle) ProcessModifyValidator(address string, power int64) (*types.Receipt, error) {
	var exist bool
	receipt := new(types.Receipt)

	validatorMaps, err := o.GetValidatorArray()
	if err != nil {
		return nil, err
	}

	elog.Info("ProcessModifyValidator", "pre validatorMaps", validatorMaps, "Modify Address", address, "Modify power", power)
	var totalPower int64
	for index, p := range validatorMaps.Validators {
		if address != p.Address {
			totalPower += p.Power
		} else {
			validatorMaps.Validators[index].Power = power
			exist = true
			totalPower += power
		}
	}

	if !exist {
		return nil, x2eTy.ErrAddressNotExist
	}

	v := types.Encode(validatorMaps)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: x2eTy.CalValidatorMapsPrefix(), Value: v})
	totalP := x2eTy.ReceiptQueryTotalPower{
		TotalPower: totalPower,
	}
	totalPBytes := types.Encode(&totalP)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: x2eTy.CalLastTotalPowerPrefix(), Value: totalPBytes})

	return receipt, nil
}

//ProcessSetConsensusNeeded ...
func (o *Oracle) ProcessSetConsensusNeeded(ConsensusThreshold int64) (int64, int64, error) {
	preCon := o.GetConsensusThreshold()
	o.SetConsensusThreshold(ConsensusThreshold)
	nowCon := o.GetConsensusThreshold()

	elog.Info("ProcessSetConsensusNeeded", "pre ConsensusThreshold", preCon, "now ConsensusThreshold", nowCon)

	return preCon, nowCon, nil
}

//GetProphecy ...
func (o *Oracle) GetProphecy(id string) (*x2eTy.ReceiptEthProphecy, error) {
	if id == "" {
		return nil, x2eTy.ErrInvalidIdentifier
	}

	bz, err := o.db.Get(x2eTy.CalProphecyPrefix(id))
	if err != nil && err != types.ErrNotFound {
		return nil, x2eTy.ErrProphecyGet
	} else if err == types.ErrNotFound {
		return nil, x2eTy.ErrProphecyNotFound
	}

	var dbProphecy x2eTy.ReceiptEthProphecy
	err = types.Decode(bz, &dbProphecy)
	if err != nil {
		return nil, types.ErrUnmarshal
	}
	return &dbProphecy, nil
}

// setProphecy saves a prophecy with an initial claim
func (o *Oracle) setProphecy(prophecy *x2eTy.ReceiptEthProphecy) error {
	err := o.checkProphecy(prophecy)
	if err != nil {
		return err
	}

	bz, err := o.db.Get(x2eTy.CalProphecyPrefix(prophecy.ID))
	if err != nil && err != types.ErrNotFound {
		return x2eTy.ErrProphecyGet
	}

	var dbProphecy x2eTy.ReceiptEthProphecy
	if err != types.ErrNotFound {
		err = types.Decode(bz, &dbProphecy)
		if err != nil {
			return types.ErrUnmarshal
		}
	}

	dbProphecy = *prophecy

	serializedProphecyBytes := types.Encode(&dbProphecy)

	err = o.db.Set(x2eTy.CalProphecyPrefix(prophecy.ID), serializedProphecyBytes)
	if err != nil {
		return x2eTy.ErrSetKV
	}
	return nil
}

func (o *Oracle) checkProphecy(prophecy *x2eTy.ReceiptEthProphecy) error {
	if prophecy.ID == "" {
		return x2eTy.ErrInvalidIdentifier
	}
	if len(prophecy.ClaimValidators) == 0 {
		return x2eTy.ErrNoClaims
	}
	return nil
}

//ProcessClaim 处理接收到的ethchain33请求
func (o *Oracle) ProcessClaim(claim x2eTy.Eth2Chain33) (*x2eTy.ProphecyStatus, error) {
	oracleClaim, err := CreateOracleClaimFromEthClaim(claim)
	if err != nil {
		elog.Error("CreateEthClaimFromOracleString", "CreateOracleClaimFromOracleString error", err)
		return nil, err
	}

	activeValidator := o.checkActiveValidator(oracleClaim.ValidatorAddress)
	if !activeValidator {
		return nil, x2eTy.ErrInvalidValidator
	}
	if strings.TrimSpace(oracleClaim.Content) == "" {
		return nil, x2eTy.ErrInvalidClaim
	}
	var claimContent x2eTy.OracleClaimContent
	err = types.Decode([]byte(oracleClaim.Content), &claimContent)
	if err != nil {
		return nil, types.ErrUnmarshal
	}
	prophecy, err := o.GetProphecy(oracleClaim.ID)
	if err != nil {
		if err != x2eTy.ErrProphecyNotFound {
			return nil, err
		}
		prophecy = NewProphecy(oracleClaim.ID)
	} else {
		var exist bool
		for _, vc := range prophecy.ValidatorClaims {
			if vc.Claim == oracleClaim.Content {
				exist = true
			}
		}
		if !exist {
			prophecy.Status.Text = x2eTy.EthBridgeStatus_FailedStatusText
			return nil, x2eTy.ErrClaimInconsist
		}
		if prophecy.Status.Text == x2eTy.EthBridgeStatus_FailedStatusText {
			return nil, x2eTy.ErrProphecyFinalized
		}
		for _, vc := range prophecy.ValidatorClaims {
			if vc.Validator == claim.ValidatorAddress && vc.Claim != "" {
				return nil, x2eTy.ErrDuplicateMessage
			}
		}
	}
	AddClaim(prophecy, oracleClaim.ValidatorAddress, oracleClaim.Content)
	prophecy, err = o.processCompletion(prophecy, claimContent.ClaimType)
	if err != nil {
		return nil, err
	}
	err = o.setProphecy(prophecy)
	if err != nil {
		return nil, err
	}
	return prophecy.Status, nil
}

func (o *Oracle) checkActiveValidator(validatorAddress string) bool {
	validatorMap, err := o.GetValidatorArray()
	if err != nil {
		return false
	}

	for _, v := range validatorMap.Validators {
		if v.Address == validatorAddress {
			return true
		}
	}
	return false
}

// 计算该prophecy是否达标
func (o *Oracle) processCompletion(prophecy *x2eTy.ReceiptEthProphecy, claimType int64) (*x2eTy.ReceiptEthProphecy, error) {
	address2power := make(map[string]int64)
	validatorArrays, err := o.GetValidatorArray()
	if err != nil {
		return nil, err
	}
	for _, validator := range validatorArrays.Validators {
		address2power[validator.Address] = validator.Power
	}
	highestClaim, highestClaimPower, totalClaimsPower := FindHighestClaim(prophecy, address2power)
	totalPower, err := o.GetLastTotalPower()
	if err != nil {
		return nil, err
	}
	highestConsensusRatio := highestClaimPower * 100
	remainingPossibleClaimPower := totalPower - totalClaimsPower
	highestPossibleClaimPower := highestClaimPower + remainingPossibleClaimPower
	highestPossibleConsensusRatio := highestPossibleClaimPower * 100
	elog.Info("processCompletion", "highestConsensusRatio", highestConsensusRatio/totalPower, "ConsensusThreshold", o.consensusThreshold, "highestPossibleConsensusRatio", highestPossibleConsensusRatio/totalPower)
	if highestConsensusRatio >= o.consensusThreshold*totalPower {
		prophecy.Status.Text = x2eTy.EthBridgeStatus_SuccessStatusText

		prophecy.Status.FinalClaim = highestClaim
	} else if highestPossibleConsensusRatio < o.consensusThreshold*totalPower {
		prophecy.Status.Text = x2eTy.EthBridgeStatus_FailedStatusText
	}
	return prophecy, nil
}

//GetLastTotalPower Load the last total validator power.
func (o *Oracle) GetLastTotalPower() (int64, error) {
	b, err := o.db.Get(x2eTy.CalLastTotalPowerPrefix())
	if err != nil && err != types.ErrNotFound {
		return 0, err
	} else if err == types.ErrNotFound {
		return 0, nil
	}
	var powers x2eTy.ReceiptQueryTotalPower
	err = types.Decode(b, &powers)
	if err != nil {
		return 0, types.ErrUnmarshal
	}
	return powers.TotalPower, nil
}

//SetLastTotalPower Set the last total validator power.
func (o *Oracle) SetLastTotalPower() error {
	var totalPower int64
	validatorArrays, err := o.GetValidatorArray()
	if err != nil {
		return err
	}
	for _, validator := range validatorArrays.Validators {
		totalPower += validator.Power
	}
	totalP := x2eTy.ReceiptQueryTotalPower{
		TotalPower: totalPower,
	}
	totalPBytes := types.Encode(&totalP)
	err = o.db.Set(x2eTy.CalLastTotalPowerPrefix(), totalPBytes)
	if err != nil {
		return x2eTy.ErrSetKV
	}
	return nil
}

//GetValidatorArray ...
func (o *Oracle) GetValidatorArray() (*x2eTy.ValidatorList, error) {
	validatorsBytes, err := o.db.Get(x2eTy.CalValidatorMapsPrefix())
	if err != nil {
		return nil, err
	}
	var validatorArrays x2eTy.ValidatorList
	err = types.Decode(validatorsBytes, &validatorArrays)
	if err != nil {
		return nil, types.ErrUnmarshal
	}
	return &validatorArrays, nil
}

//SetConsensusThreshold ...
func (o *Oracle) SetConsensusThreshold(ConsensusThreshold int64) {
	o.consensusThreshold = ConsensusThreshold
	elog.Info("SetConsensusNeeded", "nowConsensusNeeded", o.consensusThreshold)
}

//GetConsensusThreshold ...
func (o *Oracle) GetConsensusThreshold() int64 {
	return o.consensusThreshold
}
