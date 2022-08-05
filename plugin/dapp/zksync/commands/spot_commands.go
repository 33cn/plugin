/*Package commands implement dapp client commands*/
package commands

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/executor/spot"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
)

/*
 * 实现合约对应客户端
 */

func getExecname(paraName string) string {
	exec := et.ExecName
	if strings.HasPrefix(paraName, pt.ParaPrefix) {
		exec = paraName + et.ExecName
	}
	return exec
}

// spotCmd Cmd spot client command
func spotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spot",
		Short: "spot related cmd",
	}
	cmd.AddCommand(limitOrderCmd())
	cmd.AddCommand(revokeOrderCmd())
	cmd.AddCommand(nftOrderCmd())
	cmd.AddCommand(nftTakerOrderCmd())
	cmd.AddCommand(QueryNftOrderCmd())
	cmd.AddCommand(nftOrder2Cmd())
	cmd.AddCommand(nftTakerOrder2Cmd())
	cmd.AddCommand(nftTakerOrderCmd())
	cmd.AddCommand(assetOrderCmd())
	return cmd
}

func limitOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zkLimitOrder",
		Short: "create limit order transaction",
		Run:   limitOrder,
	}
	limitOrderFlag(cmd)
	return cmd
}

func markRequired(cmd *cobra.Command, params ...string) {
	for _, param := range params {
		_ = cmd.MarkFlagRequired(param)
	}
}

// ratio: p * 1e8, 1e8
func limitOrderFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("leftTokenId", "l", 0, "left token id")
	cmd.Flags().Uint64P("rightTokenId", "r", 0, "right token id")
	cmd.Flags().Uint64P("price", "p", 0, "price 1e8 lt = p rt ")
	cmd.Flags().Uint64P("amount", "a", 0, "to buy/sell amount of left token")
	cmd.Flags().StringP("op", "o", "1", "1/buy, 2/sell")

	// zkorder part
	cmd.Flags().Uint64P("accountId", "", 0, "accountid of self")
	cmd.Flags().StringP("ethAddress", "", "", "eth address of self")

	markRequired(cmd, "leftTokenId", "rightTokenId", "price", "amount", "op", "accountId", "ethAddress")
}

func limitOrder(cmd *cobra.Command, args []string) {

	lt, _ := cmd.Flags().GetUint64("leftTokenId")
	rt, _ := cmd.Flags().GetUint64("rightTokenId")
	price, _ := cmd.Flags().GetUint64("price")
	amount, _ := cmd.Flags().GetUint64("amount")
	op, _ := cmd.Flags().GetString("op")
	opInt := 1
	// 业务 buy = buy-Left, sell-Right
	// ratio参数 要求 sell的比较在前   R1:R2 = R:L = price : 1
	buy := lt
	sell := rt
	ratio1 := int64(price)
	ratio2 := int64(1e8)
	if op == "2" || op == "sell" {
		opInt = 2
		buy = rt
		sell = lt
		ratio1, ratio2 = ratio2, ratio1
	}
	accountid, _ := cmd.Flags().GetUint64("accountId")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	zkorder := et.ZkOrder{
		AccountID:  accountid,
		EthAddress: ethAddress,
		TokenSell:  sell,
		TokenBuy:   buy,
		Amount:     et.AmountToZksync(amount),
		Ratio1:     big.NewInt(ratio1).String(),
		Ratio2:     big.NewInt(ratio2).String(),
	}
	// sign

	payload := &et.SpotLimitOrder{
		LeftAsset:  lt,
		RightAsset: rt,
		Price:      int64(price),
		Amount:     int64(amount),
		Op:         int32(opInt),
		Order:      &zkorder,
	}
	paraName, _ := cmd.Flags().GetString("paraName")
	params := &rpctypes.CreateTxIn{
		Execer:     getExecname(paraName),
		ActionName: "LimitOrder",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func revokeOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zkRevokeOrder",
		Short: "create revoke limit order transaction",
		Run:   revokeOrder,
	}
	revokeOrderFlag(cmd)
	return cmd
}

func revokeOrderFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("orderid", "o", 0, "order id")
	cmd.MarkFlagRequired("orderid")
}

func revokeOrder(cmd *cobra.Command, args []string) {
	orderid, _ := cmd.Flags().GetUint64("orderid")
	paraName, _ := cmd.Flags().GetString("paraName")

	payload := &et.SpotRevokeOrder{
		OrderID: int64(orderid),
	}
	params := &rpctypes.CreateTxIn{
		Execer:     getExecname(paraName),
		ActionName: "RevokeOrder",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func nftOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sell_nft",
		Short: "create nft sell order transaction",
		Run:   nftOrder,
	}
	NftOrderFlag(cmd)
	return cmd
}

// ratio: p * 1e8, 1e8
func NftOrderFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("leftTokenId", "l", 0, "(nft)left token id")
	cmd.Flags().Uint64P("rightTokenId", "r", 0, "right token id")
	cmd.Flags().Uint64P("price", "p", 0, "price 1e8 lt = p rt ")
	cmd.Flags().Uint64P("amount", "a", 0, "to buy/sell amount of left token")

	// zkorder part
	cmd.Flags().Uint64P("accountId", "", 0, "accountid of self")
	cmd.Flags().StringP("ethAddress", "", "", "eth address of self")

	markRequired(cmd, "leftTokenId", "rightTokenId", "price", "amount", "accountId", "ethAddress")
}

func nftOrder(cmd *cobra.Command, args []string) {
	lt, _ := cmd.Flags().GetUint64("leftTokenId")
	rt, _ := cmd.Flags().GetUint64("rightTokenId")
	price, _ := cmd.Flags().GetUint64("price")
	amount, _ := cmd.Flags().GetUint64("amount")
	op := "sell"
	opInt := 1
	// 业务 buy = buy-Left, sell-Right
	// ratio参数 要求 sell的比较在前   R1:R2 = R:L = price : 1
	buy := lt
	sell := rt
	// r1:r2 = 1:price*1e10
	if op == "2" || op == "sell" {
		opInt = 2
		buy = rt
		sell = lt
	}
	accountid, _ := cmd.Flags().GetUint64("accountId")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	zkorder := et.ZkOrder{
		AccountID:  accountid,
		EthAddress: ethAddress,
		TokenSell:  sell,
		TokenBuy:   buy,
		Amount:     et.AmountToZksync(amount),
		Ratio1:     big.NewInt(1).String(),
		Ratio2:     big.NewInt(0).Mul(big.NewInt(int64(price)), big.NewInt(1e10)).String(),
	}
	// sign

	payload := &et.SpotNftOrder{
		LeftAsset:  lt,
		RightAsset: rt,
		Price:      int64(price),
		Amount:     int64(amount),
		Op:         int32(opInt),
		Order:      &zkorder,
	}
	paraName, _ := cmd.Flags().GetString("paraName")
	params := &rpctypes.CreateTxIn{
		Execer:     getExecname(paraName),
		ActionName: "SpotNTFOrder",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func nftTakerOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buy_nft",
		Short: "create nft buy order transaction",
		Run:   nftTakerOrder,
	}
	NftTakerOrderFlag(cmd)
	return cmd
}

// ratio: p * 1e8, 1e8
func NftTakerOrderFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("order", "o", 0, "(nft) order id")
	// zkorder part
	cmd.Flags().Uint64P("accountId", "", 0, "accountid of self")
	cmd.Flags().StringP("ethAddress", "", "", "eth address of self")

	markRequired(cmd, "order", "accountId", "ethAddress")
}

