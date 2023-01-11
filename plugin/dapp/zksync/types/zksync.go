package types

import (
	"encoding/json"
	"reflect"

	"github.com/33cn/chain33/common/log/log15"

	"github.com/33cn/chain33/types"
)

var ztlog = log15.New("module", Zksync)

/*
 * 交易相关类型定义
 * 交易action通常有对应的log结构，用于交易回执日志记录
 * 每一种action和log需要用id数值和name名称加以区分
 */

// action类型id和name，这些常量可以自定义修改
const (
	TyNoopAction           = 0
	TyDepositAction        = 1  //eth存款
	TyWithdrawAction       = 2  //eth取款
	TyTransferAction       = 3  //转账
	TyTransferToNewAction  = 4  //向新地址转账
	TyProxyExitAction      = 5  //代理退出
	TySetPubKeyAction      = 6  //设置公钥
	TyFullExitAction       = 7  //从L1完全退出
	TySwapAction           = 8  //交换
	TyContractToTreeAction = 9  //合约账户转入叶子
	TyTreeToContractAction = 10 //叶子账户转入合约
	TyFeeAction            = 11 //手续费
	TyMintNFTAction        = 12
	TyWithdrawNFTAction    = 13
	TyTransferNFTAction    = 14

	//纯特殊电路类型，非Zksync合约使用的action
	TyContractToTreeNewAction = 30 //合约账户转入新的叶子

	//非电路action
	TySetVerifyKeyAction   = 102 //设置电路验证key
	TyCommitProofAction    = 103 //提交zk proof
	TySetVerifierAction    = 104 //设置验证者
	TySetFeeAction         = 105 //设置手续费
	TySetTokenSymbolAction = 106 //设置token的symbol 以方便在合约使用
	TySetExodusModeAction  = 107 //设置token的symbol 以方便在合约使用

	TyAssetTransferAction       = 120 //从tree转到zksync合约的资产账户之间转账
	TyAssetTransferToExecAction = 121 //从tree转到zksync合约的资产转到执行器
	TyAssetWithdrawAction       = 122 //从执行器提款到zksync合约账户

	NameNoopAction           = "Noop"
	NameDepositAction        = "Deposit"
	NameWithdrawAction       = "ZkWithdraw"
	NameContractToTreeAction = "ContractToTree"
	NameTreeToContractAction = "TreeToContract"
	NameTransferAction       = "ZkTransfer"
	NameTransferToNewAction  = "TransferToNew"
	NameForceExitAction      = "ProxyExit"
	NameSetPubKeyAction      = "SetPubKey"
	NameFullExitAction       = "FullExit"
	NameSwapAction           = "Swap"
	NameFeeAction            = "Fee"
	NameMintNFTAction        = "MintNFT"
	NameWithdrawNFTACTION    = "WithdrawNFT"
	NameTransferNFTAction    = "TransferNFT"

	NameContractToTreeNewAction = "ContractToTreeNew"

	NameSetVerifyKeyAction   = "SetVerifyKey"
	NameCommitProofAction    = "CommitProof"
	NameSetVerifierAction    = "SetVerifier"
	NameSetFeeAction         = "SetFee"
	NameSetTokenSymbolAction = "SetTokenSymbol"
	NameSetExodusMode        = "SetExodusMode"

	NameAssetTransfer      = "Transfer"
	NameAssetTransfer2Exec = "TransferToExec"
	NameAssetWithdraw      = "Withdraw"
)

