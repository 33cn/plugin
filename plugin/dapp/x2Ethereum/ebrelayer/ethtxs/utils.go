package ethtxs

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	solsha3 "github.com/miguelmota/go-solidity-sha3"

	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type EthTxStatus int32

type nonceMutex struct {
	nonce int64
	rw    *sync.RWMutex
}

var addr2Nonce = make(map[common.Address]nonceMutex)

func (ethTxStatus EthTxStatus) String() string {
	return [...]string{"Fail", "Success", "Pending"}[ethTxStatus]
}

const (
	PendingDuration4TxExeuction = 300
	EthTxFail                   = EthTxStatus(0)
	EthTxSuccess                = EthTxStatus(1)
	EthTxPending                = EthTxStatus(2)
)

// GenerateClaimHash : Generates an OracleClaim hash from a ProphecyClaim's event data
func GenerateClaimHash(prophecyID []byte, sender []byte, recipient []byte, token []byte, amount []byte, validator []byte) common.Hash {
	// Generate a hash containing the information
	rawHash := crypto.Keccak256Hash(prophecyID, sender, recipient, token, amount, validator)

	// Cast hash to hex encoded string
	return rawHash
}

func SignClaim4Eth(hash common.Hash, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	rawSignature, _ := prefixMessage(hash, privateKey)
	signature := hexutil.Bytes(rawSignature)
	return signature, nil
}

func prefixMessage(message common.Hash, key *ecdsa.PrivateKey) ([]byte, []byte) {
	prefixed := solsha3.SoliditySHA3WithPrefix(message[:])
	sig, err := secp256k1.Sign(prefixed, math.PaddedBigBytes(key.D, 32))
	if err != nil {
		panic(err)
	}

	return sig, prefixed
}

func loadPrivateKey(privateKey []byte) (key *ecdsa.PrivateKey, err error) {
	key, err = crypto.ToECDSA(privateKey)
	if nil != err {
		return nil, err
	}
	return
}

// LoadSender : uses the validator's private key to load the validator's address
func LoadSender(privateKey *ecdsa.PrivateKey) (address common.Address, err error) {
	// Parse public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, ebrelayerTypes.ErrPublicKeyType
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	return fromAddress, nil
}

func getNonce(sender common.Address, backend bind.ContractBackend) (*big.Int, error) {
	if nonceMutex, exist := addr2Nonce[sender]; exist {
		nonceMutex.rw.Lock()
		defer nonceMutex.rw.Unlock()
		nonceMutex.nonce += 1
		addr2Nonce[sender] = nonceMutex
		txslog.Debug("getNonce", "address", sender.String(), "nonce", nonceMutex.nonce)
		return big.NewInt(nonceMutex.nonce), nil
	}

	nonce, err := backend.PendingNonceAt(context.Background(), sender)
	if nil != err {
		return nil, err
	}
	txslog.Debug("getNonce", "address", sender.String(), "nonce", nonce)
	n := new(nonceMutex)
	n.nonce = int64(nonce)
	n.rw = new(sync.RWMutex)
	addr2Nonce[sender] = *n
	return big.NewInt(int64(nonce)), nil
}

func revokeNonce(sender common.Address) (*big.Int, error) {
	if nonceMutex, exist := addr2Nonce[sender]; exist {
		nonceMutex.rw.Lock()
		defer nonceMutex.rw.Unlock()
		nonceMutex.nonce -= 1
		addr2Nonce[sender] = nonceMutex
		txslog.Debug("revokeNonce", "address", sender.String(), "nonce", nonceMutex.nonce)
		return big.NewInt(nonceMutex.nonce), nil
	}
	return nil, errors.New("Address doesn't exist tx")
}

func PrepareAuth(backend bind.ContractBackend, privateKey *ecdsa.PrivateKey, transactor common.Address) (*bind.TransactOpts, error) {
	if nil == privateKey || nil == backend {
		txslog.Error("PrepareAuth", "nil input parameter", "backend", backend, "privateKey", privateKey)
		return nil, errors.New("nil input parameter")
	}

	ctx := context.Background()
	gasPrice, err := backend.SuggestGasPrice(ctx)
	if err != nil {
		txslog.Error("PrepareAuth", "Failed to SuggestGasPrice due to:", err.Error())
		return nil, errors.New("Failed to get suggest gas price")
	}
	auth := bind.NewKeyedTransactor(privateKey)
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = GasLimit4Deploy
	auth.GasPrice = gasPrice

	if auth.Nonce, err = getNonce(transactor, backend); err != nil {
		return nil, err
	}

	return auth, nil
}

func waitEthTxFinished(client *ethclient.Client, txhash common.Hash, txName string) error {
	txslog.Info(txName, "Wait for tx to be finished executing with hash", txhash.String())
	timeout := time.NewTimer(PendingDuration4TxExeuction * time.Second)
	oneSecondtimeout := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeout.C:
			return errors.New("Eth tx timeout")
		case <-oneSecondtimeout.C:
			_, err := client.TransactionReceipt(context.Background(), txhash)
			if err == ethereum.NotFound {
				continue
			} else if err != nil {
				return err
			}
			txslog.Info(txName, "Finished executing for tx", txhash.String())
			return nil
		}
	}
}

func GetEthTxStatus(client *ethclient.Client, txhash common.Hash) string {
	receipt, err := client.TransactionReceipt(context.Background(), txhash)
	if nil != err {
		return EthTxPending.String()
	}
	status := EthTxStatus(receipt.Status).String()
	if status != EthTxPending.String() {
		txslog.Info("GetEthTxStatus", "Eth tx hash", txhash.String(), "status", status, "BlockNum", receipt.BlockNumber.Int64())
	}

	return status
}
