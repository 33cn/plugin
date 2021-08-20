package relayer

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	rpctypes "github.com/33cn/chain33/rpc/types"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/chain33"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum"
	relayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	lru "github.com/hashicorp/golang-lru"
)

var (
	mlog = log15.New("relayer manager", "manager")
)

//status ...
const (
	Locked        = int32(1)
	Unlocked      = int32(99)
	EncryptEnable = int64(1)
)

//Manager ...
type Manager struct {
	chain33Relayer *chain33.Relayer4Chain33
	ethRelayer     *ethereum.Relayer4Ethereum
	store          *Store
	isLocked       int32
	mtx            sync.Mutex
	encryptFlag    int64
	passphase      string
	decimalLru     *lru.Cache
}

//NewRelayerManager ...
//1.验证人的私钥需要通过cli命令行进行导入，且chain33和ethereum两种不同的验证人需要分别导入
//2.显示或者重新替换原有的私钥首先需要通过passpin进行unlock的操作
func NewRelayerManager(chain33Relayer *chain33.Relayer4Chain33, ethRelayer *ethereum.Relayer4Ethereum, db dbm.DB) *Manager {
	l, _ := lru.New(4096)
	manager := &Manager{
		chain33Relayer: chain33Relayer,
		ethRelayer:     ethRelayer,
		store:          NewStore(db),
		isLocked:       Locked,
		mtx:            sync.Mutex{},
		encryptFlag:    0,
		passphase:      "",
		decimalLru:     l,
	}
	manager.encryptFlag = manager.store.GetEncryptionFlag()
	return manager
}

//SetPassphase ...
func (manager *Manager) SetPassphase(setPasswdReq relayerTypes.ReqSetPasswd, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()

	// 第一次设置密码的时候才使用 后面用 ChangePasswd
	if EncryptEnable == manager.encryptFlag {
		return errors.New("passphase alreade exists")
	}

	// 密码合法性校验
	if !utils.IsValidPassWord(setPasswdReq.Passphase) {
		return chain33Types.ErrInvalidPassWord
	}

	//使用密码生成passwdhash用于下次密码的验证
	newBatch := manager.store.NewBatch(true)
	err := manager.store.SetPasswordHash(setPasswdReq.Passphase, newBatch)
	if err != nil {
		mlog.Error("SetPassphase", "SetPasswordHash err", err)
		return err
	}
	//设置钱包加密标志位
	err = manager.store.SetEncryptionFlag(newBatch)
	if err != nil {
		mlog.Error("SetPassphase", "SetEncryptionFlag err", err)
		return err
	}

	err = newBatch.Write()
	if err != nil {
		mlog.Error("ProcWalletSetPasswd newBatch.Write", "err", err)
		return err
	}
	manager.passphase = setPasswdReq.Passphase
	atomic.StoreInt64(&manager.encryptFlag, EncryptEnable)

	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  "Succeed to set passphase",
	}
	return nil
}

