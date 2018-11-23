// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package relayd

import (
	"io/ioutil"
	"path/filepath"

	"github.com/33cn/chain33/types"
	"github.com/BurntSushi/toml"
	"github.com/btcsuite/btcd/rpcclient"
)

// Config relayd toml config
type Config struct {
	Title          string
	Watch          bool
	Tick33         int
	TickBTC        int
	BtcdOrWeb      int
	SyncSetup      uint64
	SyncSetupCount uint64
	Chain33        Chain33
	FirstBtcHeight uint64
	Btcd           Btcd
	Log            types.Log
	Auth           Auth
}

// Btcd adapt to btcd
type Btcd struct {
	ID                   string
	Host                 string
	Endpoint             string
	User                 string
	Pass                 string
	DisableTLS           bool
	CertPath             string
	Proxy                string
	ProxyUser            string
	ProxyPass            string
	DisableAutoReconnect bool
	DisableConnectOnNew  bool
	HTTPPostMode         bool
	EnableBCInfoHacks    bool
	ReconnectAttempts    int
}

// Auth auth key struct
type Auth struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	Address    string `json:"address"`
}

// BitConnConfig btc connect config
func (b *Btcd) BitConnConfig() *rpcclient.ConnConfig {
	conn := &rpcclient.ConnConfig{}
	conn.Host = b.Host
	conn.Endpoint = b.Endpoint
	conn.User = b.User
	conn.Pass = b.Pass
	conn.DisableTLS = b.DisableTLS
	conn.Proxy = b.Proxy
	conn.ProxyUser = b.ProxyUser
	conn.ProxyPass = b.ProxyPass
	conn.DisableAutoReconnect = b.DisableAutoReconnect
	conn.DisableConnectOnNew = b.DisableConnectOnNew
	conn.HTTPPostMode = b.HTTPPostMode
	conn.EnableBCInfoHacks = b.EnableBCInfoHacks
	certs, err := ioutil.ReadFile(filepath.Join(b.CertPath, "rpc.cert"))
	if err != nil {
		panic(err)
	}
	conn.Certificates = certs
	return conn
}

// Chain33 define adapt to chain33 relay exec
type Chain33 struct {
	ID                   string
	Host                 string
	User                 string
	Pass                 string
	DisableAutoReconnect bool
	ReconnectAttempts    int
}

// NewConfig create a new config
func NewConfig(path string) *Config {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		panic(err)
	}
	return &cfg
}
