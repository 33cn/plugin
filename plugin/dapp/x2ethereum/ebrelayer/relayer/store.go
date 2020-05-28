package relayer

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var (
	storelog = log15.New("relayer manager", "store")
)

const (
	keyEncryptionFlag     = "Encryption"
	keyEncryptionCompFlag = "EncryptionFlag" // 中间有一段时间运行了一个错误的密码版本，导致有部分用户信息发生错误，需要兼容下
	keyPasswordHash       = "PasswordHash"
)

// CalcEncryptionFlag 加密标志Key
func calcEncryptionFlag() []byte {
	return []byte(keyEncryptionFlag)
}

// calckeyEncryptionCompFlag 加密比较标志Key
func calckeyEncryptionCompFlag() []byte {
	return []byte(keyEncryptionCompFlag)
}

// CalcPasswordHash 密码hash的Key
func calcPasswordHash() []byte {
	return []byte(keyPasswordHash)
}

// NewStore 新建存储对象
func NewStore(db db.DB) *Store {
	return &Store{db: db}
}

// Store 钱包通用数据库存储类，实现对钱包账户数据库操作的基本实现
type Store struct {
	db db.DB
}

// Close 关闭数据库
func (store *Store) Close() {
	store.db.Close()
}

// GetDB 获取数据库操作接口
func (store *Store) GetDB() db.DB {
	return store.db
}

// NewBatch 新建批处理操作对象接口
func (store *Store) NewBatch(sync bool) db.Batch {
	return store.db.NewBatch(sync)
}

// Get 取值
func (store *Store) Get(key []byte) ([]byte, error) {
	return store.db.Get(key)
}

// Set 设置值
func (store *Store) Set(key []byte, value []byte) (err error) {
	return store.db.Set(key, value)
}

// NewListHelper 新建列表复制操作对象
func (store *Store) NewListHelper() *db.ListHelper {
	return db.NewListHelper(store.db)
}

// SetEncryptionFlag 设置加密方式标志
func (store *Store) SetEncryptionFlag(batch db.Batch) error {
	var flag int64 = 1
	data, err := json.Marshal(flag)
	if err != nil {
		storelog.Error("SetEncryptionFlag marshal flag", "err", err)
		return types.ErrMarshal
	}

	batch.Set(calcEncryptionFlag(), data)
	return nil
}

// GetEncryptionFlag 获取加密方式
func (store *Store) GetEncryptionFlag() int64 {
	var flag int64
	data, err := store.Get(calcEncryptionFlag())
	if data == nil || err != nil {
		data, err = store.Get(calckeyEncryptionCompFlag())
		if data == nil || err != nil {
			return 0
		}
	}
	err = json.Unmarshal(data, &flag)
	if err != nil {
		storelog.Error("GetEncryptionFlag unmarshal", "err", err)
		return 0
	}
	return flag
}

// SetPasswordHash 保存密码哈希
func (store *Store) SetPasswordHash(password string, batch db.Batch) error {
	var WalletPwHash types.WalletPwHash
	//获取一个随机字符串
	randstr := fmt.Sprintf("fuzamei:$@%s", crypto.CRandHex(16))
	WalletPwHash.Randstr = randstr

	//通过password和随机字符串生成一个hash值
	pwhashstr := fmt.Sprintf("%s:%s", password, WalletPwHash.Randstr)
	pwhash := sha256.Sum256([]byte(pwhashstr))
	WalletPwHash.PwHash = pwhash[:]

	pwhashbytes, err := json.Marshal(WalletPwHash)
	if err != nil {
		storelog.Error("SetEncryptionFlag marshal flag", "err", err)
		return types.ErrMarshal
	}
	batch.Set(calcPasswordHash(), pwhashbytes)
	return nil
}

// VerifyPasswordHash 检查密码有效性
func (store *Store) VerifyPasswordHash(password string) bool {
	var WalletPwHash types.WalletPwHash
	pwhashbytes, err := store.Get(calcPasswordHash())
	if pwhashbytes == nil || err != nil {
		return false
	}
	err = json.Unmarshal(pwhashbytes, &WalletPwHash)
	if err != nil {
		storelog.Error("VerifyPasswordHash unmarshal", "err", err)
		return false
	}
	pwhashstr := fmt.Sprintf("%s:%s", password, WalletPwHash.Randstr)
	pwhash := sha256.Sum256([]byte(pwhashstr))
	Pwhash := pwhash[:]
	//通过新的密码计算pwhash最对比
	return bytes.Equal(WalletPwHash.GetPwHash(), Pwhash)
}
