package wallet

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/33cn/plugin/plugin/crypto/legacymimc"
	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

// TestGenerateKeyCompat 验证兼容派生与 gnark-crypto v0.5.3 输出一致。
// 期望值通过 v0.5.3 的 eddsa.GenerateKey 复现得到（复现时直接以 acc1privkey
// 原始字节作为 seed，与 zksync 测试 SignTransaction 的行为一致）。
func TestGenerateKeyCompat(t *testing.T) {
	privHex := "19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
	privRaw, err := hex.DecodeString(privHex)
	require.NoError(t, err)

	priv, err := GenerateKeyCompat(bytes.NewReader(privRaw))
	require.NoError(t, err)

	// v0.5.3 派生期望值
	require.Equal(t, "12626172696580665941222283729733634448136278912784440589993887918763909125082",
		priv.PublicKey.A.X.String(), "pubkey.X 应与 v0.5.3 一致")
	require.Equal(t, "3930539417016545586650485084236670648989777351201796772004849233730099400354",
		priv.PublicKey.A.Y.String(), "pubkey.Y 应与 v0.5.3 一致")

	// 关键：mimc(pubkey.X || pubkey.Y) 必须等于 executor SetPubKey 校验使用的
	// 历史 leaf.Chain33Addr（测试硬编码值），即历史地址可继续通过校验
	h := legacymimc.NewMiMC(zksyncTypes.ZkMimcHashSeed)
	h.Write(zksyncTypes.Str2Byte(priv.PublicKey.A.X.String()))
	h.Write(zksyncTypes.Str2Byte(priv.PublicKey.A.Y.String()))
	calc := zksyncTypes.Byte2Str(h.Sum(nil))
	require.Equal(t, "19694183066356799104974294716313078444659172842638956126168373945465009608401",
		calc, "mimc(pubkey) 应匹配历史 Chain33Addr")
}

// TestGenerateKeyCompatDeterministic 验证同一 seed 派生稳定
func TestGenerateKeyCompatDeterministic(t *testing.T) {
	privHex := "19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
	privRaw, _ := hex.DecodeString(privHex)

	priv1, err := GenerateKeyCompat(bytes.NewReader(privRaw))
	require.NoError(t, err)
	priv2, err := GenerateKeyCompat(bytes.NewReader(privRaw))
	require.NoError(t, err)
	require.Equal(t, priv1.PublicKey.A.X.String(), priv2.PublicKey.A.X.String())
	require.Equal(t, priv1.PublicKey.A.Y.String(), priv2.PublicKey.A.Y.String())
}

// TestGenerateKeyCompatSignVerify 验证兼容派生的密钥能用 v0.12.1 的 eddsa 正常签名/验证
func TestGenerateKeyCompatSignVerify(t *testing.T) {
	privHex := "19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
	privRaw, _ := hex.DecodeString(privHex)

	priv, err := GenerateKeyCompat(bytes.NewReader(privRaw))
	require.NoError(t, err)

	msg := []byte("test message for eddsa sign")
	sig, err := priv.Sign(msg, legacymimc.NewMiMC(zksyncTypes.ZkMimcHashSeed))
	require.NoError(t, err)
	ok, err := priv.PublicKey.Verify(sig, msg, legacymimc.NewMiMC(zksyncTypes.ZkMimcHashSeed))
	require.NoError(t, err)
	require.True(t, ok)
}
