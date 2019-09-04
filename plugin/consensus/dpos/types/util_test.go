package types

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"os/signal"
	"strings"
	"sync"
	"testing"
	"time"
)

func init() {
	Init()
}

func TestWriteFile(t *testing.T) {
	filename := "./tmp_priv_validator.json"
	err := WriteFile(filename, []byte(privValidatorFile), 0664)
	require.Nil(t, err)

	file, err := os.Stat(filename)
	require.Nil(t, err)
	//assert.True(t, file.Mode() == 077)
	fmt.Println(file.Name())
	fmt.Println(file.Mode())

	assert.True(t, file.Name() == "tmp_priv_validator.json")
	assert.True(t, file.Mode() == 0664)

	remove(filename)
}

func TestWriteFileAtomic(t *testing.T) {
	filename := "./tmp_priv_validator.json"
	err := WriteFileAtomic(filename, []byte(privValidatorFile), 0664)
	require.Nil(t, err)

	file, err := os.Stat(filename)
	require.Nil(t, err)
	//assert.True(t, file.Mode() == 077)
	fmt.Println(file.Name())
	fmt.Println(file.Mode())

	assert.True(t, file.Name() == "tmp_priv_validator.json")
	assert.True(t, file.Mode() == 0664)

	remove(filename)
}

func TestTempfile(t *testing.T) {
	filename := "tmp_priv_validator.json"
	file, name := Tempfile(filename)
	fmt.Println(name)
	require.NotNil(t, file)

	_, err := file.Write([]byte(privValidatorFile))
	if err == nil {
		err = file.Sync()
	}
	require.Nil(t, err)

	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	require.Nil(t, err)

	if permErr := os.Chmod(file.Name(), 0777); err == nil {
		err = permErr
	}
	require.Nil(t, err)

	remove(name)
}

func TestFingerprint(t *testing.T) {
	arr := []byte("abdcdfasdf")
	finger := Fingerprint(arr)
	assert.True(t, bytes.Equal(finger, arr[0:6]))
}

func TestKill(t *testing.T) {
	c := make(chan os.Signal)
	signal.Notify(c)
	go Kill()
	s := <-c
	assert.True(t, s.String() == "terminated")
}

var (
	goIndex = 0
	goIndexMutex sync.Mutex

	goSum   = 0
	goSumMutex sync.Mutex
)

func test() {
	goIndexMutex.Lock()
	goIndex++
	goIndexMutex.Unlock()
	time.Sleep(time.Second * time.Duration(goIndex))
	goSumMutex.Lock()
	goSum++
	goSumMutex.Unlock()
}

func TestParallel(t *testing.T) {
	f1 := test
	f1()

	f2 := test
	f2()

	goSumMutex.Lock()
	goSum = 0
	goSumMutex.Unlock()

	Parallel(f1, f2)
	goSumMutex.Lock()
	assert.True(t, goSum == 2)
	goSumMutex.Unlock()

}

func TestRandInt63n(t *testing.T) {
	a := RandInt63n(10)
	assert.True(t, a < 10)

	b := RandInt63n(9999999999999999)
	assert.True(t, b < 9999999999999999)

}

func TestRandIntn(t *testing.T) {
	a := RandIntn(10)
	assert.True(t, a < 10)

	b := RandIntn(9999999999999)
	assert.True(t, b < 9999999999999)
}

func TestRandUint32(t *testing.T) {
	a := RandUint32()
	assert.True(t, a >= 0)

	b := RandUint32()
	assert.True(t, b >= 0)
}

func TestPanicSanity(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			//fmt.Println(r)
			assert.True(t, strings.HasPrefix(r.(string), "Panicked on a Sanity Check: "))
		}
	}()

	PanicSanity("hello")
}

func TestPanicCrisis(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			//fmt.Println(r)
			assert.True(t, strings.HasPrefix(r.(string), "Panicked on a Crisis: "))
		}
	}()

	PanicCrisis("hello")
}

func TestPanicQ(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			//fmt.Println(r)
			assert.True(t, strings.HasPrefix(r.(string), "Panicked questionably: "))
		}
	}()

	PanicQ("hello")
}
