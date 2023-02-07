package executor

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
)

func getChain33Addr(privateKeyString string) string {
	privateKeyBytes, err := hex.DecodeString(privateKeyString)
	if err != nil {
		panic(err)
	}
	privateKey, err := eddsa.GenerateKey(bytes.NewReader(privateKeyBytes))
	if err != nil {
		panic(err)
	}
	hash := mimc.NewMiMC(zt.ZkMimcHashSeed)
	hash.Write(zt.Str2Byte(privateKey.PublicKey.A.X.String()))
	hash.Write(zt.Str2Byte(privateKey.PublicKey.A.Y.String()))
	return hex.EncodeToString(hash.Sum(nil))
}

func TestAccountHash(t *testing.T) {
	var leaf zt.Leaf
	leaf.AccountId = 1
	leaf.EthAddress = "980818135352849559554652468538757099471386586455"
	leaf.Chain33Addr = "3415326846406104843498339737738292353412449296387254161761470177873504232418"

	leaf.TokenHash = zt.Str2Byte("14633446003514262524099709640745596521508648778482661942408784061885334136010")
	var pubkey zt.ZkPubKey
	pubkey.X = "110829526890202442231796950896186450339098004198300292113013256946470504791"
	pubkey.Y = "12207062062295480868601430817261127111444831355336859496235449885847711361351"
	//leaf.PubKey = &pubkey
	mimcHash := mimc.NewMiMC(zt.ZkMimcHashSeed)
	hash := getLeafHash(mimcHash, &leaf)
	var f fr.Element
	f.SetBytes(hash)
	t.Log("hash", f.String())
}

func TestInitRoot(t *testing.T) {
	ethFeeAddr := "832367164346888E248bd58b9A5f480299F1e88d"
	chain33FeeAddr := "2c4a5c378be2424fa7585320630eceba764833f1ec1ffb2fafc1af97f27baf5a"

	ethFeeAddrDecimal, _ := zt.HexAddr2Decimal(ethFeeAddr)
	chain33FeeAddrDecimal, _ := zt.HexAddr2Decimal(chain33FeeAddr)
	root := getInitTreeRoot(nil, ethFeeAddrDecimal, chain33FeeAddrDecimal)
	//root 10504580268631993551496725490269484542796412497472171704511488082346545793961
	t.Log("root", root)
}
