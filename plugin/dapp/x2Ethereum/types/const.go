package types

var (
	ProphecyKey                         = []byte("prefix_for_Prophecy")
	Eth2Chain33Key                      = []byte("prefix_for_Eth2Chain33")
	WithdrawEthKey                      = []byte("prefix_for_WithdrawEth")
	Chain33ToEthKey                     = []byte("prefix_for_Chain33ToEth")
	WithdrawChain33Key                  = []byte("prefix_for_WithdrawChain33")
	LastTotalPowerKey                   = []byte("prefix_for_LastTotalPower")
	ValidatorMapsKey                    = []byte("prefix_for_ValidatorMaps")
	ConsensusThresholdKey               = []byte("prefix_for_ConsensusThreshold")
	TokenSymbolTotalLockOrBurnAmountKey = []byte("prefix_for_TokenSymbolTotalLockOrBurnAmount-")
	TokenSymbolToTokenAddressKey        = []byte("prefix_for_TokenSymbolToTokenAddress-")
)

// log for x2ethereum
// log类型id值
const (
	TyUnknownLog = iota + 100
	TyEth2Chain33Log
	TyWithdrawEthLog
	TyWithdrawChain33Log
	TyChain33ToEthLog
	TyAddValidatorLog
	TyRemoveValidatorLog
	TyModifyPowerLog
	TySetConsensusThresholdLog
	TyProphecyLog
	TyTransferLog
)

// action类型id和name，这些常量可以自定义修改
const (
	TyUnknowAction = iota + 100
	TyEth2Chain33Action
	TyWithdrawEthAction
	TyWithdrawChain33Action
	TyChain33ToEthAction
	TyAddValidatorAction
	TyRemoveValidatorAction
	TyModifyPowerAction
	TySetConsensusThresholdAction
	TyTransferAction

	NameEth2Chain33Action           = "Eth2Chain33"
	NameWithdrawEthAction           = "WithdrawEth"
	NameWithdrawChain33Action       = "WithdrawChain33"
	NameChain33ToEthAction          = "Chain33ToEth"
	NameAddValidatorAction          = "AddValidator"
	NameRemoveValidatorAction       = "RemoveValidator"
	NameModifyPowerAction           = "ModifyPower"
	NameSetConsensusThresholdAction = "SetConsensusThreshold"
	NameTransferAction              = "Transfer"
)

const ModuleName = "ETH"

const DefaultConsensusNeeded = int64(70)

const (
	DirEth2Chain33  = "eth2chain33"
	DirChain33ToEth = "chain33toeth"
	LOCK_CLAIM      = "lock"
	BURN_CLAIM      = "burn"
)

var DirectionType = [3]string{"", DirEth2Chain33, DirChain33ToEth}

// query function name
const (
	FuncQueryEthProphecy               = "GetEthProphecy"
	FuncQueryValidators                = "GetValidators"
	FuncQueryTotalPower                = "GetTotalPower"
	FuncQueryConsensusThreshold        = "GetConsensusThreshold"
	FuncQuerySymbolTotalAmountByTxType = "GetSymbolTotalAmountByTxType"
	FuncQueryRelayerBalance            = "GetRelayerBalance"
)

//设置合约管理员地址
const X2ethereumAdmin = "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"

const (
	LOCK_CLAIM_TYPE = int32(1)
	BURN_CLAIM_TYPE = int32(2)
)
