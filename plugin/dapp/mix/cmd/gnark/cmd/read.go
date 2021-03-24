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

package cmd

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"

	"encoding/hex"
	"encoding/json"

	"github.com/consensys/gnark/backend"

	"strings"

	"fmt"
	"io"

	"github.com/spf13/cobra"

	"os"
	"path/filepath"
)

// verifyCmd represents the verify command
var readCmd = &cobra.Command{
	Use:     "read [file] --flags",
	Short:   "read a file and show to hex string",
	Run:     cmdRead,
	Version: Version,
}
var (
	fType int32
)

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.PersistentFlags().Int32VarP(&fType, "type", "t", 0, "0: proof or vk file, 1: input file, default 0")

}

func cmdRead(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("missing read file path -- gnark read -h for help")
		os.Exit(-1)
	}
	filePath := filepath.Clean(args[0])

	if fType == 1 {
		readInput2(filePath)
		return
	}

	readProof(filePath)

}

func readProof(file string) {
	// open file
	f, err := os.Open(file)
	if err != nil {
		fmt.Println("err", err)
	}
	defer f.Close()

	var buff bytes.Buffer
	buff.ReadFrom(f)
	fmt.Println("proof", hex.EncodeToString(buff.Bytes()))

}

func readInput2(fInputPath string) {
	// parse input file
	csvFile, err := os.Open(fInputPath)
	if err != nil {
		fmt.Println("open", err)
		return
	}
	defer csvFile.Close()

	toRead, err := readPublic(csvFile)
	if err != nil {
		fmt.Println("read", err)
		return
	}
	fmt.Println("json code =", toRead)

}

func readPublic(r io.Reader) (string, error) {
	toRead := make(map[string]interface{})
	reader := csv.NewReader(bufio.NewReader(r))
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		} else if len(line) != 3 {
			return "", errors.New("ErrInvalidInputFormat")
		}
		visibility := strings.ToLower(strings.TrimSpace(line[0]))
		name := strings.TrimSpace(line[1])
		value := strings.TrimSpace(line[2])
		if backend.Visibility(visibility) == backend.Public {
			if strings.HasPrefix(value, "0x") {
				bytes, err := hex.DecodeString(value[2:])
				if err != nil {
					return "", err
				}
				toRead[name] = bytes
			}
			toRead[name] = value
		}

	}

	//marshal 可以被unmarshal 和json.decode同时解析
	out, err := json.Marshal(toRead)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(out), nil
}
