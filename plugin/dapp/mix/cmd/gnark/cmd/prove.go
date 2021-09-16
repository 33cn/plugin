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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/consensys/gnark/backend"
	backend_bls377 "github.com/consensys/gnark/backend/bls377"
	groth16_bls377 "github.com/consensys/gnark/backend/bls377/groth16"
	backend_bls381 "github.com/consensys/gnark/backend/bls381"
	groth16_bls381 "github.com/consensys/gnark/backend/bls381/groth16"
	backend_bn256 "github.com/consensys/gnark/backend/bn256"
	groth16_bn256 "github.com/consensys/gnark/backend/bn256/groth16"
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gurvy"
	"github.com/spf13/cobra"
)

// proveCmd represents the prove command
var proveCmd = &cobra.Command{
	Use: "prove [circuit.r1cs]",

	Short:   "creates a (zk)proof for provided circuit and solution",
	Run:     cmdProve,
	Version: Version,
}

var (
	fProofPath string
	fInputPath string
	fCount     uint
)

func init() {
	rootCmd.AddCommand(proveCmd)
	proveCmd.PersistentFlags().StringVar(&fProofPath, "proof", "", "specifies full path for proof -- default is ./[circuit].proof")
	proveCmd.PersistentFlags().StringVar(&fPkPath, "pk", "", "specifies full path for proving key")
	proveCmd.PersistentFlags().StringVar(&fInputPath, "input", "", "specifies full path for input file")
	proveCmd.PersistentFlags().UintVar(&fCount, "count", 1, "specifies number of times the prover algorithm is ran (benchmarking purposes)")
	_ = proveCmd.MarkPersistentFlagRequired("pk")
	_ = proveCmd.MarkPersistentFlagRequired("input")
}

