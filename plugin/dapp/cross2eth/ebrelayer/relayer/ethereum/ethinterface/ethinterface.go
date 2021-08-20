package ethinterface

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

//EthClientSpec ...
type EthClientSpec interface {
	bind.ContractBackend
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	NetworkID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

//SimExtend ...
type SimExtend struct {
	*backends.SimulatedBackend
}

//HeaderByNumber ...
func (sim *SimExtend) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return sim.Blockchain().CurrentBlock().Header(), nil
}

//NetworkID ...
func (sim *SimExtend) NetworkID(ctx context.Context) (*big.Int, error) {
	return nil, nil
}

//func (sim *SimExtend) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
//	receipt, err := sim.SimulatedBackend.TransactionReceipt(ctx, txHash)
//	if receipt == nil {
//		err = errors.New("not found")
//	}
//
//	return receipt, err
//}
