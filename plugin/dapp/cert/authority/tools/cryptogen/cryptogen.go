// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/33cn/chain33/common/log/log15"
	"gopkg.in/yaml.v2"

	"github.com/33cn/chain33/types"
	_ "github.com/33cn/plugin/plugin/crypto/init"
	"github.com/33cn/plugin/plugin/dapp/cert/authority/tools/cryptogen/generator"
	"github.com/spf13/cobra"
)

const (
	// CONFIGFILENAME 配置文件名
	CONFIGFILENAME = "chain33.cryptogen.yaml"
	// OUTPUTDIR 证书文件输出路径
	OUTPUTDIR = "./authdir/crypto"
)

var (
	cmd = &cobra.Command{
		Use:   "cryptogen [-f configfile] [-o output directory]",
		Short: "chain33 crypto tool for generating key and certificate",
		Run:   generate,
	}
	cfg    *generator.GenConfig
	logger = log.New("module", "main")
)

func initCfg(path string) (*generator.GenConfig, error) {
	conf := &generator.GenConfig{}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	yaml.NewDecoder(f).Decode(conf)
	return conf, nil
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

	var err error
	cfg, err = initCfg(configfile)
	if err != nil {
		panic(err)
	}

	generateCert(outputdir)

}

func generateCert(baseDir string) {
	logger.Info("generate certs", "dir", baseDir)

	err := os.RemoveAll(baseDir)
	if err != nil {
		logger.Error("clean directory", "error", err.Error())
		os.Exit(1)
	}

	caDir := filepath.Join(baseDir, "cacerts")

	signType := types.GetSignType("cert", cfg.SignType)
	if signType == types.Invalid {
		logger.Error("invalid sign type", "type", cfg.SignType)
		return
	}

	signCA, err := generator.NewCA(caDir, &cfg.Root, signType)
	if err != nil {
		logger.Error("generating signCA", "error", err.Error())
		os.Exit(1)
	}

	for _, org := range cfg.Root.User {
		generateOrgs(baseDir, signCA, cfg.GetOrgCertConfig(org.Name))
	}

}

func generateOrgs(baseDir string, signCA generator.CAGenerator, orgCfg *generator.CertConfig) {
	orgDir := filepath.Join(baseDir, orgCfg.Name)
	fileName := fmt.Sprintf("%s@%s", orgCfg.Name, cfg.Root.Name)
	orgSignCA, err := signCA.GenerateLocalOrg(orgDir, fileName, orgCfg)
	if err != nil {
		logger.Error("generating local org", "org", orgCfg.Name, "error", err.Error())
		os.Exit(1)
	}

	for _, user := range orgCfg.User {
		userDir := filepath.Join(orgDir, user.Name)
		fileName = fmt.Sprintf("%s@%s", user.Name, orgCfg.Name)
		err := orgSignCA.GenerateLocalUser(userDir, fileName)
		if err != nil {
			logger.Error("generating local user", "user", user.Name, "error", err.Error())
			os.Exit(1)
		}
	}
}
