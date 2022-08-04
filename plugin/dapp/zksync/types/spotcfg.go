package types

import (
	"fmt"
)

type Econfig struct {
	Banks     []string
	Coins     []CoinCfg
	Exchanges map[string]*Trade // 现货交易、杠杠交易
}

type CoinCfg struct {
	Coin   string
	Execer string
	Name   string
}

// 交易对配置
type Trade struct {
	Symbol       string
	PriceDigits  int32
	AmountDigits int32
	Taker        int32
	Maker        int32
	MinFee       int64
}

func (f *Econfig) GetFeeAddr() string {
	return f.Banks[0]
}

func (f *Econfig) GetFeeAddrID() uint64 {
	return 1
}

// TODO
func (f *Econfig) GetSymbol(left, right string) string {
	return fmt.Sprintf("%v_%v", left, right)
}

func (f *Econfig) GetTrade(left, right string) *Trade {
	symbol := f.GetSymbol(left, right)
	c, ok := f.Exchanges[symbol]
	if !ok {
		return nil
	}
	return c
}

func (t *Trade) GetPriceDigits() int32 {
	if t == nil {
		return 0
	}
	return t.PriceDigits
}

func (t *Trade) GetAmountDigits() int32 {
	if t == nil {
		return 0
	}
	return t.AmountDigits
}

func (t *Trade) GetTaker() int32 {
	if t == nil {
		return 100000
	}
	return t.Taker
}

func (t *Trade) GetMaker() int32 {
	if t == nil {
		return 100000
	}
	return t.Maker
}

func (t *Trade) GetMinFee() int64 {
	if t == nil {
		return 0
	}
	return t.MinFee
}
