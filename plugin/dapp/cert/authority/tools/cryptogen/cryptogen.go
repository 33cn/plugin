// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/33cn/chain33/types"
	_ "github.com/33cn/plugin/plugin/crypto/init"
	"github.com/33cn/plugin/plugin/dapp/cert/authority/tools/cryptogen/generator"
	ca "github.com/33cn/plugin/plugin/dapp/cert/authority/tools/cryptogen/generator/impl"
	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

const (
	// CANAME 默认CA名称
	CANAME = "ca"
	// CONFIGFILENAME 配置文件名
	CONFIGFILENAME = "chain33.cryptogen.toml"
	// OUTPUTDIR 证书文件输出路径
	OUTPUTDIR = "./authdir/crypto"
	// ORGNAME 默认组织名
	ORGNAME = "Chain33"
)

// Config 证书生成工具配置
type Config struct {
	Name     []string
	SignType string
}

var (
	cmd = &cobra.Command{
		Use:   "cryptogen [-f configfile] [-o output directory]",
		Short: "chain33 crypto tool for generating key and certificate",
		Run:   generate,
	}
	cfg Config
)

func initCfg(path string) *Config {
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	return &cfg
}

func main() {
	cmd.Flags().StringP("configfile", "f", CONFIGFILENAME, "config file for users")
	cmd.Flags().StringP("outputdir", "o", OUTPUTDIR, "output diraction for key and certificate")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generate(cmd *cobra.Command, args []string) {
	configfile, _ := cmd.Flags().GetString("configfile")
	outputdir, _ := cmd.Flags().GetString("outputdir")

	initCfg(configfile)
	fmt.Println(cfg.Name)

	generateUsers(outputdir, ORGNAME)
}

func generateUsers(baseDir string, orgName string) {
	fmt.Printf("generateUsers\n")
	fmt.Println(baseDir)

	err := os.RemoveAll(baseDir)
	if err != nil {
		fmt.Printf("Clean directory %s error", baseDir)
		os.Exit(1)
	}

	caDir := filepath.Join(baseDir, "cacerts")

	signType := types.GetSignType("cert", cfg.SignType)
	if signType == types.Invalid {
		fmt.Printf("Invalid sign type:%s", cfg.SignType)
		return
	}

	signCA, err := ca.NewCA(caDir, CANAME, signType)
	if err != nil {
		fmt.Printf("Error generating signCA:%s", err.Error())
		os.Exit(1)
	}

	generateNodes(baseDir, signCA, orgName)
}

func generateNodes(baseDir string, signCA generator.CAGenerator, orgName string) {
	for _, name := range cfg.Name {
		userDir := filepath.Join(baseDir, name)
		fileName := fmt.Sprintf("%s@%s", name, orgName)
		err := signCA.GenerateLocalUser(userDir, fileName)
		if err != nil {
			fmt.Printf("Error generating local user")
			os.Exit(1)
		}
	}
}