// log类型id值
const (
	TyNoopLog           = 100
	TyDepositLog        = 101 //存款
	TyWithdrawLog       = 102 //取款
	TyTransferLog       = 103 //转账
	TyTransferToNewLog  = 104 //向新地址转账
	TyProxyExitLog      = 105 //强制退出
	TySetPubKeyLog      = 106 //设置公钥
	TyFullExitLog       = 107 //从L1完全退出
	TySwapLog           = 108 //交换
	TyContractToTreeLog = 109 //合约账户转入叶子
	TyTreeToContractLog = 110 //叶子账户转入合约
	TyFeeLog            = 111 //手续费
	TyMintNFTLog        = 112 //铸造NFT
	TyWithdrawNFTLog    = 113 //L2提款NFT到L1
	TyTransferNFTLog    = 114 //L2提款NFT到L1
	TyMintNFT2SystemLog = 115 //向系统账户铸造NFT,且其token ID为全局nft token id，因为其余额设置为token hash,所以使用不同的log标志

	TySetVerifyKeyLog          = 202 //设置电路验证key
	TyCommitProofLog           = 203 //提交zk proof
	TySetVerifierLog           = 204 //设置验证者
	TySetL1PriorityId          = 205 //设置 l1 priority id;
	TySetFeeLog                = 206
	TyCommitProofRecordLog     = 207 //提交zk proof
	TyLogContractAssetDeposit  = 208 //tree资产存储到contract
	TyLogContractAssetWithdraw = 209 //contract 资产withdraw到tree
	TyLogSetTokenSymbol        = 210 //设置电路验证key
	TyLogSetExodusMode         = 211 //系统设置exodus mode
	TySetL2OpQueueIdLog        = 212 //设置 L2 上的 operation 到queue
	TySetL2OpFirstQueueIdLog   = 213 //设置 L2 上的 first queue id
	TySetL2OpLastQueueIdLog    = 214 //设置 L2 上的 last queue id
	TySetProofId2QueueIdLog    = 215 //设置 proofId的pubdata的最后一个op 对应的最后一个queue id
	TyDepositRollbackLog       = 216 //deposit 回滚的log
	TyWithdrawRollbackLog      = 217 //withdraw 回滚的log
	TyPriority2QueIdLog        = 218 //priority id to queueId
)

const (
	ListDESC = int32(0)
	ListASC  = int32(1)
	ListSeek = int32(2)

	Add = int32(0)
	Sub = int32(1)

	MaxDecimalAllow = 18
	MinDecimalAllow = 4
)

//Zksync 执行器名称定义
const Zksync = "zksync"
const ZkManagerKey = "manager"
const ZkMimcHashSeed = "seed"
const ZkVerifierKey = "verifier"

//配置的系统收交易费账户
const ZkCfgEthFeeAddr = "ethFeeAddr"
const ZkCfgLayer2FeeAddr = "layer2FeeAddr"

//msg宽度
const (
	TxTypeBitWidth    = 8  //1byte
	AccountBitWidth   = 32 //4byte
	TokenBitWidth     = 32 //4byte for support NFT id
	NFTAmountBitWidth = 16
	AmountBitWidth    = 128 //16byte
	AddrBitWidth      = 160 //20byte
	HashBitWidth      = 256 //32byte
	PubKeyBitWidth    = 256 //32byte

	PacAmountManBitWidth = 35 //amount mantissa part, 比如12340000,只取1234部分，0000用exponent表示
	PacFeeManBitWidth    = 11 //fee mantissa part
	PacExpBitWidth       = 5  //amount and fee exponent part,支持31个0
	MaxExponentVal       = 32 // 2**5 by exp bit width

	ChunkBitWidth = 224               //one chunk 16 bytes
	ChunkBytes    = ChunkBitWidth / 8 //28 bytes
)

const (
	//BN254Fp=254bit,254-2 bit
	MsgFirstWidth  = 252
	MsgSecondWidth = 252
	MsgThirdWidth  = 248
	MsgWidth       = 752 //94 byte

)

//不同type chunk数量
const (
	DepositChunks          = 3
	Contract2TreeChunks    = 2
	Contract2TreeNewChunks = 3
	Tree2ContractChunks    = 2
	TransferChunks         = 2
	Transfer2NewChunks     = 3
	WithdrawChunks         = 2
	ProxyExitChunks        = 2
	FullExitChunks         = 2
	SwapChunks             = 4
	NoopChunks             = 1
	SetPubKeyChunks        = 3
	FeeChunks              = 1
	//MintNFTChunks, withrawNFT, transferNft, NFT chunks 不只是看pubdata长度，更要看需要几个chunk完成，这里chunks超出了pubdata的长度
	MintNFTChunks     = 5
	WithdrawNFTChunks = 6
	TransferNFTChunks = 3
)

const (
	//SystemDefaultAcctId 缺省备用账户
	SystemDefaultAcctId = 0
	//SystemFeeAccountId 此账户作为缺省收费账户
	SystemFeeAccountId = 1
	//SystemNFTAccountId 此特殊账户没有私钥，只记录并产生NFT token资产，不会有小于NFTTokenId的FT token记录
	SystemNFTAccountId = 2
	//SystemTree2ContractAcctId, 汇总从 tree2contract 跨链的资产总额
	SystemTree2ContractAcctId = 3
	//SystemNFTTokenId 作为一个NFT token标记 低于NFTTokenId 为FT token id, 高于NFTTokenId为 NFT token id，即从NFTTokenId+1开始作为NFT资产
	SystemNFTTokenId = 256 //2^8,

)

