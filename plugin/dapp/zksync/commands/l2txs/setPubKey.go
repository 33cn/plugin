package l2txs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
)

func setManyPubKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pubkey_many",
		Short: "set pubkeys",
		Run:   setPubKey,
	}
	setPubKeyFlag(cmd)
	return cmd
}

func setPubKeyFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("keys", "k", "", "private keys, use ',' separate")
	_ = cmd.MarkFlagRequired("keys")
	cmd.Flags().StringP("accountIDs", "a", "0", "L2 account ids on chain33, use ',' separate")
	_ = cmd.MarkFlagRequired("accountIDs")
	cmd.Flags().Uint64P("pubkeyT", "t", 0, "self default:0, proxy pubkey ty: 1: normal,2:system,3:super")
	cmd.Flags().StringP("pubkeyX", "x", "", "proxy pubkey x value")
	cmd.Flags().StringP("pubkeyY", "y", "", "proxy pubkey y value")
}

func setPubKey(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkeyT, _ := cmd.Flags().GetUint64("pubkeyT")
	pubkeyX, _ := cmd.Flags().GetString("pubkeyX")
	pubkeyY, _ := cmd.Flags().GetString("pubkeyY")
	privateKeys, _ := cmd.Flags().GetString("keys")
	accountIDs, _ := cmd.Flags().GetString("accountIDs")
	paraName, _ := cmd.Flags().GetString("paraName")
	ids := strings.Split(accountIDs, ",")
	keys := strings.Split(privateKeys, ",")

	if pubkeyT > 0 && (len(pubkeyX) == 0 || len(pubkeyY) == 0) {
		_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("set proxy pubkey, need set pubkeyX pubkeyY"))
	}

	if len(ids) != len(keys) {
		fmt.Println("err len(ids) != len(keys)", len(ids), "!=", len(keys))
		return
	}

	for i := 0; i < len(ids); i++ {
		id, _ := strconv.ParseInt(ids[i], 10, 64)
		param := &zksyncTypes.ZkSetPubKey{
			AccountId: uint64(id),
			PubKeyTy:  pubkeyT,
			PubKey: &zksyncTypes.ZkPubKey{
				X: pubkeyX,
				Y: pubkeyY,
			},
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TySetPubKeyAction,
			Value: &zksyncTypes.ZksyncAction_SetPubKey{
				SetPubKey: param,
			},
		}

		tx, err := createChain33Tx(keys[i], getRealExecName(paraName, zksyncTypes.Zksync), action)
		if nil != err {
			fmt.Println("SetPubKey failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}