func nftTakerOrder(cmd *cobra.Command, args []string) {
	orderId, _ := cmd.Flags().GetInt64("order")
	getNftOrder(cmd, args)
	var order2 et.SpotOrder
	if order2.Ty != et.TyNftOrderAction {
		fmt.Printf("%022d the order is not nft sell order", orderId)
		return
	}
	// 业务 buy = buy-Left, sell-Right
	// ratio参数 要求 sell的比较在前   R1:R2 = R:L = price : 1
	accountid, _ := cmd.Flags().GetUint64("accountId")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	zkorder := et.ZkOrder{
		AccountID:  accountid,
		EthAddress: ethAddress,
		TokenSell:  order2.GetNftOrder().Order.TokenBuy,
		TokenBuy:   order2.GetNftOrder().Order.TokenSell,
		Amount:     order2.GetNftOrder().Order.Amount,
		Ratio1:     order2.GetNftOrder().Order.Ratio2,
		Ratio2:     order2.GetNftOrder().Order.Ratio1,
	}
	// sign

	payload := &et.SpotNftTakerOrder{
		OrderID: orderId,
		Order:   &zkorder,
	}
	paraName, _ := cmd.Flags().GetString("paraName")
	params := &rpctypes.CreateTxIn{
		Execer:     getExecname(paraName),
		ActionName: "SpotNTFTakerOrder",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func QueryNftOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nft_order",
		Short: "query nft sell order transaction",
		Run:   queryNftOrder,
	}
	queryNftOrderFlag(cmd)
	return cmd
}

// ratio: p * 1e8, 1e8
func queryNftOrderFlag(cmd *cobra.Command) {
	cmd.Flags().Int64P("order", "o", 0, "(nft) order id")
	markRequired(cmd, "order")
}

func queryNftOrder1(cmd *cobra.Command, args []string) *et.SpotOrder {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	orderId, _ := cmd.Flags().GetInt64("order")

	var params rpctypes.Query4Jrpc

	paraName, _ := cmd.Flags().GetString("paraName")
	params.Execer = getExecname(paraName)
	req := &et.SpotQueryOrder{
		OrderID: orderId,
	}

	params.FuncName = "QueryNftOrder"
	params.Payload = types.MustPBToJSON(req)

	var resp et.SpotOrder
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
	return &resp
}

func queryNftOrder(cmd *cobra.Command, args []string) {
	queryNftOrder1(cmd, args)
}

func getNftOrder(cmd *cobra.Command, args []string) *et.SpotOrder {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	orderId, _ := cmd.Flags().GetInt64("order")

	var params rpctypes.Query4Jrpc

	paraName, _ := cmd.Flags().GetString("paraName")
	params.Execer = getExecname(paraName)
	req := &et.SpotQueryOrder{
		OrderID: orderId,
	}

	params.FuncName = "QueryNftOrder"
	params.Payload = types.MustPBToJSON(req)

	var resp et.SpotOrder
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.RunResult()
	return &resp
}

func nftOrder2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sell_nft2",
		Short: "create nft sell order transaction",
		Run:   nftOrder2,
	}
	NftOrder2Flag(cmd)
	return cmd
}

// ratio: p * 1e8, 1e8
func NftOrder2Flag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("leftTokenId", "l", 0, "(nft)left token id")
	cmd.Flags().Uint64P("rightTokenId", "r", 0, "right token id")
	cmd.Flags().Uint64P("price", "p", 0, "price 1e8 lt = p rt ")
	cmd.Flags().Uint64P("amount", "a", 0, "to buy/sell amount of left token")

	// zkorder part
	cmd.Flags().Uint64P("accountId", "", 0, "accountid of self")
	cmd.Flags().StringP("ethAddress", "", "", "eth address of self")

	markRequired(cmd, "leftTokenId", "rightTokenId", "price", "amount", "accountId", "ethAddress")
}