//ChangePassphase ...
func (manager *Manager) ChangePassphase(setPasswdReq relayerTypes.ReqChangePasswd, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if setPasswdReq.OldPassphase == setPasswdReq.NewPassphase {
		return errors.New("the old password is the same as the new one")
	}
	// 新密码合法性校验
	if !utils.IsValidPassWord(setPasswdReq.NewPassphase) {
		return chain33Types.ErrInvalidPassWord
	}
	//保存钱包的锁状态，需要暂时的解锁，函数退出时再恢复回去
	tempislock := atomic.LoadInt32(&manager.isLocked)
	atomic.CompareAndSwapInt32(&manager.isLocked, Locked, Unlocked)

	defer func() {
		//wallet.isWalletLocked = tempislock
		atomic.CompareAndSwapInt32(&manager.isLocked, Unlocked, tempislock)
	}()

	// 钱包已经加密需要验证oldpass的正确性
	if len(manager.passphase) == 0 && manager.encryptFlag == EncryptEnable {
		isok := manager.store.VerifyPasswordHash(setPasswdReq.OldPassphase)
		if !isok {
			mlog.Error("ChangePassphase Verify Oldpasswd fail!")
			return chain33Types.ErrVerifyOldpasswdFail
		}
	}

	if len(manager.passphase) != 0 && setPasswdReq.OldPassphase != manager.passphase {
		mlog.Error("ChangePassphase Oldpass err!")
		return chain33Types.ErrVerifyOldpasswdFail
	}

	//使用新的密码生成passwdhash用于下次密码的验证
	newBatch := manager.store.NewBatch(true)
	err := manager.store.SetPasswordHash(setPasswdReq.NewPassphase, newBatch)
	if err != nil {
		mlog.Error("ChangePassphase", "SetPasswordHash err", err)
		return err
	}
	//设置钱包加密标志位
	err = manager.store.SetEncryptionFlag(newBatch)
	if err != nil {
		mlog.Error("ChangePassphase", "SetEncryptionFlag err", err)
		return err
	}

	err = manager.ethRelayer.StoreAccountWithNewPassphase(setPasswdReq.NewPassphase, setPasswdReq.OldPassphase)
	if err != nil {
		mlog.Error("ChangePassphase", "StoreAccountWithNewPassphase err", err)
		return err
	}

	err = manager.chain33Relayer.StoreAccountWithNewPassphase(setPasswdReq.NewPassphase, setPasswdReq.OldPassphase)
	if err != nil {
		mlog.Error("ChangePassphase", "StoreAccountWithNewPassphase err", err)
		return err
	}

	err = newBatch.Write()
	if err != nil {
		mlog.Error("ProcWalletSetPasswd newBatch.Write", "err", err)
		return err
	}
	manager.passphase = setPasswdReq.NewPassphase
	atomic.StoreInt64(&manager.encryptFlag, EncryptEnable)

	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  "Succeed to change passphase",
	}
	return nil
}

//Unlock 进行unlok操作
func (manager *Manager) Unlock(passphase string, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if EncryptEnable != manager.encryptFlag {
		return errors.New("pls set passphase first")
	}
	if Unlocked == manager.isLocked {
		return errors.New("unlock already")
	}

	if !manager.store.VerifyPasswordHash(passphase) {
		return errors.New("wrong passphase")
	}

	if err := manager.chain33Relayer.RestorePrivateKeys(passphase); nil != err {
		info := fmt.Sprintf("Failed to RestorePrivateKeys for chain33Relayer due to:%s", err.Error())
		return errors.New(info)
	}
	if err := manager.ethRelayer.RestorePrivateKeys(passphase); nil != err {
		info := fmt.Sprintf("Failed to RestorePrivateKeys for ethRelayer due to:%s", err.Error())
		return errors.New(info)
	}

	manager.isLocked = Unlocked
	manager.passphase = passphase

	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  "Succeed to unlock",
	}

	return nil
}

//Lock 锁定操作，该操作一旦执行，就不能替换验证人的私钥，需要重新unlock之后才能修改
func (manager *Manager) Lock(param interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	manager.isLocked = Locked
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  "Succeed to lock",
	}
	return nil
}

//ImportChain33RelayerPrivateKey 导入chain33relayer验证人的私钥,该私钥实际用于向ethereum提交验证交易时签名使用
func (manager *Manager) ImportChain33RelayerPrivateKey(importKeyReq relayerTypes.ImportKeyReq, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	privateKey := importKeyReq.PrivateKey
	if err := manager.checkPermission(); nil != err {
		return err
	}
	err := manager.chain33Relayer.ImportPrivateKey(manager.passphase, privateKey)
	if err != nil {
		mlog.Error("ImportChain33ValidatorPrivateKey", "Failed due to cause:", err.Error())
		return err
	}

	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  "Succeed to import private key for chain33 relayer",
	}
	return nil
}

