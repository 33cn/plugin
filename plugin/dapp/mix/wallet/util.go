// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"

	"github.com/consensys/gnark/frontend"

	"github.com/pkg/errors"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"

	"github.com/consensys/gnark/backend"

	"github.com/33cn/plugin/plugin/dapp/mix/executor/zksnark"
)

//产生随机秘钥和receivingPk对data DH加密，返回随机秘钥的公钥
func encryptSecretData(req *mixTy.EncryptSecretData) (*mixTy.DHSecret, error) {
	secret, err := common.FromHex(req.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "decode secret")
	}

	return encryptData(req.PeerSecretPubKey, secret)
}

func decryptSecretData(req *mixTy.DecryptSecretData) (*mixTy.SecretData, error) {
	secret, err := common.FromHex(req.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "decode req.secret")
	}
	decrypt, err := decryptData(req.SecretPriKey, req.OneTimePubKey, secret)
	if err != nil {
		return nil, errors.Wrap(err, "decrypt secret")
	}
	var raw mixTy.SecretData
	err = types.Decode(decrypt, &raw)
	if err != nil {
		return nil, errors.Wrap(mixTy.ErrDecryptDataFail, "decode decrypt.secret")
	}
	return &raw, nil
}

func (p *mixPolicy) verifyProofOnChain(ty mixTy.VerifyType, proof *mixTy.ZkProofInfo, vkPath string, verifyOnChain bool) error {
	//vkpath verify
	if !verifyOnChain && len(vkPath) > 0 {
		verifyKey, err := readZkKeyFile(vkPath)
		if err != nil {
			return errors.Wrapf(err, "getVerifyKey path=%s", vkPath)
		}

		pass, err := zksnark.Verify(verifyKey, proof.Proof, proof.PublicInput)
		if err != nil || !pass {
			return errors.Wrapf(err, "zk verify fail")
		}
		return nil
	}

	//线上验证proof,失败的原因有可能circuit,Pk和线上vk不匹配，或不是一起产生的版本
	verify := &mixTy.VerifyProofInfo{
		Ty:    ty,
		Proof: proof,
	}
	//onchain verify
	_, err := p.walletOperate.GetAPI().Query(mixTy.MixX, "VerifyProof", verify)
	return err
}

func (p *mixPolicy) getPaymentKey(addr string) (*mixTy.NoteAccountKey, error) {
	msg, err := p.walletOperate.GetAPI().QueryChain(&types.ChainExecutor{
		Driver:   "mix",
		FuncName: "PaymentPubKey",
		Param:    types.Encode(&types.ReqString{Data: addr}),
	})
	if err != nil {
		return nil, err
	}
	return msg.(*mixTy.NoteAccountKey), err
}

func (p *mixPolicy) getPathProof(exec, symbol, leaf string) (*mixTy.CommitTreeProve, error) {
	msg, err := p.walletOperate.GetAPI().QueryChain(&types.ChainExecutor{
		Driver:   "mix",
		FuncName: "GetTreePath",
		Param:    types.Encode(&mixTy.TreeInfoReq{AssetExec: exec, AssetSymbol: symbol, LeafHash: leaf}),
	})
	if err != nil {
		return nil, err
	}
	return msg.(*mixTy.CommitTreeProve), nil
}

func (p *mixPolicy) getNoteInfo(noteHash string) (*mixTy.WalletNoteInfo, error) {
	if p.walletOperate.IsWalletLocked() {
		return nil, types.ErrWalletIsLocked
	}

	var index mixTy.WalletMixIndexReq
	index.NoteHash = noteHash
	msg, err := p.listMixInfos(&index)
	if err != nil {
		return nil, errors.Wrapf(err, "list  noteHash=%s", noteHash)
	}
	resp := msg.(*mixTy.WalletNoteResp)
	if len(resp.Notes) < 1 {
		return nil, errors.Wrapf(err, "list not found noteHash=%s", noteHash)
	}

	note := msg.(*mixTy.WalletNoteResp).Notes[0]

	return note, nil
}

func (p *mixPolicy) getTreeProof(exec, symbol, leaf string) (*mixTy.TreePathProof, error) {
	//get tree path
	path, err := p.getPathProof(exec, symbol, leaf)
	if err != nil {
		return nil, errors.Wrapf(err, "get tree proof for noteHash=%s,exec=%s,symbol=%s", leaf, exec, symbol)
	}
	var proof mixTy.TreePathProof
	proof.TreePath = path.ProofSet[1:]
	proof.Helpers = path.Helpers
	proof.TreeRootHash = path.RootHash
	return &proof, nil
}

