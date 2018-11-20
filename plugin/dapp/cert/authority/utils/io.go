// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// DirMissingOrEmpty
func DirMissingOrEmpty(path string) (bool, error) {
	dirExists, err := DirExists(path)
	if err != nil {
		return false, err
	}
	if !dirExists {
		return true, nil
	}

	dirEmpty, err := DirEmpty(path)
	if err != nil {
		return false, err
	}
	if dirEmpty {
		return true, nil
	}
	return false, nil
}

// DirExists
func DirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DirEmpty
func DirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// ReadFile
func ReadFile(file string) ([]byte, error) {
	fileCont, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Could not read file %s, err %s", file, err)
	}

	return fileCont, nil
}

// ReadPemFile
func ReadPemFile(file string) ([]byte, error) {
	bytes, err := ReadFile(file)
	if err != nil {
		return nil, err
	}

	b, _ := pem.Decode(bytes)
	if b == nil {
		return nil, fmt.Errorf("No pem content for file %s", file)
	}

	return bytes, nil
}

// DeleteFile
func DeleteFile(file string) error {
	return os.Remove(file)
}