//GenerateEthereumPrivateKey 生成以太坊私钥
func (manager *Manager) GenerateEthereumPrivateKey(param interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	account4Show := relayerTypes.Account4Show{}
	var err error
	account4Show.Privkey, account4Show.Addr, err = manager.ethRelayer.NewAccount(manager.passphase)
	if nil != err {
		return err
	}
	*result = account4Show
	return nil
}

//ImportEthereumPrivateKey4EthRelayer 为ethrelayer导入chain33私钥，为向chain33发送交易时进行签名使用
func (manager *Manager) ImportEthereumPrivateKey4EthRelayer(privateKey string, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	addr, err := manager.ethRelayer.ImportPrivateKey(manager.passphase, privateKey)
	if err != nil {
		mlog.Error("ImportEthereumPrivateKey4EthRelayer", "Failed due to cause:", err.Error())
		return err
	}

	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  fmt.Sprintf("Succeed to import for address:%s", addr),
	}
	return nil
}

//ShowChain33RelayerValidator 显示在chain33中以验证人validator身份进行登录的地址
func (manager *Manager) ShowChain33RelayerValidator(param interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	var err error
	*result, err = manager.chain33Relayer.GetAccountAddr()
	if nil != err {
		return err
	}

	return nil
}

//ShowEthRelayerValidator 显示在Ethereum中以验证人validator身份进行登录的地址
func (manager *Manager) ShowEthRelayerValidator(param interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	var err error
	*result, err = manager.ethRelayer.GetValidatorAddr()
	if nil != err {
		return err
	}
	return nil
}

//IsValidatorActive ...
func (manager *Manager) IsValidatorActive(vallidatorAddr string, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	active, err := manager.ethRelayer.IsValidatorActive(vallidatorAddr)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: active,
		Msg:  "",
	}
	return nil
}

//ShowOperator ...
func (manager *Manager) ShowOperator(param interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	operator, err := manager.ethRelayer.ShowOperator()
	if nil != err {
		return err
	}
	*result = operator
	return nil
}

//DeployContrcts ...
func (manager *Manager) DeployContrcts(param interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	bridgeRegistry, err := manager.ethRelayer.DeployContrcts()
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  bridgeRegistry,
	}
	return nil
}

func (manager *Manager) Deploy2Chain33(param interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	bridgeRegistry, err := manager.chain33Relayer.DeployContracts()
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  bridgeRegistry,
	}
	return nil
}

func (manager *Manager) CreateERC20ToChain33(param relayerTypes.ERC20Token, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	bridgeRegistry, err := manager.chain33Relayer.CreateERC20ToChain33(param)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  bridgeRegistry,
	}
	return nil
}

func (manager *Manager) DeployMulsign2Chain33(_ interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	mulSign, err := manager.chain33Relayer.DeployMulsign()
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  mulSign,
	}
	return nil
}

func (manager *Manager) SetupOwner4Chain33(setupMulSign relayerTypes.SetupMulSign, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	mulSign, err := manager.chain33Relayer.SetupMulSign(setupMulSign)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  mulSign,
	}
	return nil
}

func (manager *Manager) SafeTransfer4Chain33(para relayerTypes.SafeTransfer, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	mulSign, err := manager.chain33Relayer.SafeTransfer(para)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  mulSign,
	}
	return nil
}

//CreateBridgeToken ...
func (manager *Manager) CreateBridgeToken(symbol string, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	tokenAddr, err := manager.ethRelayer.CreateBridgeToken(symbol)
	if nil != err {
		return err
	}
	*result = relayerTypes.ReplyAddr{
		IsOK: true,
		Addr: tokenAddr,
	}
	return nil
}

