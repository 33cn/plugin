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
	"github.com/consensys/gurvy/bn256/twistededwards"
)

// verifyCmd represents the verify command
var calcCmd = &cobra.Command{
	Use:     "calc",
	Short:   "calc",
	Version: Version,
}

var (
	fCommit string
	fRandom string
)

func init() {
	calcCmd.AddCommand(hashCmd)
	calcCmd.AddCommand(commitCmd)

	rootCmd.AddCommand(calcCmd)

	commitCmd.PersistentFlags().StringVar(&fCommit, "value", "", "specifies commit value")
	commitCmd.PersistentFlags().StringVar(&fRandom, "random", "", "specifies random value")
	_ = commitCmd.MarkPersistentFlagRequired("value")
	_ = commitCmd.MarkPersistentFlagRequired("random")

}

var hashCmd = &cobra.Command{
	Use:     "hash  ...",
	Short:   "read strings to calc hash",
	Run:     hash,
	Version: Version,
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

var commitCmd = &cobra.Command{
	Use:     "commit",
	Short:   "commit value",
	Run:     commit,
	Version: Version,
}

func commit(cmd *cobra.Command, args []string) {
	//if len(args) < 1 {
	//	fmt.Println("missing input strings")
	//	os.Exit(-1)
	//}

	var basex, basey, baseHx, baseHy fr_bn256.Element
	basex.SetString("5299619240641551281634865583518297030282874472190772894086521144482721001553")
	basey.SetString("16950150798460657717958625567821834550301663161624707787222815936182638968203")

	baseHx.SetString("10190477835300927557649934238820360529458681672073866116232821892325659279502")
	baseHy.SetString("7969140283216448215269095418467361784159407896899334866715345504515077887397")

	basePoint := twistededwards.NewPoint(basex, basey)
	baseHPoint := twistededwards.NewPoint(baseHx, baseHy)

	var frCommit, frRandom fr_bn256.Element
	frCommit.SetString(fCommit).FromMont()
	frRandom.SetString(fRandom).FromMont()

	fmt.Println("commit", fCommit, "random", fRandom)
	var commitPoint, randomPoint, finalPoint twistededwards.Point

	commitPoint.ScalarMul(&basePoint, frCommit)
	randomPoint.ScalarMul(&baseHPoint, frRandom)

	finalPoint.Add(&commitPoint, &randomPoint)

	fmt.Println("finalPoint X:", finalPoint.X.String())
	fmt.Println("finalPoint Y:", finalPoint.Y.String())

	//
	//fmt.Println("commitX:",commitPoint.X.String())
	//fmt.Println("commitY:",commitPoint.Y.String())
	//
	//fmt.Println("randomX:",randomPoint.X.String())
	//fmt.Println("randomY:",randomPoint.Y.String())

}
