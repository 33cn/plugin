// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const (
	// RFC3339Millis ...
	RFC3339Millis = "2006-01-02T15:04:05.000Z" // forced microseconds
	timeFormat    = RFC3339Millis
)

var (
	randgen *rand.Rand

	// Fmt ...
	Fmt = fmt.Sprintf
	once sync.Once
)

// Init ...
func Init() {
	once.Do(func () {
		if randgen == nil {
			randgen = rand.New(rand.NewSource(time.Now().UnixNano()))
		}
	})
}

// WriteFile ...
func WriteFile(filePath string, contents []byte, mode os.FileMode) error {
	return ioutil.WriteFile(filePath, contents, mode)
}

// WriteFileAtomic ...
func WriteFileAtomic(filePath string, newBytes []byte, mode os.FileMode) error {
	dir := filepath.Dir(filePath)
	f, err := ioutil.TempFile(dir, "")
	if err != nil {
		return err
	}
	_, err = f.Write(newBytes)
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if permErr := os.Chmod(f.Name(), mode); err == nil {
		err = permErr
	}
	if err == nil {
		err = os.Rename(f.Name(), filePath)
	}
	// any err should result in full cleanup
	if err != nil {
		if er := os.Remove(f.Name()); er != nil {
			fmt.Printf("WriteFileAtomic Remove failed:%v", er)
		}
	}
	return err
}

// Tempfile ...
func Tempfile(prefix string) (*os.File, string) {
	file, err := ioutil.TempFile("", prefix)
	if err != nil {
		panic(err)
	}
	return file, file.Name()
}

// Fingerprint ...
func Fingerprint(slice []byte) []byte {
	fingerprint := make([]byte, 6)
	copy(fingerprint, slice)
	return fingerprint
}

// Kill ...
func Kill() error {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	return p.Signal(syscall.SIGTERM)
}

// Exit ...
func Exit(s string) {
	fmt.Printf(s + "\n")
	os.Exit(1)
}

// Parallel ...
func Parallel(tasks ...func()) {
	var wg sync.WaitGroup
	wg.Add(len(tasks))
	for _, task := range tasks {
		go func(task func()) {
			task()
			wg.Done()
		}(task)
	}
	wg.Wait()
}

// MinInt ...
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt ...
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RandIntn ...
func RandIntn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	if n <= 1<<31-1 {
		//randMux.Lock()
		i32 := randgen.Int31n(int32(n))
		//randMux.Unlock()
		return int(i32)
	}
	//randMux.Lock()
	i64 := randgen.Int63n(int64(n))
	//randMux.Unlock()
	return int(i64)
}

// RandUint32 ...
func RandUint32() uint32 {
	//randMux.Lock()
	u32 := randgen.Uint32()
	//randMux.Unlock()
	return u32
}

// RandInt63n ...
func RandInt63n(n int64) int64 {
	//randMux.Lock()
	i64 := randgen.Int63n(n)
	//randMux.Unlock()
	return i64
}

// PanicSanity ...
func PanicSanity(v interface{}) {
	panic(Fmt("Panicked on a Sanity Check: %v", v))
}

// PanicCrisis ...
func PanicCrisis(v interface{}) {
	panic(Fmt("Panicked on a Crisis: %v", v))
}

// PanicQ ...
func PanicQ(v interface{}) {
	panic(Fmt("Panicked questionably: %v", v))
}