func (manager *Manager) AddToken2LockList(token relayerTypes.ETHTokenLockAddress, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	txhash, err := manager.ethRelayer.AddToken2LockList(token.Symbol, token.Address)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

//DeployERC20 ...
func (manager *Manager) DeployERC20(Erc20Token relayerTypes.ERC20Token, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}

	Erc20Addr, err := manager.ethRelayer.DeployERC20(Erc20Token.Owner, Erc20Token.Name, Erc20Token.Symbol, Erc20Token.Amount)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  Erc20Addr,
	}
	return nil
}

//ApproveAllowance ...
func (manager *Manager) ApproveAllowance(approveAllowance relayerTypes.ApproveAllowance, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	txhash, err := manager.ethRelayer.ApproveAllowance(approveAllowance.OwnerKey, approveAllowance.TokenAddr, approveAllowance.Amount)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

//Burn ...
func (manager *Manager) Burn(burn relayerTypes.Burn, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	txhash, err := manager.ethRelayer.Burn(burn.OwnerKey, burn.TokenAddr, burn.Chain33Receiver, burn.Amount)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

//BurnAsync ...
func (manager *Manager) BurnAsync(burn relayerTypes.Burn, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	txhash, err := manager.ethRelayer.BurnAsync(burn.OwnerKey, burn.TokenAddr, burn.Chain33Receiver, burn.Amount)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

// SimBurnFromEth : 模拟从eth销毁资产，提币回到chain33,使用LockBTY仅为测试使用
func (manager *Manager) SimBurnFromEth(burn relayerTypes.Burn, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	err := manager.ethRelayer.SimBurnFromEth(burn)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
	}
	return nil
}

// SimLockFromEth : 模拟从eth锁住eth/erc20，转移到chain33
func (manager *Manager) SimLockFromEth(lock relayerTypes.LockEthErc20, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	err := manager.ethRelayer.SimLockFromEth(lock)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
	}
	return nil
}

func (manager *Manager) BurnAsyncFromChain33(burn relayerTypes.BurnFromChain33, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	txhash, err := manager.chain33Relayer.BurnAsyncFromChain33(burn.OwnerKey, burn.TokenAddr, burn.EthereumReceiver, burn.Amount)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

func (manager *Manager) LockBTYAssetAsync(lockEthErc20Asset relayerTypes.LockBTY, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	txhash, err := manager.chain33Relayer.LockBTYAssetAsync(lockEthErc20Asset.OwnerKey, lockEthErc20Asset.Amount, lockEthErc20Asset.EtherumReceiver)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

//LockEthErc20AssetAsync ...
func (manager *Manager) LockEthErc20AssetAsync(lockEthErc20Asset relayerTypes.LockEthErc20, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	txhash, err := manager.ethRelayer.LockEthErc20AssetAsync(lockEthErc20Asset.OwnerKey, lockEthErc20Asset.TokenAddr, lockEthErc20Asset.Amount, lockEthErc20Asset.Chain33Receiver)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

//LockEthErc20Asset ...
func (manager *Manager) LockEthErc20Asset(lockEthErc20Asset relayerTypes.LockEthErc20, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	txhash, err := manager.ethRelayer.LockEthErc20Asset(lockEthErc20Asset.OwnerKey, lockEthErc20Asset.TokenAddr, lockEthErc20Asset.Amount, lockEthErc20Asset.Chain33Receiver)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

//IsProphecyPending ...
func (manager *Manager) IsProphecyPending(claimID [32]byte, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	active, err := manager.ethRelayer.IsProphecyPending(claimID)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: active,
	}
	return nil
}

//GetBalance ...
func (manager *Manager) GetBalance(balanceAddr relayerTypes.BalanceAddr, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	balance, err := manager.ethRelayer.GetBalance(balanceAddr.TokenAddr, balanceAddr.Owner)
	if nil != err {
		return err
	}

	var d int64
	if balanceAddr.TokenAddr == "" || balanceAddr.TokenAddr == "0x0000000000000000000000000000000000000000" {
		d = 18
	} else {
		d, err = manager.GetDecimals(balanceAddr.TokenAddr)
		if err != nil {
			return errors.New("get decimals error")
		}
	}

	*result = relayerTypes.ReplyBalance{
		IsOK:    true,
		Balance: utils.TrimZeroAndDot(strconv.FormatFloat(utils.Toeth(balance, d), 'f', 4, 64)),
	}
	return nil
}

func (manager *Manager) ShowMultiBalance(balanceAddr relayerTypes.BalanceAddr, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	balance, err := manager.ethRelayer.ShowMultiBalance(balanceAddr.TokenAddr, balanceAddr.Owner)
	if nil != err {
		return err
	}

	*result = relayerTypes.ReplyBalance{
		IsOK:    true,
		Balance: balance,
	}
	return nil
}

//ShowBridgeBankAddr ...
func (manager *Manager) ShowBridgeBankAddr(para interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	addr, err := manager.ethRelayer.ShowBridgeBankAddr()
	if nil != err {
		return err
	}
	*result = relayerTypes.ReplyAddr{
		IsOK: true,
		Addr: addr,
	}
	return nil
}

//ShowBridgeRegistryAddr ...
func (manager *Manager) ShowBridgeRegistryAddr(para interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	addr, err := manager.ethRelayer.ShowBridgeRegistryAddr()
	if nil != err {
		return err
	}
	*result = relayerTypes.ReplyAddr{
		IsOK: true,
		Addr: addr,
	}
	return nil
}

//ShowBridgeRegistryAddr4chain33 ...
func (manager *Manager) ShowBridgeRegistryAddr4chain33(para interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	addr, err := manager.chain33Relayer.ShowBridgeRegistryAddr()
	if nil != err {
		return err
	}
	*result = relayerTypes.ReplyAddr{
		IsOK: true,
		Addr: addr,
	}
	return nil
}

//SetTokenAddress ...
func (manager *Manager) SetTokenAddress(token2set relayerTypes.TokenAddress, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}

	if relayerTypes.EthereumBlockChainName == token2set.ChainName {
		err := manager.ethRelayer.SetTokenAddress(token2set)
		if nil != err {
			return err
		}
	} else {
		err := manager.chain33Relayer.SetTokenAddress(token2set)
		if nil != err {
			return err
		}
	}

	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  "",
	}
	return nil
}

//ShowTokenAddress ...
func (manager *Manager) ShowTokenAddress(token2show relayerTypes.TokenAddress, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}

	var res *relayerTypes.TokenAddressArray
	var err error
	if relayerTypes.EthereumBlockChainName == token2show.ChainName {
		res, err = manager.ethRelayer.ShowTokenAddress(token2show)
		if nil != err {
			return err
		}
	} else {
		res, err = manager.chain33Relayer.ShowTokenAddress(token2show)
		if nil != err {
			return err
		}
	}

	*result = *res

	return nil
}

func (manager *Manager) ShowETHLockTokenAddress(token2show relayerTypes.TokenAddress, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}

	res, err := manager.ethRelayer.ShowETHLockTokenAddress(token2show)
	if nil != err {
		return err
	}

	*result = *res

	return nil
}

