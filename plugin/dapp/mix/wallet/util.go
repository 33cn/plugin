// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	backend_bn256 "github.com/consensys/gnark/backend/bn256"
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gurvy"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"

	"github.com/consensys/gnark/backend"
	groth16_bn256 "github.com/consensys/gnark/backend/bn256/groth16"

	"github.com/33cn/plugin/plugin/dapp/mix/executor/zksnark"
)

//对secretData 编码为string,同时增加随机值
//func encodeSecretData(secret *mixTy.SecretData) (*mixTy.EncodedSecretData, error) {
//	if secret == nil {
//		return nil, errors.Wrap(types.ErrInvalidParam, "para is nil")
//	}
//	if len(secret.ReceiverKey) <= 0 {
//		return nil, errors.Wrap(types.ErrInvalidParam, "spendPubKey is nil")
//	}
//	var val big.Int
//	ret, succ := val.SetString(secret.Amount, 10)
//	if !succ {
//		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong amount = %s", secret.Amount)
//	}
//	if ret.Sign() <= 0 {
//		return nil, errors.Wrapf(types.ErrInvalidParam, "amount = %s, need bigger than 0", secret.Amount)
//	}
//
//	//获取随机值
//	var fr fr_bn256.Element
//	fr.SetRandom()
//	secret.NoteRandom = fr.String()
//	code := types.Encode(secret)
//	var resp mixTy.EncodedSecretData
//
//	resp.Encoded = common.ToHex(code)
//	resp.RawData = secret
//
//	return &resp, nil
//
//}

//产生随机秘钥和receivingPk对data DH加密，返回随机秘钥的公钥
func encryptSecretData(req *mixTy.EncryptSecretData) (*mixTy.DHSecret, error) {
	secret, err := common.FromHex(req.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "decode secret")
	}

	return encryptData(req.PeerKey, secret)
}

func decryptSecretData(req *mixTy.DecryptSecretData) (*mixTy.SecretData, error) {
	secret, err := common.FromHex(req.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "decode req.secret")
	}
	decrypt, err := decryptData(req.PriKey, req.PeerKey, secret)
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

