package client

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

var conns sync.Map

type GRPCCli struct {
	client types.Chain33Client
}

func NewGRPCCli(grpcAddr string) *GRPCCli {
	client, err := getClient(grpcAddr)
	if err != nil {
		return nil
	}
	return &GRPCCli{client: client}
}

func getClient(target string) (types.Chain33Client, error) {
	val, ok := conns.Load(target)
	if !ok {
		conn, err := grpc.Dial(target, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		client := types.NewChain33Client(conn)
		conns.Store(target, client)
		return client, nil
	} else {
		return val.(types.Chain33Client), nil
	}
}

func (c *GRPCCli) Query(fn string, msg proto.Message) ([]byte, error) {
	ss := strings.Split(fn, ".")
	var in types.ChainExecutor
	if len(ss) == 2 {
		in.Driver = ss[0]
		in.FuncName = ss[1]
	} else {
		in.Driver = et.ExchangeX
		in.FuncName = fn
	}
	in.Param = types.Encode(msg)

	r, err := c.client.QueryChain(context.Background(), &in)
	if err != nil {
		return nil, err
	}
	if !r.IsOk {
		return nil, errors.New(string(r.Msg))
	}
	return r.Msg, nil
}

func (c *GRPCCli) Send(tx *types.Transaction, hexKey string) ([]*types.ReceiptLog, error) {
	logs, err := c.sendAndWaitReceipt(tx, hexKey)
	if err != nil {
		return nil, parseError(err)
	}
	for _, l := range logs {
		if l.Ty == types.TyLogErr {
			return nil, errors.New(string(l.Log))
		}
	}
	return logs, nil
}

// 发送交易并等待执行结果
// 如果交易非法，返回错误信息
// 如果交易执行成功，返回 交易哈希、回报
func (c *GRPCCli) sendAndWaitReceipt(tx *types.Transaction, hexKey string) (logs []*types.ReceiptLog, err error) {
	r, err := c.sendTx(tx, hexKey)
	if err != nil {
		// rpc error: code = Unknown desc = ErrNotBank
		return nil, err
	}
	if !r.IsOk {
		return nil, errors.New(string(r.Msg))
	}
	time.Sleep(time.Second)
	d, _ := c.client.QueryTransaction(context.Background(), &types.ReqHash{Hash: r.Msg})
	return d.Receipt.Logs, nil
}

//SendTx ...s
func (c *GRPCCli) sendTx(tx *types.Transaction, hexKey string) (reply *types.Reply, err error) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	tx, err = types.FormatTx(cfg, et.ExchangeX, tx)
	if err != nil {
		return nil, err
	}
	tx, err = signTx(tx, hexKey)
	if err != nil {
		return nil, err
	}

	return c.client.SendTransaction(context.Background(), tx)
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName("", signType), -1)
	if err != nil {
		return tx, err
	}

	bytes, err := common.FromHex(hexPrivKey[:])
	if err != nil {
		return tx, err
	}

	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		return tx, err
	}

	tx.Sign(int32(signType), privKey)
	return tx, nil
}

func parseError(err error) error {
	// rpc error: code = Unknown desc = ErrNotBank
	str := err.Error()
	sep := "desc = "
	i := strings.Index(str, sep)
	if i != -1 {
		return errors.New(str[i+len(sep):])
	}
	return err
}