//ShowTxReceipt ...
func (manager *Manager) ShowTxReceipt(txhash string, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	receipt, err := manager.ethRelayer.ShowTxReceipt(txhash)
	if nil != err {
		return err
	}
	*result = *receipt
	return nil
}

func (manager *Manager) checkPermission() error {
	if EncryptEnable != manager.encryptFlag {
		return errors.New("pls set passphase first")
	}
	if Locked == manager.isLocked {
		return errors.New("pls unlock this relay-manager first")
	}
	return nil
}

// ShowTokenStatics ShowEthRelayer2Chain33Txs ...
func (manager *Manager) ShowTokenStatics(request relayerTypes.TokenStaticsRequest, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}

	if request.From != 0 && 1 != request.From {
		return errors.New("wrong source chain flag")
	}

	if request.Operation != 2 && 1 != request.Operation {
		return errors.New("wrong Operation flag")
	}

	if request.Status < 0 || request.Status > 3 {
		return errors.New("wrong Status flag")
	}

	if relayerTypes.Source_Chain_Chain33 == request.From {
		res, err := manager.ethRelayer.ShowStatics(request)
		if nil != err {
			return err
		}
		*result = *res
	} else {
		res, err := manager.chain33Relayer.ShowStatics(request)
		if nil != err {
			return err
		}
		*result = *res
	}
	return nil
}