func nftOrder2(cmd *cobra.Command, args []string) {
	lt, _ := cmd.Flags().GetUint64("leftTokenId")
	rt, _ := cmd.Flags().GetUint64("rightTokenId")
	price, _ := cmd.Flags().GetUint64("price")
	amount, _ := cmd.Flags().GetUint64("amount")
	op := "sell"
	opInt := 2
	// 业务 buy = buy-Left, sell-Right
	// ratio参数 要求 sell的比较在前   R1:R2 = R:L = price : 1
	buy := lt
	sell := rt
	// r1:r2 = 1:price*1e10
	if op == "2" || op == "sell" {
		opInt = 2
		buy = rt
		sell = lt
	}
	accountid, _ := cmd.Flags().GetUint64("accountId")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	zkorder := et.ZkOrder{
		AccountID:  accountid,
		EthAddress: ethAddress,
		TokenSell:  sell,
		TokenBuy:   buy,
		Amount:     et.NftAmountToZksync(amount),
		Ratio1:     big.NewInt(1).String(),
		Ratio2:     big.NewInt(0).Mul(big.NewInt(int64(price)), big.NewInt(1e10)).String(),
	}
	// sign

	payload := &et.SpotNftOrder{
		LeftAsset:  lt,
		RightAsset: rt,
		Price:      int64(price),
		Amount:     int64(amount),
		Op:         int32(opInt),
		Order:      &zkorder,
	}
	paraName, _ := cmd.Flags().GetString("paraName")
	params := &rpctypes.CreateTxIn{
		Execer:     getExecname(paraName),
		ActionName: "NftOrder2",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func nftTakerOrder2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buy_nft2",
		Short: "create nft buy order transaction",
		Run:   nftTakerOrder2,
	}
	NftTakerOrder2Flag(cmd)
	return cmd
}

// ratio: p * 1e8, 1e8
func NftTakerOrder2Flag(cmd *cobra.Command) {
	cmd.Flags().Int64P("order", "o", 0, "(nft) order id")
	// zkorder part
	cmd.Flags().Uint64P("accountId", "", 0, "accountid of self")
	cmd.Flags().StringP("ethAddress", "", "", "eth address of self")

	markRequired(cmd, "order", "accountId", "ethAddress")
}

func nftTakerOrder2(cmd *cobra.Command, args []string) {
	orderId, _ := cmd.Flags().GetInt64("order")
	order2 := getNftOrder(cmd, args)
	if order2 == nil {
		fmt.Println("get nft order failed")
		return
	}
	if order2.Ty != et.TyNftOrder2Action {
		fmt.Printf("%022d the order is not nft sell order", orderId)
		return
	}
	// 业务 buy = buy-Left, sell-Right
	// ratio参数 要求 sell的比较在前   R1:R2 = R:L = price : 1
	accountid, _ := cmd.Flags().GetUint64("accountId")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	zkorder := et.ZkOrder{
		AccountID:  accountid,
		EthAddress: ethAddress,
		TokenSell:  order2.GetNftOrder().Order.TokenBuy,
		TokenBuy:   order2.GetNftOrder().Order.TokenSell,
		Amount:     order2.GetNftOrder().Order.Amount,
		Ratio1:     order2.GetNftOrder().Order.Ratio2,
		Ratio2:     order2.GetNftOrder().Order.Ratio1,
	}
	// sign

	payload := &et.SpotNftTakerOrder{
		OrderID: orderId,
		Order:   &zkorder,
	}
	paraName, _ := cmd.Flags().GetString("paraName")
	params := &rpctypes.CreateTxIn{
		Execer:     getExecname(paraName),
		ActionName: "NftTakerOrder2",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func assetOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zkAssetOrder",
		Short: "create asset limit order transaction",
		Run:   assetLimitOrder,
	}
	assetLimitOrderFlag(cmd)
	return cmd
}

// ratio: p * 1e8, 1e8
func assetLimitOrderFlag(cmd *cobra.Command) {
	assetTypeInfo := "1 for zksync token, 2 for token,  3 for zksync nft, 4 for evm nft, default 1"
	cmd.Flags().Int32P("leftAssetType", "", 1, assetTypeInfo)
	cmd.Flags().StringP("leftAssetExec", "", "zksync", "default zksync")
	cmd.Flags().StringP("leftAssetSymbol", "", "", "token symbol or nft id")
	cmd.Flags().Int32P("rightAssetType", "", 1, assetTypeInfo)
	cmd.Flags().StringP("rightAssetExec", "", "zksync", "default zksync")
	cmd.Flags().StringP("rightAssetSymbol", "", "", "token symbol or nft id")
	cmd.Flags().Uint64P("price", "p", 0, "price 1e8 lt = p rt ")
	cmd.Flags().Uint64P("amount", "a", 0, "to buy/sell amount of left token")
	cmd.Flags().StringP("op", "o", "1", "1/buy, 2/sell")

	// zkorder part
	cmd.Flags().Uint64P("accountId", "", 0, "accountid of self")
	cmd.Flags().StringP("ethAddress", "", "", "eth address of self")

	markRequired(cmd, "leftAssetSymbol", "rightAssetSymbol", "price", "amount", "op", "accountId", "ethAddress")
}

