package executor

import (
	"testing"
	"unsafe"

	"github.com/33cn/chain33/common/address"
	lru "github.com/hashicorp/golang-lru"
	"github.com/stretchr/testify/assert"
)

var (
//privOpener = getprivkey("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944") //opener
//privPlayer = getprivkey("4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01") //player
//opener     = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
//player     = "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
)

func Test_getJvmExector_fail(t *testing.T) {
	address1 := &address.Address{}
	envHandleUintptr := uintptr(unsafe.Pointer(address1))
	jvmsCached, _ = lru.New(1000)
	_, ok := getJvmExector(envHandleUintptr)
	assert.Equal(t, false, ok)

	address2 := &address.Address{}

	jvmsCached.Add(envHandleUintptr, address2)
	_, ok = getJvmExector(envHandleUintptr)
	assert.Equal(t, false, ok)
}

func Test_execFrozen_fail(t *testing.T) {
	address1 := &address.Address{}
	envHandleUintptr := uintptr(unsafe.Pointer(address1))
	jvmsCached, _ = lru.New(1000)
	ok := execFrozen(opener, 100, envHandleUintptr)
	assert.Equal(t, false, ok)

	jvmTestEnv := setupTestEnv()

	jvmsCached.Add(envHandleUintptr, jvmTestEnv.jvm)
	ok = execFrozen(opener, 100, envHandleUintptr)
	assert.Equal(t, false, ok)

}

func Test_execActive_fail(t *testing.T) {
	address1 := &address.Address{}
	envHandleUintptr := uintptr(unsafe.Pointer(address1))
	jvmsCached, _ = lru.New(1000)
	ok := execActive(opener, 100, envHandleUintptr)
	assert.Equal(t, false, ok)

	jvmTestEnv := setupTestEnv()

	jvmsCached.Add(envHandleUintptr, jvmTestEnv.jvm)
	ok = execActive(opener, 100, envHandleUintptr)
	assert.Equal(t, false, ok)
}

func Test_execTransfer_fail(t *testing.T) {
	address1 := &address.Address{}
	envHandleUintptr := uintptr(unsafe.Pointer(address1))
	jvmsCached, _ = lru.New(1000)
	ok := execTransfer(opener, player, 100, envHandleUintptr)
	assert.Equal(t, false, ok)

	jvmTestEnv := setupTestEnv()

	jvmsCached.Add(envHandleUintptr, jvmTestEnv.jvm)
	ok = execTransfer(opener, player, 100, envHandleUintptr)
	assert.Equal(t, false, ok)
}

func Test_getFrom_fail(t *testing.T) {
	address1 := &address.Address{}
	envHandleUintptr := uintptr(unsafe.Pointer(address1))
	jvmsCached, _ = lru.New(1000)
	from := getFrom(envHandleUintptr)
	assert.Equal(t, "", from)

	jvmTestEnv := setupTestEnv()

	jvmsCached.Add(envHandleUintptr, jvmTestEnv.jvm)
	from = getFrom(envHandleUintptr)
	assert.Equal(t, "", from)
}

func Test_getHeight_fail(t *testing.T) {
	address1 := &address.Address{}
	envHandleUintptr := uintptr(unsafe.Pointer(address1))
	jvmsCached, _ = lru.New(1000)
	jvmsCached.Add(envHandleUintptr, envHandleUintptr)
	height := getHeight(envHandleUintptr)
	assert.Equal(t, int64(0), height)

	jvmTestEnv := setupTestEnv()

	jvmsCached.Add(envHandleUintptr, jvmTestEnv.jvm)
	height = getHeight(envHandleUintptr)
	assert.Equal(t, int64(10), height)
}

func Test_stopTransWithErrInfo_fail(t *testing.T) {
	address1 := &address.Address{}
	envHandleUintptr := uintptr(unsafe.Pointer(address1))
	jvmsCached, _ = lru.New(1000)
	jvmsCached.Add(envHandleUintptr, envHandleUintptr)
	ok := stopTransWithErrInfo("err", envHandleUintptr)
	assert.Equal(t, false, ok)
}
