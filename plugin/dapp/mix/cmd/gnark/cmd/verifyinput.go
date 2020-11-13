/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/consensys/gnark/backend"
	groth16_bn256 "github.com/consensys/gnark/backend/bn256/groth16"
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gurvy"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyInputCmd = &cobra.Command{
	Use:     "verifystr",
	Short:   "verifies a proof against a verifying key and a partial / public solution",
	Run:     cmdVerifyStr,
	Version: Version,
}
var (
	fProofStr, fInputStr, fVerifyStr string
)

func init() {
	rootCmd.AddCommand(verifyInputCmd)
	verifyInputCmd.PersistentFlags().StringVar(&fVerifyStr, "vk", "", "specifies full path for verifying key")
	verifyInputCmd.PersistentFlags().StringVar(&fInputStr, "input", "", "specifies full path for input file")
	verifyInputCmd.PersistentFlags().StringVar(&fProofStr, "proof", "", "specifies full path for input file")

	_ = verifyInputCmd.MarkPersistentFlagRequired("vk")
	_ = verifyInputCmd.MarkPersistentFlagRequired("input")
	_ = verifyInputCmd.MarkPersistentFlagRequired("proof")
}

func cmdVerifyStr(cmd *cobra.Command, args []string) {
	curveID := gurvy.BN256

	//verify key
	var buffVk bytes.Buffer
	res, err := hex.DecodeString(fVerifyStr)
	if err != nil {
		log.Fatal(err)
	}
	buffVk.Write(res)
	var vk groth16_bn256.VerifyingKey
	if err := gob.Deserialize(&buffVk, &vk, curveID); err != nil {
		fmt.Println("can't load verifying key")
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Printf("%-30s %-30s\n", "loaded verifying key", fVkPath)

	//public input
	rst, err := deserializeInput(fInputStr)
	if err != nil {
		log.Fatal(err)
	}
	r1csInput := backend.NewAssignment()
	for k, v := range rst {
		r1csInput.Assign(backend.Public, k, v)
	}

	// load proof
	var proof groth16_bn256.Proof
	var buffProof bytes.Buffer
	res, err = hex.DecodeString(fProofStr)
	if err != nil {
		log.Fatal(err)
	}
	buffProof.Write(res)
	if err := gob.Deserialize(&buffProof, &proof, curveID); err != nil {
		fmt.Println("can't parse proof", err)
		os.Exit(-1)
	}

	// verify proof
	start := time.Now()
	result, err := groth16_bn256.Verify(&proof, &vk, r1csInput)
	if err != nil || !result {
		fmt.Printf("%-30s %-30s %-30s\n", "proof is invalid", "proofPath", time.Since(start))
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(-1)
	}
	fmt.Printf("%-30s %-30s %-30s\n", "proof is valid", "proofPath", time.Since(start))

}

func deserializeInput(input string) (map[string]interface{}, error) {
	var buffInput bytes.Buffer
	res, err := hex.DecodeString(input)
	if err != nil {
		log.Fatal(err)
	}
	buffInput.Write(res)

	decoder := json.NewDecoder(&buffInput)
	toRead := make(map[string]interface{})

	if err := decoder.Decode(&toRead); err != nil {
		return nil, err
	}

	return toRead, nil
}
