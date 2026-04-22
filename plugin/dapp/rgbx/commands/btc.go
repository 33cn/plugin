package commands

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/spf13/cobra"
)

func btcAddrScriptCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addrScript",
		Short: "convert btc address to pkScript hex",
		Run:   btcAddrScript,
	}
	cmd.Flags().StringP("address", "a", "", "bitcoin address")
	cmd.Flags().String("net", "mainnet", "bitcoin network: mainnet|testnet|regtest|simnet")
	markRequired(cmd, "address")
	return cmd
}

func btcAddrScript(cmd *cobra.Command, _ []string) {
	address, _ := cmd.Flags().GetString("address")
	netName, _ := cmd.Flags().GetString("net")

	params, err := parseNetParams(netName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid net: %s, err: %v\n", netName, err)
		return
	}
	addr, err := btcutil.DecodeAddress(address, params)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid address: %s, decode err: %v\n", address, err)
		return
	}
	script, err := txscript.PayToAddrScript(addr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "convert address to script failed: %v\n", err)
		return
	}
	fmt.Println(hex.EncodeToString(script))
}

func btcDepositTxCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btcDepositTx",
		Short: "build, sign and broadcast btc deposit tx",
		Run:   btcDepositTx,
		Example: "btcDepositTx --net regtest --rpcHost 127.0.0.1:18443 " +
			"--wif <wif> --utxo <txid:vout:amountSats:pkScriptHex> --tssAddress <btcAddr> --chain33Address <addr> " +
			"--amount 100000 --fee 300",
	}
	cmd.Flags().String("net", "regtest", "bitcoin network: mainnet|testnet|regtest|simnet")
	cmd.Flags().String("rpcHost", "127.0.0.1:18443", "bitcoin rpc host")
	cmd.Flags().String("rpcUser", "", "bitcoin rpc user (optional)")
	cmd.Flags().String("rpcPass", "", "bitcoin rpc password (optional)")
	cmd.Flags().Bool("disableTLS", true, "disable rpc tls")
	cmd.Flags().String("rpcCertFile", "", "bitcoin rpc cert file path (optional, required when TLS enabled)")
	cmd.Flags().String("wif", "", "sender private key in WIF format")
	cmd.Flags().String("utxo", "", "single input utxo, format: txid:vout:amountSats:pkScriptHex")
	cmd.Flags().String("tssAddress", "", "tss deposit address")
	cmd.Flags().String("chain33Address", "", "chain33 deposit address for OP_RETURN rgbx:deposit:<addr>")
	cmd.Flags().Int64("amount", 0, "deposit amount in satoshis")
	cmd.Flags().Int64("fee", 0, "tx fee in satoshis")
	cmd.Flags().String("changeAddress", "", "optional change address, default from private key")
	markRequired(cmd, "wif", "utxo", "tssAddress", "chain33Address", "amount", "fee")
	return cmd
}

type depositUTXO struct {
	hash     *chainhash.Hash
	vout     uint32
	amount   int64
	pkScript []byte
}

func parseDepositUTXO(raw string) (*depositUTXO, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid utxo format: %s", raw)
	}
	hash, err := chainhash.NewHashFromStr(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid utxo txid: %w", err)
	}
	vout64, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid utxo vout: %w", err)
	}
	amount, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
	if err != nil || amount <= 0 {
		return nil, fmt.Errorf("invalid utxo amount: %s", parts[2])
	}
	pkScript, err := hex.DecodeString(strings.TrimSpace(parts[3]))
	if err != nil {
		return nil, fmt.Errorf("invalid utxo pkScript: %w", err)
	}
	return &depositUTXO{
		hash:     hash,
		vout:     uint32(vout64),
		amount:   amount,
		pkScript: pkScript,
	}, nil
}

