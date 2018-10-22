package types

//unfreeze action ty
const (
	UnfreezeActionCreate = iota + 1
	UnfreezeActionWithdraw
	UnfreezeActionTerminate

	//log for unfreeze
	TyLogCreateUnfreeze    = 2001 // TODO 修改具体编号
	TyLogWithdrawUnfreeze  = 2002
	TyLogTerminateUnfreeze = 2003
)

//包的名字可以通过配置文件来配置
//建议用github的组织名称，或者用户名字开头, 再加上自己的插件的名字
//如果发生重名，可以通过配置文件修改这些名字
var (
	PackageName    = "chain33.unfreeze"
	RpcName        = "Chain33.Unfreeze"
	UnfreezeX      = "unfreeze"
	ExecerUnfreeze = []byte(UnfreezeX)
)

const (
	Action_CreateUnfreeze    = "createUnfreeze"
	Action_WithdrawUnfreeze  = "withdrawUnfreeze"
	Action_TerminateUnfreeze = "terminateUnfreeze"
)

const (
	FuncName_QueryUnfreezeWithdraw = "QueryUnfreezeWithdraw"
)
