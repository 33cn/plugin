package lightclient

import (
	"context"
	"sync"

	"github.com/33cn/chain33/client"
)

var (
	clients = make(map[string]func() Lighter)
	lock    sync.Mutex
)

// Lighter light client interface
type Lighter interface {
	// Init init client context
	Init(ctx context.Context, api client.QueueProtocolAPI, cfg *Config) error
	// Start client routine
	Start()
}

// Config light client config
type Config struct {
	EnableClients []string    `json:"clients,omitempty"`
	CommitAddr    string      `json:"commitAddr,omitempty"`
	Neutrino      interface{} `json:"neutrino,omitempty"`
	Btc           interface{} `json:"btc,omitempty"`
}

// Register register light client
func Register(name string, newCli func() Lighter) {

	lock.Lock()
	defer lock.Unlock()
	_, ok := clients[name]
	if ok {
		panic("Register duplicate client, name=" + name)
	}

	clients[name] = newCli
}

// New  create light client
func New(name string) Lighter {
	create := clients[name]
	if create != nil {
		return create()
	}
	return nil
}