func makeAsset(ty int32, exec, symbol string) (*et.ZkAsset, error) {
	x := et.AssetType(ty)
	switch x {
	case et.AssetType_L1Erc20:
		n, err := strconv.ParseUint(symbol, 10, 64)
		if err != nil {
			return nil, err
		}
		return spot.NewZkAsset(n), nil
	case et.AssetType_Token:
		return &et.ZkAsset{
			Ty: et.AssetType_Token,
			Value: &et.ZkAsset_TokenAsset{
				&et.TokenAsset{
					Execer: exec,
					Symbol: symbol,
				},
			},
		}, nil
	case et.AssetType_ZkNft:
		n, err := strconv.ParseUint(symbol, 10, 64)
		if err != nil {
			return nil, err
		}
		return spot.NewZkNftAsset(n), nil
	case et.AssetType_EvmNft:
		n, err := strconv.ParseUint(symbol, 10, 64)
		if err != nil {
			return nil, err
		}
		return spot.NewEvmNftAsset(n), nil
	}
	panic("not support asset type")
}

func assetLimitOrder(cmd *cobra.Command, args []string) {
	leftAssetType, _ := cmd.Flags().GetInt32("leftAssetType")
	leftAssetExec, _ := cmd.Flags().GetString("leftAssetExec")
	leftAssetSymbol, _ := cmd.Flags().GetString("leftAssetSymbol")

	rightAssetType, _ := cmd.Flags().GetInt32("rightAssetType")
	rightAssetExec, _ := cmd.Flags().GetString("rightAssetExec")
	rightAssetSymbol, _ := cmd.Flags().GetString("rightAssetSymbol")

	left, err := makeAsset(leftAssetType, leftAssetExec, leftAssetSymbol)
	if err != nil {
		return
	}
	right, err := makeAsset(rightAssetType, rightAssetExec, rightAssetSymbol)
	if err != nil {
		return
	}

	price, _ := cmd.Flags().GetUint64("price")
	amount, _ := cmd.Flags().GetUint64("amount")
	op, _ := cmd.Flags().GetString("op")
	opInt := 1
	// 业务 buy = buy-Left, sell-Right
	// ratio参数 要求 sell的比较在前   R1:R2 = R:L = price : 1

	buy := left
	sell := right
	ratio1 := int64(price)
	ratio2 := int64(1e8)
	if op == "2" || op == "sell" {
		opInt = 2
		buy = right
		sell = left
		ratio1, ratio2 = ratio2, ratio1
	}
	_, _ = buy, sell
	accountid, _ := cmd.Flags().GetUint64("accountId")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	zkorder := et.ZkOrder{
		AccountID:  accountid,
		EthAddress: ethAddress,
		TokenSell:  1, // TODO
		TokenBuy:   2, // TODO
		Amount:     et.AmountToZksync(amount),
		Ratio1:     big.NewInt(ratio1).String(),
		Ratio2:     big.NewInt(ratio2).String(),
	}
	// sign

	payload := &et.SpotAssetLimitOrder{
		LeftAsset:  left,
		RightAsset: right,
		Price:      int64(price),
		Amount:     int64(amount),
		Op:         int32(opInt),
		Order:      &zkorder,
	}
	paraName, _ := cmd.Flags().GetString("paraName")
	params := &rpctypes.CreateTxIn{
		Execer:     getExecname(paraName),
		ActionName: "AssetLimitOrder",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}
