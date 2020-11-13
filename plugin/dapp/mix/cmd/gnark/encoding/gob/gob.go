/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gob

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
)

var (
	ErrInvalidCurve = errors.New("trying to deserialize an object serialized with another curve")
)

// Write serialize object into file
// uses gob + gzip
func Write(path string, from interface{}) error {
	// create file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return Serialize(f, from)
}

// Read read and deserialize input into object
// provided interface must be a pointer
// uses gob + gzip
func Read(path string, into interface{}) error {
	// open file
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fileinfo, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}

	fileSize := fileinfo.Size()
	buffer := make([]byte, fileSize)

	_, err = f.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	buf.Write(buffer)

	return Deserialize(&buf, into)
}

func ReadBuf(str string, into interface{}) error {
	strByts, err := hex.DecodeString(str)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	buf.Write(strByts)

	return Deserialize(&buf, into)
}

// Serialize object from into f
// uses gob
func Serialize(f io.Writer, from interface{}) error {

	// gzip writer
	encoder := gob.NewEncoder(f)

	// encode our object
	if err := encoder.Encode(from); err != nil {
		return err
	}

	return nil
}

// Deserialize f into object into
// uses gob + gzip
func Deserialize(f io.Reader, into interface{}) error {

	// gzip reader
	decoder := gob.NewDecoder(f)

	if err := decoder.Decode(into); err != nil {
		return err
	}

	return nil
}