func (p *mixPolicy) verifyProofOnChain(ty mixTy.VerifyType, proof *mixTy.ZkProofInfo, vkPath string, verifyOnChain int32) error {
	//vkpath verify
	if verifyOnChain > 0 && len(vkPath) > 0 {
		vk, err := getVerifyKey(vkPath)
		if err != nil {
			return errors.Wrapf(err, "getVerifyKey path=%s", vkPath)
		}
		verifyKey, err := serializeObj(vk)
		if err != nil {
			return errors.Wrapf(err, "serial vk path=%s", vkPath)
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

func (p *mixPolicy) getPaymentKey(addr string) (*mixTy.PaymentKey, error) {
	msg, err := p.walletOperate.GetAPI().QueryChain(&types.ChainExecutor{
		Driver:   "mix",
		FuncName: "PaymentPubKey",
		Param:    types.Encode(&types.ReqString{Data: addr}),
	})
	if err != nil {
		return nil, err
	}
	return msg.(*mixTy.PaymentKey), err
}

func (p *mixPolicy) getPathProof(leaf string) (*mixTy.CommitTreeProve, error) {
	msg, err := p.walletOperate.GetAPI().QueryChain(&types.ChainExecutor{
		Driver:   "mix",
		FuncName: "GetTreePath",
		Param:    types.Encode(&mixTy.TreeInfoReq{LeafHash: leaf}),
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

func (p *mixPolicy) getTreeProof(leaf string) (*mixTy.TreePathProof, error) {
	//get tree path
	path, err := p.getPathProof(leaf)
	if err != nil {
		return nil, errors.Wrapf(err, "get tree proof for noteHash=%s", leaf)
	}
	var proof mixTy.TreePathProof
	proof.TreePath = path.ProofSet[1:]
	proof.Helpers = path.Helpers
	proof.TreeRootHash = path.RootHash
	return &proof, nil
}

//文件信息过大，pk文件超过1M，作为参数传递不合适，这里传路径信息
func getCircuit(path string) (*backend_bn256.R1CS, error) {
	var bigIntR1cs frontend.R1CS
	if err := gob.Read(path, &bigIntR1cs, gurvy.BN256); err != nil {
		return nil, errors.Wrapf(err, "getCircuit path=%s", path)
	}
	r1cs := backend_bn256.Cast(&bigIntR1cs)
	return &r1cs, nil
}

func getProveKey(path string) (*groth16_bn256.ProvingKey, error) {
	var pk groth16_bn256.ProvingKey
	if err := gob.Read(path, &pk, gurvy.BN256); err != nil {
		return nil, errors.Wrapf(err, "getProveKey path=%s", path)
	}

	return &pk, nil
}

func getVerifyKey(path string) (*groth16_bn256.VerifyingKey, error) {
	var vk groth16_bn256.VerifyingKey
	if err := gob.Read(path, &vk, gurvy.BN256); err != nil {
		return nil, errors.Wrapf(err, "zk.verify.Deserize.VK=%s", path)
	}

	return &vk, nil
}

func createProof(circuit *backend_bn256.R1CS, pk *groth16_bn256.ProvingKey, inputs backend.Assignments) (*groth16_bn256.Proof, error) {
	return groth16_bn256.Prove(circuit, pk, inputs)

}

func verifyProof(proof *groth16_bn256.Proof, vk *groth16_bn256.VerifyingKey, input backend.Assignments) bool {
	ok, err := groth16_bn256.Verify(proof, vk, input)
	if err != nil {
		fmt.Println("err", err)
		return false
	}
	return ok
}

func getAssignments(obj interface{}) (backend.Assignments, error) {
	ty := reflect.TypeOf(obj)
	tv := reflect.ValueOf(obj)
	n := ty.NumField()
	assigns := backend.NewAssignment()
	for i := 0; i < n; i++ {
		name := ty.Field(i).Name
		v, ok := ty.Field(i).Tag.Lookup("tag")
		if !ok {
			return nil, errors.Wrapf(types.ErrNotFound, "fieldname=%s not set tag", ty.Field(i).Name)
		}

		if v != string(backend.Secret) && v != string(backend.Public) {
			return nil, errors.Wrapf(types.ErrInvalidParam, "tag=%s not correct", v)
		}
		assigns.Assign(backend.Visibility(v), name, tv.FieldByName(name).Interface())
	}
	return assigns, nil

}

func serializeObj(from interface{}) (string, error) {
	var buf bytes.Buffer

	err := gob.Serialize(&buf, from, gurvy.BN256)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil

}

func serialInputs(assignments backend.Assignments) (string, error) {
	rst := make(map[string]interface{})
	publics := assignments.DiscardSecrets()
	for k, v := range publics {
		rst[k] = v.Value.String()
	}

	out, err := json.Marshal(rst)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(out), nil
}

func initTreePath(obj interface{}) {
	tv := reflect.ValueOf(obj)
	for i := 0; i < mixTy.TreeLevel; i++ {
		tv.Elem().FieldByName(fmt.Sprintf("Path%d", i)).SetString("0")
		tv.Elem().FieldByName(fmt.Sprintf("Helper%d", i)).SetString("0")
		tv.Elem().FieldByName(fmt.Sprintf("Valid%d", i)).SetString("0")
	}

}
func updateTreePath(obj interface{}, treeProof *mixTy.TreePathProof) {
	tv := reflect.ValueOf(obj)
	for i, t := range treeProof.TreePath {
		tv.Elem().FieldByName("Path" + strconv.Itoa(i)).SetString(t)
		tv.Elem().FieldByName("Helper" + strconv.Itoa(i)).SetString(strconv.Itoa(int(treeProof.Helpers[i])))
		tv.Elem().FieldByName("Valid" + strconv.Itoa(i)).SetString("1")
	}
}

func printObj(obj interface{}) {
	ty := reflect.TypeOf(obj)
	tv := reflect.ValueOf(obj)
	n := ty.NumField()

	for i := 0; i < n; i++ {
		name := ty.Field(i).Name
		v, ok := ty.Field(i).Tag.Lookup("tag")
		if !ok {
			fmt.Println("fieldname=", ty.Field(i).Name, "not set tag")
		}

		fmt.Println("fieldname=", ty.Field(i).Name, "| value=", tv.FieldByName(name).Interface(), "| tag=", v)
	}

}

func getZkProofKeys(circuitFile, pkFile string, inputs interface{}, privacyPrint int32) (*mixTy.ZkProofInfo, error) {
	if privacyPrint > 0 {
		fmt.Println("--output zk parameters for circuit:", circuitFile)
		rst, err := json.MarshalIndent(inputs, "", "    ")
		if err != nil {
			fmt.Println(err)

		}
		fmt.Println(string(rst))
	}

	assignments, err := getAssignments(inputs)
	if err != nil {
		return nil, err
	}

	//从电路文件获取电路约束
	circuit, err := getCircuit(circuitFile)
	if err != nil {
		return nil, err
	}
	//从pv 文件读取Pk结构
	pk, err := getProveKey(pkFile)
	if err != nil {
		return nil, err
	}
	//产生zk 证明
	proof, err := createProof(circuit, pk, assignments)
	if err != nil {
		return nil, err
	}

	//序列号成字符串
	proofKey, err := serializeObj(proof)
	if err != nil {
		return nil, err
	}

	//序列号成字符串
	proofInput, err := serialInputs(assignments)
	if err != nil {
		return nil, err
	}

	return &mixTy.ZkProofInfo{
		Proof:       proofKey,
		PublicInput: proofInput,
	}, nil
}