func btcDepositTx(cmd *cobra.Command, _ []string) {
	netName, _ := cmd.Flags().GetString("net")
	rpcHost, _ := cmd.Flags().GetString("rpcHost")
	rpcUser, _ := cmd.Flags().GetString("rpcUser")
	rpcPass, _ := cmd.Flags().GetString("rpcPass")
	disableTLS, _ := cmd.Flags().GetBool("disableTLS")
	rpcCertFile, _ := cmd.Flags().GetString("rpcCertFile")
	wifStr, _ := cmd.Flags().GetString("wif")
	utxoRaw, _ := cmd.Flags().GetString("utxo")
	tssAddrStr, _ := cmd.Flags().GetString("tssAddress")
	chain33Addr, _ := cmd.Flags().GetString("chain33Address")
	amount, _ := cmd.Flags().GetInt64("amount")
	fee, _ := cmd.Flags().GetInt64("fee")
	changeAddrStr, _ := cmd.Flags().GetString("changeAddress")

	params, err := parseNetParams(netName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid net: %s, err: %v\n", netName, err)
		return
	}
	if amount <= 0 || fee < 0 {
		_, _ = fmt.Fprintf(os.Stderr, "invalid amount(%d) or fee(%d)\n", amount, fee)
		return
	}
	utxo, err := parseDepositUTXO(utxoRaw)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid utxo: %v\n", err)
		return
	}
	if utxo.amount < amount+fee {
		_, _ = fmt.Fprintf(os.Stderr, "insufficient utxo amount, have=%d need=%d\n", utxo.amount, amount+fee)
		return
	}

	wif, err := btcutil.DecodeWIF(strings.TrimSpace(wifStr))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid wif: %v\n", err)
		return
	}
	tssAddr, err := btcutil.DecodeAddress(strings.TrimSpace(tssAddrStr), params)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid tssAddress: %v\n", err)
		return
	}
	var changeAddr btcutil.Address
	if strings.TrimSpace(changeAddrStr) == "" {
		changeAddr, err = btcutil.NewAddressPubKeyHash(
			btcutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed()), params,
		)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "create default change address failed: %v\n", err)
			return
		}
	} else {
		changeAddr, err = btcutil.DecodeAddress(strings.TrimSpace(changeAddrStr), params)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "invalid changeAddress: %v\n", err)
			return
		}
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(utxo.hash, utxo.vout), nil, nil))

	tssScript, err := txscript.PayToAddrScript(tssAddr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "build tss script failed: %v\n", err)
		return
	}
	tx.AddTxOut(wire.NewTxOut(amount, tssScript))

	opData := []byte("rgbx:deposit:" + strings.TrimSpace(chain33Addr))
	opScript, err := txscript.NullDataScript(opData)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "build op_return failed: %v\n", err)
		return
	}
	tx.AddTxOut(wire.NewTxOut(0, opScript))

	change := utxo.amount - amount - fee
	if change > 546 {
		changeScript, err := txscript.PayToAddrScript(changeAddr)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "build change script failed: %v\n", err)
			return
		}
		tx.AddTxOut(wire.NewTxOut(change, changeScript))
	}

	class := txscript.GetScriptClass(utxo.pkScript)
	switch class {
	case txscript.PubKeyHashTy:
		sigScript, err := txscript.SignatureScript(tx, 0, utxo.pkScript, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "sign p2pkh failed: %v\n", err)
			return
		}
		tx.TxIn[0].SignatureScript = sigScript
	default:
		_, _ = fmt.Fprintf(os.Stderr, "unsupported utxo script class: %s, only p2pkh supported now\n", class.String())
		return
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         rpcHost,
		User:         rpcUser,
		Pass:         rpcPass,
		HTTPPostMode: true,
		DisableTLS:   disableTLS,
	}
	if !disableTLS && strings.TrimSpace(rpcCertFile) != "" {
		certs, err := os.ReadFile(strings.TrimSpace(rpcCertFile))
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "read rpc cert file failed: %v\n", err)
			return
		}
		connCfg.Certificates = certs
	}
	rpcCli, err := rpcclient.New(connCfg, nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "create rpc client failed: %v\n", err)
		return
	}
	defer rpcCli.Shutdown()

	txHash, err := rpcCli.SendRawTransaction(tx, false)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "broadcast tx failed: %v\n", err)
		return
	}
	fmt.Println(txHash.String())
}

func btcKeyInfoCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "btcKeyInfo",
		Short:   "derive btc key info from private key hex",
		Run:     btcKeyInfo,
		Example: "btcKeyInfo --net regtest --privHex 0000000000000000000000000000000000000000000000000000000000000001",
	}
	cmd.Flags().String("net", "regtest", "bitcoin network: mainnet|testnet|regtest|simnet")
	cmd.Flags().String("privHex", "", "private key hex (32 bytes)")
	markRequired(cmd, "privHex")
	return cmd
}

func btcKeyInfo(cmd *cobra.Command, _ []string) {
	netName, _ := cmd.Flags().GetString("net")
	privHex, _ := cmd.Flags().GetString("privHex")

	params, err := parseNetParams(netName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid net: %s, err: %v\n", netName, err)
		return
	}
	keyBytes, err := hex.DecodeString(strings.TrimSpace(privHex))
	if err != nil || len(keyBytes) != 32 {
		_, _ = fmt.Fprintf(os.Stderr, "invalid privHex, require 32-byte hex\n")
		return
	}
	privKey, _ := btcec.PrivKeyFromBytes(keyBytes)
	wif, err := btcutil.NewWIF(privKey, params, true)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "create wif failed: %v\n", err)
		return
	}
	addr, err := btcutil.NewAddressPubKeyHash(
		btcutil.Hash160(privKey.PubKey().SerializeCompressed()), params,
	)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "derive address failed: %v\n", err)
		return
	}
	script, err := txscript.PayToAddrScript(addr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "derive pkScript failed: %v\n", err)
		return
	}

	fmt.Printf("{\"wif\":\"%s\",\"address\":\"%s\",\"pkScript\":\"%s\"}\n",
		wif.String(), addr.String(), hex.EncodeToString(script))
}

func parseNetParams(net string) (*chaincfg.Params, error) {
	switch strings.ToLower(strings.TrimSpace(net)) {
	case "mainnet", "main":
		return &chaincfg.MainNetParams, nil
	case "testnet", "testnet3", "test":
		return &chaincfg.TestNet3Params, nil
	case "regtest":
		return &chaincfg.RegressionNetParams, nil
	case "simnet":
		return &chaincfg.SimNetParams, nil
	default:
		return nil, fmt.Errorf("unsupported net %q", net)
	}
}

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
