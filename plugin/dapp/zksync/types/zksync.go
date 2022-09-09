package types

import (
	"reflect"

	"github.com/33cn/chain33/types"
)

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
	TySetVerifyKeyAction        = 102 //设置电路验证key
	TyCommitProofAction         = 103 //提交zk proof
	TySetVerifierAction         = 104 //设置验证者
	TySetFeeAction              = 105 //设置手续费
	TySetTokenSymbolAction      = 106 //设置token的symbol 以方便在合约使用
	TyAssetTransferAction       = 107 //从tree转到zksync合约的资产账户之间转账
	TyAssetTransferToExecAction = 108 //从tree转到zksync合约的资产转到执行器
	TyAssetWithdrawAction       = 109 //从执行器提款到zksync合约账户

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

	NameSetVerifyKeyAction = "SetVerifyKey"
	NameCommitProofAction  = "CommitProof"
	NameSetVerifierAction  = "SetVerifier"
	NameSetFeeAction       = "SetFee"
	NameSetTokenSymbolAction = "SetTokenSymbol"
	NameAssetTransfer        = "Transfer"
	NameAssetTransfer2Exec   = "TransferToExec"
	NameAssetWithdraw        = "Withdraw"
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
	TySetEthPriorityQueueId    = 205 //设置 eth上 priority queue id;
	TySetFeeLog                = 206
	TyCommitProofRecordLog     = 207 //提交zk proof
	TyLogContractAssetDeposit  = 208 //tree资产存储到contract
	TyLogContractAssetWithdraw = 209 //contract 资产withdraw到tree
	TyLogSetTokenSymbol        = 210 //设置电路验证key
	TyLogSetExodusMode         = 211 //系统设置exodus mode
)

const (
	ListDESC = int32(0)
	ListASC  = int32(1)
	ListSeek = int32(2)

	Add = int32(0)
	Sub = int32(1)
)

//Zksync 执行器名称定义
const Zksync = "zksync"
const ZkManagerKey = "manager"
const ZkMimcHashSeed = "seed"
const ZkVerifierKey = "verifier"

//配置的系统收交易费账户
const ZkCfgEthFeeAddr = "ethFeeAddr"
const ZkCfgLayer2FeeAddr = "layer2FeeAddr"

//配置的无效交易和无效证明，用于平行链zksync交易的回滚(假设proof和eth不一致，无法fix时候)
const ZkCfgInvalidTx = "invalidTxHash"
const ZkCfgInvalidProof = "invalidProofRootHash"

//ZkParaChainInnerTitleId 平行链内部只有一个titleId，缺省为1，在主链上不同平行链有自己的titleId
const ZkParaChainInnerTitleId = "1"

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
	ExodusPrepareMode = 1 //逃生舱预备阶段  所有和L1相关的 onChain tx都不执行(deposit,withdraw,proxyexit)
	ExodusClearMode   = 2 //逃生舱清算阶段 除contract2tree外,所有L2相关的tx都不允许执行，收敛最终treeRoot,保证尽快退出资产到L1
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
		NameAssetTransfer:        TyAssetTransferAction,
		NameAssetTransfer2Exec:   TyAssetTransferToExecAction,
		NameAssetWithdraw:        TyAssetWithdrawAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		//TyNoopLog:           {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TyNoopLog"},
		TyDepositLog:            {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyDepositLog"},
		TyWithdrawLog:           {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyWithdrawLog"},
		TyContractToTreeLog:     {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyContractToTreeLog"},
		TyTreeToContractLog:     {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyTreeToContractLog"},
		TyTransferLog:           {Ty: reflect.TypeOf(TransferReceipt4L2{}), Name: "TyTransferLog"},
		TyTransferToNewLog:      {Ty: reflect.TypeOf(TransferReceipt4L2{}), Name: "TyTransferToNewLog"},
		TyProxyExitLog:          {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyForceExitLog"},
		TySetPubKeyLog:          {Ty: reflect.TypeOf(SetPubKeyReceipt{}), Name: "TySetPubKeyLog"},
		TyFullExitLog:           {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyFullExitLog"},
		TySwapLog:               {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TySwapLog"},
		TyFeeLog:                {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyFeeLog"},
		TyMintNFTLog:            {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyMintNFTLog"},
		TyWithdrawNFTLog:        {Ty: reflect.TypeOf(AccountTokenBalanceReceipt{}), Name: "TyWithdrawNFTLog"},
		TyTransferNFTLog:        {Ty: reflect.TypeOf(TransferReceipt4L2{}), Name: "TyTransferNFTLog"},
		TySetVerifyKeyLog:       {Ty: reflect.TypeOf(ReceiptSetVerifyKey{}), Name: "TySetVerifyKey"},
		TyCommitProofLog:        {Ty: reflect.TypeOf(ReceiptCommitProof{}), Name: "TyCommitProof"},
		TySetVerifierLog:        {Ty: reflect.TypeOf(ReceiptSetVerifier{}), Name: "TySetVerifierLog"},
		TySetEthPriorityQueueId: {Ty: reflect.TypeOf(ReceiptEthPriorityQueueID{}), Name: "TySetEthPriorityQueueID"},
		TySetFeeLog:             {Ty: reflect.TypeOf(ReceiptSetFee{}), Name: "TySetFeeLog"},
		TyLogSetTokenSymbol:        {Ty: reflect.TypeOf(ReceiptSetTokenSymbol{}), Name: "TySetTokenSymbolLog"},
		TyLogContractAssetWithdraw: {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogContractAssetWithdraw"},
		TyLogContractAssetDeposit:  {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogContractAssetDeposit"},
		TyLogSetExodusMode:         {Ty: reflect.TypeOf(ReceiptExodusMode{}), Name: "TySetExodusModeLog"},
	}

	FeeMap = map[int64]string{
		TyWithdrawAction:      "1000000",
		TyTransferAction:      "100000",
		TyTransferToNewAction: "100000",
		TyProxyExitAction:     "1000000",
		TyFullExitAction:      "1000000",
		TySwapAction:          "100000",
		TyMintNFTAction:       "100",
		TyWithdrawNFTAction:   "100",
		TyTransferNFTAction:   "100",
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
