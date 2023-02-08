package l2txs

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"strings"
	"time"

	chain33Common "github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	chain33Crypto "github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/golang/protobuf/proto"
)

func createChain33Tx(privateKeyStr, execer string, action proto.Message) (*types.Transaction, error) {
	var driver secp256k1.Driver
	privateKeySli, err := chain33Common.FromHex(privateKeyStr)
	if nil != err {
		return nil, err
	}
	priKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		return nil, err
	}

	fee := int64(1e7)
	toAddr := address.ExecAddress(execer)
	tx := &types.Transaction{Execer: []byte(execer), Payload: types.Encode(action), Fee: fee, To: toAddr}
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()

	err = SignTransaction(priKey, tx)
	if nil != err {
		return nil, err
	}

	return tx, nil
}

func SignTransaction(key chain33Crypto.PrivKey, tx *types.Transaction) (err error) {
	action := new(zksyncTypes.ZksyncAction)
	if err = types.Decode(tx.Payload, action); err != nil {
		return
	}

	privateKey, err := eddsa.GenerateKey(bytes.NewReader(key.Bytes()))
	if err != nil {
		return
	}

	var msg *zksyncTypes.ZkMsg
	var signInfo *zksyncTypes.ZkSignature
	switch action.GetTy() {
	case zksyncTypes.TyDepositAction:
		deposit := action.GetDeposit()
		msg = wallet.GetDepositMsg(deposit)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		deposit.Signature = signInfo
	case zksyncTypes.TyWithdrawAction:
		withDraw := action.GetZkWithdraw()
		msg = wallet.GetWithdrawMsg(withDraw)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		withDraw.Signature = signInfo
	case zksyncTypes.TyContractToTreeAction:
		contractToLeaf := action.GetContractToTree()
		msg = wallet.GetContractToTreeMsg(contractToLeaf)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		contractToLeaf.Signature = signInfo
	case zksyncTypes.TyTreeToContractAction:
		leafToContract := action.GetTreeToContract()
		msg = wallet.GetTreeToContractMsg(leafToContract)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		leafToContract.Signature = signInfo
	case zksyncTypes.TyTransferAction:
		transfer := action.GetZkTransfer()
		msg = wallet.GetTransferMsg(transfer)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		transfer.Signature = signInfo
	case zksyncTypes.TyTransferToNewAction:
		transferToNew := action.GetTransferToNew()
		msg = wallet.GetTransferToNewMsg(transferToNew)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		transferToNew.Signature = signInfo
	case zksyncTypes.TyProxyExitAction:
		forceQuit := action.GetProxyExit()
		msg = wallet.GetProxyExitMsg(forceQuit)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		forceQuit.Signature = signInfo
	case zksyncTypes.TySetPubKeyAction:
		setPubKey := action.GetSetPubKey()
		//如果是添加公钥的操作，则默认设置这里生成的公钥 todo:要是未来修改可以自定义公钥，这里需要删除
		//如果是添加公钥的操作，则默认设置这里生成的公钥
		if setPubKey.PubKeyTy == 0 {
			pubKey := &zksyncTypes.ZkPubKey{
				X: privateKey.PublicKey.A.X.String(),
				Y: privateKey.PublicKey.A.Y.String(),
			}
			setPubKey.PubKey = pubKey
		}

		msg = wallet.GetSetPubKeyMsg(setPubKey)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		setPubKey.Signature = signInfo
	case zksyncTypes.TyFullExitAction:
		forceQuit := action.GetFullExit()
		msg = wallet.GetFullExitMsg(forceQuit)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		forceQuit.Signature = signInfo

	case zksyncTypes.TyMintNFTAction:
		nft := action.GetMintNFT()
		msg := wallet.GetMintNFTMsg(nft)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		nft.Signature = signInfo
	case zksyncTypes.TyTransferNFTAction:
		nft := action.GetTransferNFT()
		msg := wallet.GetTransferNFTMsg(nft)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		nft.Signature = signInfo
	case zksyncTypes.TyWithdrawNFTAction:
		nft := action.GetWithdrawNFT()
		msg := wallet.GetWithdrawNFTMsg(nft)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		nft.Signature = signInfo
	}

	tx.Payload = types.Encode(action)
	tx.Sign(types.SECP256K1, key)
	return
}

func SignTxInEddsa(msg *zksyncTypes.ZkMsg, privateKey eddsa.PrivateKey) (*zksyncTypes.ZkSignature, error) {
	signInfo, err := privateKey.Sign(wallet.GetMsgHash(msg), mimc.NewMiMC(zksyncTypes.ZkMimcHashSeed))
	if err != nil {
		return nil, err
	}
	pubKey := &zksyncTypes.ZkPubKey{
		X: privateKey.PublicKey.A.X.String(),
		Y: privateKey.PublicKey.A.Y.String(),
	}
	sign := &zksyncTypes.ZkSignature{
		PubKey:   pubKey,
		SignInfo: hex.EncodeToString(signInfo),
		Msg:      msg,
	}
	return sign, nil
}

func getRealExecName(paraName string, name string) string {
	if strings.HasPrefix(name, pt.ParaPrefix) {
		return name
	}
	return paraName + name
}
