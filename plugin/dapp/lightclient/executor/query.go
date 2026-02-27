package executor

import (
	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
)

func (l *lightclient) Query_GetBtcLastHeader(req *types.ReqNil) (types.Message, error) {

	header, err := getBtcLastHeader(l.GetStateDB())
	return header, err
}

func (l *lightclient) Query_GetBtcHeader(req *ltypes.ReqGetBtcHeader) (types.Message, error) {

	header, err := getBtcHeader(l.GetLocalDB(), req.GetHeight())
	return header, err
}

func (l *lightclient) Query_GetBtcHeaderByHash(req *types.ReqString) (types.Message, error) {

	height, err := getBtcHeight(l.GetLocalDB(), req.GetData())
	if err != nil {
		elog.Error("Query_GetBtcHeaderByHash", "hash", req.GetData(), "err", err)
		return nil, err
	}
	header, err := getBtcHeader(l.GetLocalDB(), uint64(height.GetData()))
	return header, err
}
