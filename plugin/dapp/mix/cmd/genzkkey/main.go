package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

var circuits = []struct {
	ty   mixTy.VerifyType
	name string
}{
	{mixTy.VerifyType_DEPOSIT, "deposit"},
	{mixTy.VerifyType_WITHDRAW, "withdraw"},
	{mixTy.VerifyType_TRANSFERINPUT, "transfer_input"},
	{mixTy.VerifyType_TRANSFEROUTPUT, "transfer_output"},
	{mixTy.VerifyType_AUTHORIZE, "auth"},
}

func getCircuit(ty mixTy.VerifyType) (constraint.ConstraintSystem, error) {
	switch ty {
	case mixTy.VerifyType_DEPOSIT:
		return frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &mixTy.DepositCircuit{})
	case mixTy.VerifyType_WITHDRAW:
		return frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &mixTy.WithdrawCircuit{})
	case mixTy.VerifyType_TRANSFERINPUT:
		return frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &mixTy.TransferInputCircuit{})
	case mixTy.VerifyType_TRANSFEROUTPUT:
		return frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &mixTy.TransferOutputCircuit{})
	case mixTy.VerifyType_AUTHORIZE:
		return frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &mixTy.AuthorizeCircuit{})
	}
	return nil, fmt.Errorf("unknown ty %d", ty)
}

func writeHex(f func(io.Writer) (int64, error)) []byte {
	var buf bytes.Buffer
	if _, err := f(&buf); err != nil {
		panic(err)
	}
	return []byte(hex.EncodeToString(buf.Bytes()))
}

func main() {
	outDir := "zkgen/keys"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	os.MkdirAll(outDir, 0755)
	for _, c := range circuits {
		start := time.Now()
		ccs, err := getCircuit(c.ty)
		if err != nil {
			fmt.Printf("circuit %s compile FAIL: %v\n", c.name, err)
			continue
		}
		fmt.Printf("circuit %s compiled (%d constraints) in %s, setup...\n", c.name, ccs.GetNbConstraints(), time.Since(start).Round(time.Second))
		pk, vk, err := groth16.Setup(ccs)
		if err != nil {
			fmt.Printf("circuit %s setup FAIL: %v\n", c.name, err)
			continue
		}
		pkHex := writeHex(pk.WriteTo)
		vkHex := writeHex(vk.WriteTo)
		if err := os.WriteFile(filepath.Join(outDir, "circuit_"+c.name+".pk"), pkHex, 0644); err != nil {
			fmt.Printf("circuit %s write pk FAIL: %v\n", c.name, err)
			continue
		}
		if err := os.WriteFile(filepath.Join(outDir, "circuit_"+c.name+".vk"), vkHex, 0644); err != nil {
			fmt.Printf("circuit %s write vk FAIL: %v\n", c.name, err)
			continue
		}
		fmt.Printf("circuit %s DONE in %s, pk=%dB vk=%dB\n", c.name, time.Since(start).Round(time.Second), len(pkHex), len(vkHex))
	}
	fmt.Println("ALL DONE")
}
