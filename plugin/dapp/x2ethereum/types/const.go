package types

//key
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
	TyTransferToExecLog
	TyWithdrawFromExecLog
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
	TyTransferToExecAction
	TyWithdrawFromExecAction

	NameEth2Chain33Action           = "Eth2Chain33Lock"
	NameWithdrawEthAction           = "Eth2Chain33Burn"
	NameWithdrawChain33Action       = "Chain33ToEthBurn"
	NameChain33ToEthAction          = "Chain33ToEthLock"
	NameAddValidatorAction          = "AddValidator"
	NameRemoveValidatorAction       = "RemoveValidator"
	NameModifyPowerAction           = "ModifyPower"
	NameSetConsensusThresholdAction = "SetConsensusThreshold"
	NameTransferAction              = "Transfer"
	NameTransferToExecAction        = "TransferToExec"
	NameWithdrawFromExecAction      = "WithdrawFromExec"
)

//DefaultConsensusNeeded ...
const DefaultConsensusNeeded = int64(70)

//direct ...
const (
	DirEth2Chain33  = "eth2chain33"
	DirChain33ToEth = "chain33toeth"
	LockClaim       = "lock"
	BurnClaim       = "burn"
)

//DirectionType type
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

//lock type
const (
	LockClaimType = int32(1)
	BurnClaimType = int32(2)
)
