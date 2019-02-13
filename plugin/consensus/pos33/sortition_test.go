package pos33

import (
	"fmt"
	"math/rand"
	"time"

	// "bytes"

	"github.com/33cn/chain33/common/crypto"
	pb "github.com/33cn/chain33/types"
)

// 随机分配N个节点, 每个节点的矿机不能超过7
// 从中选出33个矿机

func init() {
	rand.Seed(time.Now().Unix())
}

var Seed = []byte("0123456789abcdefghijklmnopqrstuvwxyz")

const (
	N  = 1000
	MW = 3
)

type no struct {
	priv crypto.PrivKey
	w    int
}

func genNodes() (map[string]*no, int) {
	c, _ := crypto.New("ed25519")
	allw := 0
	nos := make(map[string]*no)
	for i := 0; i < N; i++ {
		g := 1
		if i < N*99/100 {
			g = rand.Intn(2) + 1
		} else {
			g = rand.Intn(100) + 1
		}
		allw += g
		priv, _ := c.GenKey()
		nos[string(priv.PubKey().Bytes())] = &no{priv, g}
	}
	return nos, allw
}

func doGenRands(seed []byte, nos map[string]*no, allw int) (map[string]*ty.Pos33Rands, int) {
	rands := make(map[string]*ty.Pos33Rands)
	vw := 0
	for _, n := range nos {
		rms := genRands(seed, allw, n.w, n.priv, 0)
		if rms != nil {
			//			fmt.Println("@@@@@@@@@@@@@@: ", n.w, len(rms.Rands))
			rands[string(n.priv.PubKey().Bytes())] = rms
			vw += len(rms.Rands)
		}
	}
	fmt.Println("@@@@@@@@@@@@@@@ genRands done @@@@@@@@@@@@@", len(rands), len(nos), vw)
	return rands, vw
}

// func testSortition(seed []byte, nos map[string]*no, allw int, t *testing.T) {
// 	rands, vw := doGenRands(seed, nos, allw)
// 	vw2 := 0
// 	f := func(pub string) int {
// 		w := nos[pub].w
// 		vw2 += w
// 		return w
// 	}
// 	committee, _, _ := sortition(rands, seed, allw, f, 0)
// 	fmt.Println(len(committee), allw, vw, vw2)
// 	if len(committee) == 0 {
// 		t.Fatal("fuck..........")
// 	}
// 	fmt.Println("@@@@@@@@@@@@@@@ sortition test done @@@@@@@@@@@@@")
// }

// func TestSortition(t *testing.T) {
// 	nos, allw := genNodes()
// 	for i := 0; i < 100; i++ {
// 		p := rand.Intn(13)
// 		s := string(Seed[p:p+13]) + fmt.Sprintf("%d", time.Now().Unix())
// 		seed := crypto.Sha256([]byte(s))

// 		testSortition([]byte(seed), nos, allw, t)
// 	}
// }