//ERC protocol
const (
	ZKERC1155 = 1
	ZKERC721  = 2
)

const (
	NormalProxyPubKey = 1
	SystemProxyPubKey = 2
	SuperProxyPubKey  = 3
)

const (
	InitMode        = 0
	NormalMode      = 1 //从pause模式可以恢复为normal mode
	PauseMode       = 2 //暂停模式，管理员在监测到存款异常后可以设置暂停模式，暂停所有操作，以防止错误的deposit转账，检查正常可以恢复
	ExodusMode      = 3 //逃生舱预备阶段，停止所有除contract2tree外的操作，不可以恢复到Normal
	ExodusFinalMode = 4 //逃生舱最终阶段 回滚last successed proof后的deposit和withdraw操作，统计balance gap信息，收敛最终treeRoot,保证尽快退出资产到L1
)

const (
	ModeValNo  = 0 //
	ModeValYes = 1 //

)

var (
	//定义actionMap
	actionMap = map[string]int32{
		//NameNoopAction:           TyNoopAction,
		NameDepositAction:        TyDepositAction,
		NameWithdrawAction:       TyWithdrawAction,
		NameContractToTreeAction: TyContractToTreeAction,
		NameTreeToContractAction: TyTreeToContractAction,
		NameTransferAction:       TyTransferAction,
		NameTransferToNewAction:  TyTransferToNewAction,
		NameForceExitAction:      TyProxyExitAction,
		NameSetPubKeyAction:      TySetPubKeyAction,
		NameFullExitAction:       TyFullExitAction,
		NameSwapAction:           TySwapAction,
		NameSetVerifyKeyAction:   TySetVerifyKeyAction,
		NameCommitProofAction:    TyCommitProofAction,
		NameSetVerifierAction:    TySetVerifierAction,
		NameSetFeeAction:         TySetFeeAction,
		NameMintNFTAction:        TyMintNFTAction,
		NameWithdrawNFTACTION:    TyWithdrawNFTAction,
		NameTransferNFTAction:    TyTransferNFTAction,
		NameSetTokenSymbolAction: TySetTokenSymbolAction,
		NameSetExodusMode:        TySetExodusModeAction,
		NameAssetTransfer:        TyAssetTransferAction,
		NameAssetTransfer2Exec:   TyAssetTransferToExecAction,
		NameAssetWithdraw:        TyAssetWithdrawAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		//TyNoopLog:           {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TyNoopLog"},
		TyDepositLog:               {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyDepositLog"},
		TyWithdrawLog:              {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyWithdrawLog"},
		TyContractToTreeLog:        {Ty: reflect.TypeOf(TransferReceipt4L2{}), Name: "TyContractToTreeLog"},
		TyTreeToContractLog:        {Ty: reflect.TypeOf(TransferReceipt4L2{}), Name: "TyTreeToContractLog"},
		TyTransferLog:              {Ty: reflect.TypeOf(TransferReceipt4L2{}), Name: "TyTransferLog"},
		TyTransferToNewLog:         {Ty: reflect.TypeOf(TransferReceipt4L2{}), Name: "TyTransferToNewLog"},
		TyProxyExitLog:             {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyForceExitLog"},
		TySetPubKeyLog:             {Ty: reflect.TypeOf(SetPubKeyReceipt{}), Name: "TySetPubKeyLog"},
		TyFullExitLog:              {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyFullExitLog"},
		TySwapLog:                  {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TySwapLog"},
		TyFeeLog:                   {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyFeeLog"},
		TyMintNFTLog:               {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyMintNFTLog"},
		TyWithdrawNFTLog:           {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyWithdrawNFTLog"},
		TyTransferNFTLog:           {Ty: reflect.TypeOf(TransferReceipt4L2{}), Name: "TyTransferNFTLog"},
		TySetVerifyKeyLog:          {Ty: reflect.TypeOf(ReceiptSetVerifyKey{}), Name: "TySetVerifyKey"},
		TyCommitProofLog:           {Ty: reflect.TypeOf(ReceiptCommitProof{}), Name: "TyCommitProof"},
		TySetVerifierLog:           {Ty: reflect.TypeOf(ReceiptSetVerifier{}), Name: "TySetVerifierLog"},
		TySetL1PriorityId:          {Ty: reflect.TypeOf(ReceiptL1PriorityID{}), Name: "TySetL1PriorityId"},
		TySetFeeLog:                {Ty: reflect.TypeOf(ReceiptSetFee{}), Name: "TySetFeeLog"},
		TyCommitProofRecordLog:     {Ty: reflect.TypeOf(ReceiptCommitProofRecord{}), Name: "TyCommitProofRecordLog"},
		TyLogSetTokenSymbol:        {Ty: reflect.TypeOf(ReceiptSetTokenSymbol{}), Name: "TySetTokenSymbolLog"},
		TyLogContractAssetWithdraw: {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogContractAssetWithdraw"},
		TyLogContractAssetDeposit:  {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogContractAssetDeposit"},
		TyLogSetExodusMode:         {Ty: reflect.TypeOf(ReceiptExodusMode{}), Name: "TySetExodusModeLog"},
		TySetL2OpQueueIdLog:        {Ty: reflect.TypeOf(ReceiptL2QueueIDData{}), Name: "TySetL2QueueIdLog"},
		TySetL2OpFirstQueueIdLog:   {Ty: reflect.TypeOf(ReceiptL2FirstQueueID{}), Name: "TySetL2FirstQueueIdLog"},
		TySetL2OpLastQueueIdLog:    {Ty: reflect.TypeOf(ReceiptL2LastQueueID{}), Name: "TySetL2LastQueueIdLog"},
		TySetProofId2QueueIdLog:    {Ty: reflect.TypeOf(ReceiptProofId2QueueIDData{}), Name: "TySetProofId2QueueIdLog"},
		TyDepositRollbackLog:       {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyDepositRollbackLog"},
		TyWithdrawRollbackLog:      {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyWithdrawRollbackLog"},
		TyPriority2QueIdLog:        {Ty: reflect.TypeOf(Priority2QueueId{}), Name: "TyPriority2QueueIdLog"},
	}

	FeeMap = map[int64]string{
		TyWithdrawAction:       "1000000",
		TyTransferAction:       "100000",
		TyTransferToNewAction:  "100000",
		TyProxyExitAction:      "1000000",
		TyFullExitAction:       "1000000",
		TySwapAction:           "100000",
		TyContractToTreeAction: "10000",
		TyTreeToContractAction: "10000",
		TyMintNFTAction:        "100",
		TyWithdrawNFTAction:    "100",
		TyTransferNFTAction:    "100",
	}
)

// init defines a register function
func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(Zksync))
	//注册合约启用高度
	types.RegFork(Zksync, InitFork)
	types.RegExec(Zksync, InitExecutor)
}

// InitFork defines register fork
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(Zksync, "Enable", 0)
}

// InitExecutor defines register executor
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(Zksync, NewType(cfg))
}

//ZksyncType ...
type ZksyncType struct {
	types.ExecTypeBase
}

//NewType ...
func NewType(cfg *types.Chain33Config) *ZksyncType {
	c := &ZksyncType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取合约action结构
func (e *ZksyncType) GetPayload() types.Message {
	return &ZksyncAction{}
}

// GetTypeMap 获取合约action的id和name信息
func (e *ZksyncType) GetTypeMap() map[string]int32 {
	return actionMap
}

// GetLogMap 获取合约log相关信息
func (e *ZksyncType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}

// CreateTx zksync 创建交易，系统构造缺省to地址为合约地址，对于transfer/transferToExec 的to地址需要特殊处理为paylaod的to地址，
func (e *ZksyncType) CreateTx(action string, msg json.RawMessage) (*types.Transaction, error) {
	tx, err := e.ExecTypeBase.CreateTx(action, msg)
	if err != nil {
		ztlog.Error("zksync CreateTx failed", "err", err, "action", action, "msg", string(msg))
		return nil, err
	}
	cfg := e.GetConfig()
	if !cfg.IsPara() {
		var transfer ZksyncAction
		err = types.Decode(tx.Payload, &transfer)
		if err != nil {
			ztlog.Error("zksync CreateTx failed", "decode payload err", err, "action", action, "msg", string(msg))
			return nil, err
		}
		if action == "Transfer" {
			tx.To = transfer.GetTransfer().To
		} else if action == "Withdraw" {
			tx.To = transfer.GetWithdraw().To
		} else if action == "TransferToExec" {
			tx.To = transfer.GetTransferToExec().To
		}
	}
	return tx, nil
}
