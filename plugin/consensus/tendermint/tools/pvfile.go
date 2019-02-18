package main

import (
	"fmt"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/tendermint/types"
	// "github.com/inconshreveable/log15"
	"math/rand"
	"strconv"
	"time"
)

var (
	// tendermintlog = log15.New("module", "tendermint")
	strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
	genFile  = "genesis.json"
	pvFile   = "priv_validator_"
)

func RandStr(length int) string {
	chars := []byte{}
MAIN_LOOP:
	for {
		val := rand.Int63()
		for i := 0; i < 10; i++ {
			v := int(val & 0x3f) // rightmost 6 bits
			if v >= 62 {         // only 62 characters in strChars
				val >>= 6
				continue
			} else {
				chars = append(chars, strChars[v])
				if len(chars) == length {
					break MAIN_LOOP
				}
				val >>= 6
			}
		}
	}

	return string(chars)
}

func initCryptoImpl() error {
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		// tendermintlog.Error("New crypto impl failed", "err", err)
		return err
	}
	ttypes.ConsensusCrypto = cr
	return nil
}

func createFiles(num int) {
	// init crypto instance
	err := initCryptoImpl()
	if err != nil {
		return
	}

	// genesis file
	genDoc := ttypes.GenesisDoc{
		ChainID:     ttypes.Fmt("chain33-%v", RandStr(6)),
		GenesisTime: time.Now(),
	}

	for i := 0; i < num; i++ {
		// create private validator filegen
		pvFileName := pvFile + strconv.Itoa(i) + ".json"
		privValidator := ttypes.LoadOrGenPrivValidatorFS(pvFileName)
		if privValidator == nil {
			// tendermintlog.Error("Create priv_validator file failed.")
			break
		}

		// create genesis validator by the pubkey of private validator
		gv := ttypes.GenesisValidator{
			PubKey: ttypes.KeyText{"ed25519", privValidator.GetPubKey().KeyString()},
			Power:  10,
		}
		genDoc.Validators = append(genDoc.Validators, gv)
	}

	if err := genDoc.SaveAs(genFile); err != nil {
		// tendermintlog.Error("Generated genesis file failed.")
		return
	}
	// tendermintlog.Info("Generated genesis file", "path", genFile)

	return
}

func main() {
	var num int
	fmt.Printf("Please enter the number of key file:")
	fmt.Scan(&num)
	createFiles(num)
}