//TransferToken ...
func (manager *Manager) TransferToken(transfer relayerTypes.TransferToken, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	txhash, err := manager.ethRelayer.TransferToken(transfer.TokenAddr, transfer.FromKey, transfer.ToAddr, transfer.Amount)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

//GetDecimals ...
func (manager *Manager) GetDecimals(tokenAddr string) (int64, error) {
	if d, ok := manager.decimalLru.Get(tokenAddr); ok {
		mlog.Info("GetDecimals", "from cache", d)
		return d.(int64), nil
	}

	if d, err := manager.store.Get(utils.CalAddr2DecimalsPrefix(tokenAddr)); err == nil {
		decimal, err := strconv.ParseInt(string(d), 10, 64)
		if err != nil {
			return 0, err
		}
		manager.decimalLru.Add(tokenAddr, decimal)
		mlog.Info("GetDecimals", "from DB", d)

		return decimal, nil
	}

	d, err := manager.ethRelayer.GetDecimals(tokenAddr)
	if err != nil {
		return 0, err
	}
	_ = manager.store.Set(utils.CalAddr2DecimalsPrefix(tokenAddr), []byte(strconv.FormatInt(int64(d), 10)))
	manager.decimalLru.Add(tokenAddr, int64(d))

	mlog.Info("GetDecimals", "from Node", d)

	return int64(d), nil
}

func (manager *Manager) DeployMulsign2Eth(param interface{}, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	mulSign, err := manager.ethRelayer.DeployMulsign()
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  mulSign,
	}
	return nil
}

func (manager *Manager) SetupOwner4Eth(setupMulSign relayerTypes.SetupMulSign, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	mulSign, err := manager.ethRelayer.SetupMulSign(setupMulSign)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  mulSign,
	}
	return nil
}

func (manager *Manager) SafeTransfer4Eth(para relayerTypes.SafeTransfer, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	mulSign, err := manager.ethRelayer.SafeTransfer(para)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  mulSign,
	}
	return nil
}

func (manager *Manager) ConfigOfflineSaveAccount(addr string, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	txhash, err := manager.ethRelayer.ConfigOfflineSaveAccount(addr)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

func (manager *Manager) ConfigLockedTokenOfflineSave(config relayerTypes.ETHConfigLockedTokenOffline, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	txhash, err := manager.ethRelayer.ConfigLockedTokenOfflineSave(config.Address, config.Symbol, config.Threshold, config.Percents)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

//TransferEth ...
func (manager *Manager) TransferEth(transfer relayerTypes.TransferToken, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	txhash, err := manager.ethRelayer.TransferEth(transfer.FromKey, transfer.ToAddr, transfer.Amount)
	if nil != err {
		return err
	}
	*result = rpctypes.Reply{
		IsOk: true,
		Msg:  txhash,
	}
	return nil
}

func (manager *Manager) SetChain33MultiSignAddr(multiSignAddr string, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	manager.chain33Relayer.SetMultiSignAddr(multiSignAddr)
	*result = rpctypes.Reply{
		IsOk: true,
	}
	return nil
}

func (manager *Manager) SetEthMultiSignAddr(multiSignAddr string, result *interface{}) error {
	manager.mtx.Lock()
	defer manager.mtx.Unlock()
	if err := manager.checkPermission(); nil != err {
		return err
	}
	manager.ethRelayer.SetMultiSignAddr(multiSignAddr)
	*result = rpctypes.Reply{
		IsOk: true,
	}
	return nil
}
