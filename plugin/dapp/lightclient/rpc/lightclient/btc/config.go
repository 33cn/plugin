// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package btc

import (
	"os"

	"github.com/btcsuite/btcd/rpcclient"
)

// btc client config
type config struct {
	ID                   string
	Host                 string
	Endpoint             string
	User                 string
	Pass                 string
	DisableTLS           bool
	CertFile             string
	Proxy                string
	ProxyUser            string
	ProxyPass            string
	DisableAutoReconnect bool
	DisableConnectOnNew  bool
	HTTPPostMode         bool
	EnableBCInfoHacks    bool
	ReconnectAttempts    int
}

// transfer to connection config
func (c *config) toConnConfig() *rpcclient.ConnConfig {
	conn := &rpcclient.ConnConfig{}
	conn.Host = c.Host
	conn.Endpoint = c.Endpoint
	conn.User = c.User
	conn.Pass = c.Pass
	conn.DisableTLS = c.DisableTLS
	conn.Proxy = c.Proxy
	conn.ProxyUser = c.ProxyUser
	conn.ProxyPass = c.ProxyPass
	conn.DisableAutoReconnect = c.DisableAutoReconnect
	conn.DisableConnectOnNew = c.DisableConnectOnNew
	conn.HTTPPostMode = c.HTTPPostMode
	conn.EnableBCInfoHacks = c.EnableBCInfoHacks

	if c.CertFile == "" {
		return conn
	}

	certs, err := os.ReadFile(c.CertFile)
	if err != nil {
		panic(err)
	}
	conn.Certificates = certs
	return conn
}
