package executor

import (
	"sync"

	"github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
)

/*
 * 执行器相关定义
 * 重载基类相关接口
 */

type config struct {
	CommitAddress        string `json:"commitAddress"`
	BtcNetName           string `json:"btcNetName"`
	AllowRegtestTimeWarp bool   `json:"allowRegtestTimeWarp"`
}

var (
	//日志
	elog        = log.New("module", "lightclient.exec")
	lightCfg    config
	cfgInitOnce sync.Once
)

var driverName = ltypes.LightclientX

// Init register dapp
func Init(_ string, cfg *types.Chain33Config, sub []byte) {
	initCfg(sub)
	drivers.Register(cfg, GetName(), newLightclient, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

func initCfg(sub []byte) {

	cfgInitOnce.Do(func() {
		types.MustDecode(sub, &lightCfg)
		if lightCfg.BtcNetName == "" {
			lightCfg.BtcNetName = "mainnet"
		}
	})
}

// InitExecType Init Exec Type
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&lightclient{}))
}

type lightclient struct {
	drivers.DriverBase
}

func newLightclient() drivers.Driver {
	t := &lightclient{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetName get driver name
func GetName() string {
	return newLightclient().GetName()
}

func (l *lightclient) GetDriverName() string {
	return driverName
}

// ExecutorOrder 设置localdb的EnableRead
func (l *lightclient) ExecutorOrder() int64 {
	return drivers.ExecLocalSameTime
}

func readDB(kdb db.KV, key []byte, result types.Message) error {

	val, err := kdb.Get(key)
	if err != nil {
		return err
	}
	return types.Decode(val, result)
}

func getBtcLastHeader(sdb db.KV) (*ltypes.BtcHeader, error) {

	header := &ltypes.BtcHeader{}
	err := readDB(sdb, btcLastHeaderKey(), header)
	if err == types.ErrNotFound {
		return &ltypes.BtcHeader{}, nil
	}
	return header, err
}

func getBtcHeader(ldb db.KV, height uint64) (*ltypes.BtcHeader, error) {

	header := &ltypes.BtcHeader{}
	err := readDB(ldb, btcHeaderKey(height), header)
	return header, err
}

func getBtcHeight(ldb db.KV, hash string) (*types.Int64, error) {

	height := &types.Int64{}
	err := readDB(ldb, btcHeaderHashHeightKey(hash), height)
	return height, err
}
