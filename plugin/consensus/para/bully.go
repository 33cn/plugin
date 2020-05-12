package para

import (
	"errors"
	"sync"
	"time"

	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

const (
	ELECTION = iota
	COORDINATOR
	OK
	CLOSE
)

// Bully is a `struct` representing a single node used by the `Bully Algorithm`.
//
// NOTE: More details about the `Bully algorithm` can be found here
// https://en.wikipedia.org/wiki/Bully_algorithm .
type Bully struct {
	ID           string
	coordinator  string
	nodegroup    map[string]bool
	inNodeGroup  bool
	mu           *sync.RWMutex
	receiveChan  chan *pt.ElectionMsg
	electionChan chan *pt.ElectionMsg
	paraClient   *client
	qClient      queue.Client
	wg           *sync.WaitGroup
	quit         chan struct{}
}

func NewBully(para *client, ID string, wg *sync.WaitGroup) (*Bully, error) {
	b := &Bully{
		paraClient:   para,
		ID:           ID,
		nodegroup:    make(map[string]bool),
		electionChan: make(chan *pt.ElectionMsg, 1),
		receiveChan:  make(chan *pt.ElectionMsg),
		wg:           wg,
		mu:           &sync.RWMutex{},
		quit:         make(chan struct{}),
	}
	return b, nil
}

func (b *Bully) SetParaAPI(cli queue.Client) {
	b.qClient = cli
}

func (b *Bully) UpdateValidNodes(nodes []string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nodegroup = make(map[string]bool)
	for _, n := range nodes {
		plog.Info("bully node update", "node", n)
		b.nodegroup[n] = true
	}

	//退出nodegroup
	if b.inNodeGroup && !b.nodegroup[b.ID] {
		_ = b.Send("", CLOSE, nil)
	}
	b.inNodeGroup = b.nodegroup[b.ID]
}

func (b *Bully) isValidNode(ID string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.nodegroup[ID]
}

func (b *Bully) Close() {
	close(b.quit)
}

func (b *Bully) Receive(msg *pt.ElectionMsg) {
	plog.Info("bully rcv", "type", msg.Type)
	switch msg.Type {
	case CLOSE:
		if msg.PeerID == b.Coordinator() {
			b.SetCoordinator(b.ID)
			b.Elect()
		}
	case OK:
		if msg.ToID != b.ID {
			return
		}
		select {
		case b.electionChan <- msg:
			return
		case <-time.After(200 * time.Millisecond):
			return
		}
	default:
		b.receiveChan <- msg
	}

}

func (b *Bully) sendMsg(ty int64, data interface{}) error {
	msg := b.qClient.NewMessage("p2p", ty, data)
	err := b.qClient.Send(msg, true)
	if err != nil {
		return err
	}
	resp, err := b.qClient.Wait(msg)
	if err != nil {
		return err
	}
	if resp.GetData().(*types.Reply).IsOk {
		return nil
	}
	return errors.New(string(resp.GetData().(*types.Reply).GetMsg()))
}

func (b *Bully) Send(toId string, msgTy int32, data []byte) error {
	act := &pt.ParaP2PSubMsg{Ty: P2pSubElectMsg}
	act.Value = &pt.ParaP2PSubMsg_Election{Election: &pt.ElectionMsg{ToID: toId, PeerID: b.ID, Type: msgTy, Data: data}}
	plog.Info("bull sendmsg")
	err := b.paraClient.SendPubP2PMsg(types.Encode(act))
	//err := b.sendMsg(types.EventPubTopicMsg, &types.PublishTopicMsg{Topic: "consensus", Msg: types.Encode(act)})
	plog.Info("bully ret")
	return err

}

// SetCoordinator sets `ID` as the new `b.coordinator` if `ID` is greater than
// `b.coordinator` or equal to `b.ID`.
func (b *Bully) SetCoordinator(ID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ID > b.coordinator || ID == b.ID {
		b.coordinator = ID
	}
}

// Coordinator returns `b.coordinator`.
//
// NOTE: This function is thread-safe.
func (b *Bully) Coordinator() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.coordinator
}

func (b *Bully) IsSelfCoordinator() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.ID == b.coordinator
}

// Elect handles the leader election mechanism of the `Bully algorithm`.
func (b *Bully) Elect() {
	_ = b.Send("", ELECTION, nil)

	select {
	case <-b.electionChan:
		return
	case <-time.After(time.Second):
		b.SetCoordinator(b.ID)
		_ = b.Send("", COORDINATOR, nil)
		return
	}
}

func (b *Bully) Run() {
	defer b.wg.Done()
	var feedDog = false
	var feedDogTiker <-chan time.Time
	var watchDogTiker <-chan time.Time
	plog.Info("bully init")
	onceTimer := time.NewTimer(time.Minute)

out:
	for {
		select {
		case msg := <-b.receiveChan:
			switch msg.Type {
			case ELECTION:
				if msg.PeerID < b.ID {
					_ = b.Send(msg.PeerID, OK, nil)
					b.Elect()
				}
			case COORDINATOR:
				if !b.isValidNode(msg.PeerID) {
					continue
				}
				b.SetCoordinator(msg.PeerID)
				if b.coordinator < b.ID {
					b.Elect()
				}
				feedDog = true

			}
		case <-onceTimer.C:
			feedDogTiker = time.NewTicker(20 * time.Second).C

			watchDogTiker = time.NewTicker(time.Minute).C

		case <-feedDogTiker:
			plog.Info("bully feed dog tiker", "is", b.IsSelfCoordinator(), "valid", b.isValidNode(b.ID))
			//leader需要定期喂狗
			if b.IsSelfCoordinator() && b.isValidNode(b.ID) {
				_ = b.Send("", COORDINATOR, nil)
			}

		case <-watchDogTiker:
			//至少1分钟内要收到leader喂狗消息，否则认为leader挂了，重新选举
			if !feedDog {
				b.Elect()
				plog.Info("bully watchdog triger")
			}
			feedDog = false

		case <-b.quit:
			break out
		}
	}

}