func cmdProve(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("missing circuit path -- gnark prove -h for help")
		os.Exit(-1)
	}
	circuitPath := filepath.Clean(args[0])
	circuitName := filepath.Base(circuitPath)
	circuitExt := filepath.Ext(circuitName)
	circuitName = circuitName[0 : len(circuitName)-len(circuitExt)]

	// ensure pk and input flags are set and valid
	if fPkPath == "" {
		fmt.Println("please specify proving key path")
		_ = cmd.Usage()
		os.Exit(-1)
	}
	if fInputPath == "" {
		fmt.Println("please specify input file path")
		_ = cmd.Usage()
		os.Exit(-1)
	}
	fPkPath = filepath.Clean(fPkPath)
	if !fileExists(fPkPath) {
		fmt.Println(fPkPath, errNotFound)
		os.Exit(-1)
	}
	fInputPath = filepath.Clean(fInputPath)
	if !fileExists(fInputPath) {
		fmt.Println(fInputPath, errNotFound)
		os.Exit(-1)
	}

	// load circuit
	if !fileExists(circuitPath) {
		fmt.Println("error:", errNotFound)
		os.Exit(-1)
	}

	// check curve ID
	curveID, err := gob.PeekCurveID(circuitPath)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(-1)
	}
	// TODO clean that up with interfaces and type casts
	var bigIntR1cs frontend.R1CS
	switch curveID {
	case gurvy.BLS377:
		if err := gob.Read(circuitPath, &bigIntR1cs, curveID); err != nil {
			fmt.Println("error:", err)
			os.Exit(-1)
		}
		r1cs := backend_bls377.Cast(&bigIntR1cs)
		fmt.Printf("%-30s %-30s %-d constraints\n", "loaded circuit", circuitPath, r1cs.NbConstraints)
		// run setup
		var pk groth16_bls377.ProvingKey
		if err := gob.Read(fPkPath, &pk, curveID); err != nil {
			fmt.Println("can't load proving key")
			fmt.Println(err)
			os.Exit(-1)
		}
		fmt.Printf("%-30s %-30s\n", "loaded proving key", fPkPath)

		// parse input file
		r1csInput := backend.NewAssignment()
		err = r1csInput.ReadFile(fInputPath)
		if err != nil {
			fmt.Println("can't parse input", err)
			os.Exit(-1)
		}
		fmt.Printf("%-30s %-30s %-d inputs\n", "loaded input", fInputPath, len(r1csInput))

		// compute proof
		start := time.Now()
		proof, err := groth16_bls377.Prove(&r1cs, &pk, r1csInput)
		if err != nil {
			fmt.Println("Error proof generation", err)
			os.Exit(-1)
		}
		for i := uint(1); i < fCount; i++ {
			_, _ = groth16_bls377.Prove(&r1cs, &pk, r1csInput)
		}
		duration := time.Since(start)
		if fCount > 1 {
			duration = time.Duration(int64(duration) / int64(fCount))
		}

		// default proof path
		proofPath := filepath.Join(".", circuitName+".proof")
		if fProofPath != "" {
			proofPath = fProofPath
		}

		if err := gob.Write(proofPath, proof, curveID); err != nil {
			fmt.Println("error:", err)
			os.Exit(-1)
		}

		fmt.Printf("%-30s %-30s %-30s\n", "generated proof", proofPath, duration)
	case gurvy.BLS381:
		if err := gob.Read(circuitPath, &bigIntR1cs, curveID); err != nil {
			fmt.Println("error:", err)
			os.Exit(-1)
		}
		r1cs := backend_bls381.Cast(&bigIntR1cs)
		fmt.Printf("%-30s %-30s %-d constraints\n", "loaded circuit", circuitPath, r1cs.NbConstraints)
		// run setup
		var pk groth16_bls381.ProvingKey
		if err := gob.Read(fPkPath, &pk, curveID); err != nil {
			fmt.Println("can't load proving key")
			fmt.Println(err)
			os.Exit(-1)
		}
		fmt.Printf("%-30s %-30s\n", "loaded proving key", fPkPath)

		// parse input file
		r1csInput := backend.NewAssignment()
		err = r1csInput.ReadFile(fInputPath)
		if err != nil {
			fmt.Println("can't parse input", err)
			os.Exit(-1)
		}
		fmt.Printf("%-30s %-30s %-d inputs\n", "loaded input", fInputPath, len(r1csInput))

		// compute proof
		start := time.Now()
		proof, err := groth16_bls381.Prove(&r1cs, &pk, r1csInput)
		if err != nil {
			fmt.Println("Error proof generation", err)
			os.Exit(-1)
		}
		for i := uint(1); i < fCount; i++ {
			_, _ = groth16_bls381.Prove(&r1cs, &pk, r1csInput)
		}
		duration := time.Since(start)
		if fCount > 1 {
			duration = time.Duration(int64(duration) / int64(fCount))
		}

		// default proof path
		proofPath := filepath.Join(".", circuitName+".proof")
		if fProofPath != "" {
			proofPath = fProofPath
		}

		if err := gob.Write(proofPath, proof, curveID); err != nil {
			fmt.Println("error:", err)
			os.Exit(-1)
		}

		fmt.Printf("%-30s %-30s %-30s\n", "generated proof", proofPath, duration)
	case gurvy.BN256:
		if err := gob.Read(circuitPath, &bigIntR1cs, curveID); err != nil {
			fmt.Println("error:", err)
			os.Exit(-1)
		}
		r1cs := backend_bn256.Cast(&bigIntR1cs)
		fmt.Printf("%-30s %-30s %-d constraints\n", "loaded circuit", circuitPath, r1cs.NbConstraints)
		// run setup
		var pk groth16_bn256.ProvingKey
		if err := gob.Read(fPkPath, &pk, curveID); err != nil {
			fmt.Println("can't load proving key")
			fmt.Println(err)
			os.Exit(-1)
		}
		fmt.Printf("%-30s %-30s\n", "loaded proving key", fPkPath)

		// parse input file
		r1csInput := backend.NewAssignment()
		err = r1csInput.ReadFile(fInputPath)
		if err != nil {
			fmt.Println("can't parse input", err)
			os.Exit(-1)
		}
		fmt.Printf("%-30s %-30s %-d inputs\n", "loaded input", fInputPath, len(r1csInput))

		// compute proof
		start := time.Now()
		proof, err := groth16_bn256.Prove(&r1cs, &pk, r1csInput)
		if err != nil {
			fmt.Println("Error proof generation", err)
			os.Exit(-1)
		}
		for i := uint(1); i < fCount; i++ {
			_, _ = groth16_bn256.Prove(&r1cs, &pk, r1csInput)
		}
		duration := time.Since(start)
		if fCount > 1 {
			duration = time.Duration(int64(duration) / int64(fCount))
		}

		// default proof path
		proofPath := filepath.Join(".", circuitName+".proof")
		if fProofPath != "" {
			proofPath = fProofPath
		}

		if err := gob.Write(proofPath, proof, curveID); err != nil {
			fmt.Println("error:", err)
			os.Exit(-1)
		}

		fmt.Printf("%-30s %-30s %-30s\n", "generated proof", proofPath, duration)
	default:
		fmt.Println("error:", errUnknownCurve)
		os.Exit(-1)
	}

}
