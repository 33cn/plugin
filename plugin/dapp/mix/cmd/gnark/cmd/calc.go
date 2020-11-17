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

	"github.com/spf13/cobra"

	"os"

	mimcbn256 "github.com/consensys/gnark/crypto/hash/mimc/bn256"
	fr_bn256 "github.com/consensys/gurvy/bn256/fr"
)

// verifyCmd represents the verify command
var calcCmd = &cobra.Command{
	Use:     "calc",
	Short:   "calc",
	Version: Version,
}

var hashCmd = &cobra.Command{
	Use:     "hash  ...",
	Short:   "read strings to calc hash",
	Run:     hash,
	Version: Version,
}

func init() {
	calcCmd.AddCommand(hashCmd)
	rootCmd.AddCommand(calcCmd)

}

func hash(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("missing input strings")
		os.Exit(-1)
	}

	var sum []byte
	for _, k := range args {
		fmt.Println("input:", k)
		sum = append(sum, getByte(k)...)
	}
	hash := mimcbn256.Sum("seed", sum)
	fmt.Println("hash=", getFrString(hash))

}
func getByte(v string) []byte {
	var fr fr_bn256.Element
	fr.SetString(v)
	return fr.Bytes()
}
func getFrString(v []byte) string {
	var f fr_bn256.Element
	f.SetBytes(v)
	return f.String()
}
