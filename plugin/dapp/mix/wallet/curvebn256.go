package wallet

import (
	mixtypes "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gurvy/bn256/fr"
	"github.com/consensys/gurvy/bn256/twistededwards"
)

type curveBn256ECDH struct {
}

// NewCurve25519ECDH creates a new ECDH instance that uses djb's curve25519
// elliptical curve.
func NewCurveBn256ECDH() ECDH {
	return &curveBn256ECDH{}
}

func (e *curveBn256ECDH) GenerateKey(generator []byte) (*mixtypes.PrivKey, *mixtypes.PubKey) {
	var sk fr.Element
	if len(generator) <= 0 {
		sk.SetRandom()
	} else {
		sk.SetBytes(generator)
	}

	ed := twistededwards.GetEdwardsCurve()

	var point twistededwards.Point
	point.ScalarMul(&ed.Base, sk)

	priv := &mixtypes.PrivKey{
		Data: sk.String(),
	}
	pub := &mixtypes.PubKey{
		X: point.X.String(),
		Y: point.Y.String(),
	}

	return priv, pub
}

func (e *curveBn256ECDH) GenerateSharedSecret(priv *mixtypes.PrivKey, pub *mixtypes.PubKey) ([]byte, error) {

	var point, pubPoint twistededwards.Point

	pubPoint.X.SetString(pub.X)
	pubPoint.Y.SetString(pub.Y)

	var frPriv fr.Element
	frPriv.SetString(priv.Data)
	point.ScalarMul(&pubPoint, frPriv)

	return point.X.Bytes(), nil
}
