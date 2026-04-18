package commands

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/spf13/cobra"
)

func commitDKGCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commitDKG",
		Aliases: []string{"cdkg"},
		Short:   "commit dkg address for cross-chain asset",
		Run:     commitDKG,
		Example: "commitDKG -s BTC -d <dkgAddress> -p <pkScriptHex>",
	}
	commitDKGFlags(cmd)
	return cmd
}

func commitDKGFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("assetSymbol", "s", rtypes.BTCSymbol, "cross-chain asset symbol")
	cmd.Flags().StringP("dkgAddress", "d", "", "dkg/tss bitcoin address")
	cmd.Flags().StringP("pkScript", "p", "", "dkg address pkScript hex")
	markRequired(cmd, "dkgAddress", "pkScript")
}

func commitDKG(cmd *cobra.Command, _ []string) {
	symbol, _ := cmd.Flags().GetString("assetSymbol")
	dkgAddress, _ := cmd.Flags().GetString("dkgAddress")
	pkScriptHex, _ := cmd.Flags().GetString("pkScript")

	pkScript, err := hex.DecodeString(pkScriptHex)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid pkScript: %s, decode err: %v\n", pkScriptHex, err)
		return
	}

	sendCreateTxRPC(cmd, rtypes.NameCommitDKGAction, &rtypes.CommitDKG{
		AssetSymbol: symbol,
		DkgAddress:  dkgAddress,
		PkScript:    pkScript,
	})
}

func depositAssetCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deposit",
		Aliases: []string{"dep"},
		Short:   "submit btc deposit proof",
		Run:     depositAsset,
		Example: "deposit -a 1000 -d <chain33Addr> --txData <btcTxHex> --blockHeight 1 --txIndex 0 --blockHash <hash>",
	}
	depositAssetFlags(cmd)
	return cmd
}

func depositAssetFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("amount", "a", 0, "deposit amount")
	cmd.Flags().StringP("depositAddress", "d", "", "target chain33 address")
	cmd.Flags().StringP("assetSymbol", "s", rtypes.BTCSymbol, "cross-chain asset symbol")
	cmd.Flags().Uint64("blockHeight", 0, "btc proof block height")
	cmd.Flags().Uint32("txIndex", 0, "btc tx index in block")
	cmd.Flags().String("blockHash", "", "btc proof block hash")
	cmd.Flags().String("txData", "", "btc raw transaction hex")
	cmd.Flags().String("merkleProof", "", "comma-separated merkle proof hex list")
	markRequired(cmd, "amount", "depositAddress", "blockHeight", "txIndex", "blockHash", "txData")
}

func depositAsset(cmd *cobra.Command, _ []string) {
	amount, _ := cmd.Flags().GetInt64("amount")
	depositAddress, _ := cmd.Flags().GetString("depositAddress")
	symbol, _ := cmd.Flags().GetString("assetSymbol")
	blockHeight, _ := cmd.Flags().GetUint64("blockHeight")
	txIndex, _ := cmd.Flags().GetUint32("txIndex")
	blockHash, _ := cmd.Flags().GetString("blockHash")
	txDataHex, _ := cmd.Flags().GetString("txData")
	merkleProofStr, _ := cmd.Flags().GetString("merkleProof")

	txData, err := hex.DecodeString(txDataHex)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid txData: %s, decode err: %v\n", txDataHex, err)
		return
	}
	merkleProof, err := decodeHexList(merkleProofStr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid merkleProof: %s, decode err: %v\n", merkleProofStr, err)
		return
	}

	sendCreateTxRPC(cmd, rtypes.NameDepositAssetAction, &rtypes.DepositAsset{
		Amount:         amount,
		DepositAddress: depositAddress,
		AssetSymbol:    symbol,
		TxProof: &rtypes.BtcTxProof{
			BlockHeight: blockHeight,
			TxIndex:     txIndex,
			BlockHash:   blockHash,
			TxData:      txData,
			MerkleProof: merkleProof,
		},
	})
}

func withdrawAssetCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw",
		Aliases: []string{"wd"},
		Short:   "withdraw cross-chain asset to btc address",
		Run:     withdrawAsset,
		Example: "withdraw -a 1000 -f 10 -d <btcAddress> -s BTC",
	}
	withdrawAssetFlags(cmd)
	return cmd
}

func withdrawAssetFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("amount", "a", 0, "withdraw amount")
	cmd.Flags().Int64P("feeRate", "f", 1, "btc fee rate (sat/vbyte)")
	cmd.Flags().StringP("destinationAddr", "d", "", "btc destination address")
	cmd.Flags().StringP("assetSymbol", "s", rtypes.BTCSymbol, "cross-chain asset symbol")
	markRequired(cmd, "amount", "feeRate", "destinationAddr")
}

func withdrawAsset(cmd *cobra.Command, _ []string) {
	amount, _ := cmd.Flags().GetInt64("amount")
	feeRate, _ := cmd.Flags().GetInt64("feeRate")
	destAddr, _ := cmd.Flags().GetString("destinationAddr")
	symbol, _ := cmd.Flags().GetString("assetSymbol")

	sendCreateTxRPC(cmd, rtypes.NameWithdrawAssetAction, &rtypes.WithdrawAsset{
		Amount:          amount,
		FeeRate:         feeRate,
		DestinationAddr: destAddr,
		AssetSymbol:     symbol,
	})
}

func confirmTxCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "confirm",
		Aliases: []string{"cf"},
		Short:   "confirm pending rgbx tx",
		Run:     confirmTx,
		Example: "confirm --actionType 106 --txBlockHeight 100 --txIndex 1 --txHash <hashHex>",
	}
	confirmTxFlags(cmd)
	return cmd
}

func confirmTxFlags(cmd *cobra.Command) {
	cmd.Flags().Int32("actionType", 0, "pending action type")
	cmd.Flags().Int64("confirmedBlockHeight", 0, "confirmed btc block height")
	cmd.Flags().Int64("txBlockHeight", 0, "pending tx block height")
	cmd.Flags().Int64("txIndex", 0, "pending tx index")
	cmd.Flags().String("txHash", "", "pending tx hash hex")
	cmd.Flags().Bool("timeout", false, "whether confirm by timeout")

	cmd.Flags().Uint32("spendingInputIdx", 0, "btc spending tx input index")
	cmd.Flags().Int32("opRetOutputIdx", -1, "btc op_return output index")
	cmd.Flags().String("spendingTx", "", "btc spending tx raw hex")
	cmd.Flags().String("opRetOutputPkScript", "", "op_return pkScript hex")

	cmd.Flags().Uint64("btcBlockHeight", 0, "btc proof block height")
	cmd.Flags().Uint32("btcTxIndex", 0, "btc proof tx index")
	cmd.Flags().String("btcBlockHash", "", "btc proof block hash")
	cmd.Flags().String("btcTxData", "", "btc proof tx raw hex")
	cmd.Flags().String("btcMerkleProof", "", "comma-separated merkle proof hex list")
	markRequired(cmd, "actionType", "txBlockHeight", "txIndex", "txHash")
}

func confirmTx(cmd *cobra.Command, _ []string) {
	actionType, _ := cmd.Flags().GetInt32("actionType")
	confirmedHeight, _ := cmd.Flags().GetInt64("confirmedBlockHeight")
	txBlockHeight, _ := cmd.Flags().GetInt64("txBlockHeight")
	txIndex, _ := cmd.Flags().GetInt64("txIndex")
	txHashHex, _ := cmd.Flags().GetString("txHash")
	timeout, _ := cmd.Flags().GetBool("timeout")

	spendingInputIdx, _ := cmd.Flags().GetUint32("spendingInputIdx")
	opRetOutputIdx, _ := cmd.Flags().GetInt32("opRetOutputIdx")
	spendingTxHex, _ := cmd.Flags().GetString("spendingTx")
	opRetPkScriptHex, _ := cmd.Flags().GetString("opRetOutputPkScript")

	btcBlockHeight, _ := cmd.Flags().GetUint64("btcBlockHeight")
	btcTxIndex, _ := cmd.Flags().GetUint32("btcTxIndex")
	btcBlockHash, _ := cmd.Flags().GetString("btcBlockHash")
	btcTxDataHex, _ := cmd.Flags().GetString("btcTxData")
	btcMerkleProofStr, _ := cmd.Flags().GetString("btcMerkleProof")

	txHash, err := hex.DecodeString(txHashHex)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid txHash: %s, decode err: %v\n", txHashHex, err)
		return
	}

	spendingTx, err := decodeHexOptional(spendingTxHex)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid spendingTx: %s, decode err: %v\n", spendingTxHex, err)
		return
	}
	opRetPkScript, err := decodeHexOptional(opRetPkScriptHex)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid opRetOutputPkScript: %s, decode err: %v\n", opRetPkScriptHex, err)
		return
	}
	btcTxData, err := decodeHexOptional(btcTxDataHex)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid btcTxData: %s, decode err: %v\n", btcTxDataHex, err)
		return
	}
	btcMerkleProof, err := decodeHexList(btcMerkleProofStr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid btcMerkleProof: %s, decode err: %v\n", btcMerkleProofStr, err)
		return
	}

	sendCreateTxRPC(cmd, rtypes.NameConfirmAction, &rtypes.ConfirmTx{
		ActionType:           actionType,
		ConfirmedBlockHeight: confirmedHeight,
		TxBlockHeight:        txBlockHeight,
		TxIndex:              txIndex,
		TxHash:               txHash,
		Timeout:              timeout,
		UtxoProof: &rtypes.UtxoSpendingProof{
			SpendingInputIdx:    spendingInputIdx,
			OpRetOutputIdx:      opRetOutputIdx,
			SpendingTx:          spendingTx,
			OpRetOutputPkScript: opRetPkScript,
		},
		BtcTxProof: &rtypes.BtcTxProof{
			BlockHeight: btcBlockHeight,
			TxIndex:     btcTxIndex,
			BlockHash:   btcBlockHash,
			TxData:      btcTxData,
			MerkleProof: btcMerkleProof,
		},
	})
}

func decodeHexOptional(s string) ([]byte, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	return hex.DecodeString(s)
}

func decodeHexList(raw string) ([][]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	result := make([][]byte, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		b, err := hex.DecodeString(part)
		if err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, nil
}