func getCircuit(circuitTy mixTy.VerifyType) (frontend.CompiledConstraintSystem, error) {
	switch circuitTy {
	case mixTy.VerifyType_DEPOSIT:
		return frontend.Compile(ecc.BN254, backend.GROTH16, &mixTy.DepositCircuit{})
	case mixTy.VerifyType_WITHDRAW:
		return frontend.Compile(ecc.BN254, backend.GROTH16, &mixTy.WithdrawCircuit{})
	case mixTy.VerifyType_TRANSFERINPUT:
		return frontend.Compile(ecc.BN254, backend.GROTH16, &mixTy.TransferInputCircuit{})
	case mixTy.VerifyType_TRANSFEROUTPUT:
		return frontend.Compile(ecc.BN254, backend.GROTH16, &mixTy.TransferOutputCircuit{})
	case mixTy.VerifyType_AUTHORIZE:
		return frontend.Compile(ecc.BN254, backend.GROTH16, &mixTy.AuthorizeCircuit{})
	default:
		return nil, errors.Wrapf(types.ErrInvalidParam, "ty=%d", circuitTy)
	}
}

func getCircuitKeyFileName(circuitTy mixTy.VerifyType) (string, string, error) {
	switch circuitTy {
	case mixTy.VerifyType_DEPOSIT:
		return mixTy.DepositPk, mixTy.DepositVk, nil
	case mixTy.VerifyType_WITHDRAW:
		return mixTy.WithdrawPk, mixTy.WithdrawVk, nil
	case mixTy.VerifyType_TRANSFERINPUT:
		return mixTy.TransInputPk, mixTy.TransInputVk, nil
	case mixTy.VerifyType_TRANSFEROUTPUT:
		return mixTy.TransOutputPk, mixTy.TransOutputVk, nil
	case mixTy.VerifyType_AUTHORIZE:
		return mixTy.AuthPk, mixTy.AuthVk, nil
	default:
		return "", "", errors.Wrapf(types.ErrInvalidParam, "ty=%d", circuitTy)
	}
}

//文件内容存储的是hex string，读的时候直接转换为string即可
func readZkKeyFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrapf(err, "open file=%s", path)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(f)
	if err != nil {
		return "", errors.Wrapf(err, "read file=%s", path)
	}

	return buf.String(), nil
}

func createProof(circuit frontend.CompiledConstraintSystem, pk groth16.ProvingKey, witness frontend.Circuit) (groth16.Proof, error) {
	return groth16.Prove(circuit, pk, witness)

}

func updateTreePath(obj interface{}, treeProof *mixTy.TreePathProof) {
	tv := reflect.ValueOf(obj)
	if tv.Kind() == reflect.Ptr {
		tv = tv.Elem()
	}
	index := 0
	for i, t := range treeProof.TreePath {
		tv.FieldByName("Path" + strconv.Itoa(i)).Addr().Interface().(*frontend.Variable).Assign(t)
		tv.FieldByName("Helper" + strconv.Itoa(i)).Addr().Interface().(*frontend.Variable).Assign(strconv.Itoa(int(treeProof.Helpers[i])))
		tv.FieldByName("Valid" + strconv.Itoa(i)).Addr().Interface().(*frontend.Variable).Assign("1")
		index = i + 1
	}

	//电路变量必须设置
	for i := index; i < mixTy.TreeLevel; i++ {
		tv.FieldByName("Path" + strconv.Itoa(i)).Addr().Interface().(*frontend.Variable).Assign("0")
		tv.FieldByName("Helper" + strconv.Itoa(i)).Addr().Interface().(*frontend.Variable).Assign("0")
		tv.FieldByName("Valid" + strconv.Itoa(i)).Addr().Interface().(*frontend.Variable).Assign("0")

	}
}

func getZkProofKeys(circuitTy mixTy.VerifyType, path, file string, inputs frontend.Circuit) (*mixTy.ZkProofInfo, error) {
	//从电路文件获取电路约束
	circuit, err := getCircuit(circuitTy)
	if err != nil {
		return nil, err
	}
	//从pv 文件读取Pk结构
	pkFile := filepath.Join(path, file)
	pkStr, err := readZkKeyFile(pkFile)
	if err != nil {
		return nil, errors.Wrapf(err, "readZkKeyFile")
	}
	pkBuf, err := mixTy.GetByteBuff(pkStr)
	if err != nil {
		return nil, err
	}

	pk := groth16.NewProvingKey(ecc.BN254)
	if _, err := pk.ReadFrom(pkBuf); err != nil {
		return nil, errors.Wrapf(err, "read pk")
	}

	//产生zk 证明
	proof, err := createProof(circuit, pk, inputs)
	if err != nil {
		return nil, errors.Wrapf(err, "create proof to %s", pkFile)
	}

	var proofKey bytes.Buffer
	if _, err := proof.WriteTo(&proofKey); err != nil {
		return nil, errors.Wrapf(err, "write proof")
	}

	//公开输入序列化
	var pubBuf bytes.Buffer
	_, err = witness.WritePublicTo(&pubBuf, ecc.BN254, inputs)
	if err != nil {
		return nil, errors.Wrapf(err, "write public input")
	}

	return &mixTy.ZkProofInfo{
		Proof:       hex.EncodeToString(proofKey.Bytes()),
		PublicInput: hex.EncodeToString(pubBuf.Bytes()),
	}, nil
}
