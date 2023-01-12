package rollup

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/system/consensus"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/crypto/bls"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/stretchr/testify/require"
)

func Test_SubMsg(t *testing.T) {

	r := &RollUp{subChan: make(chan *types.TopicData, 32)}

	r.SubMsg(nil)
	r.SubMsg(&queue.Message{})
	r.SubMsg(&queue.Message{Ty: types.EventReceiveSubData})
	topic := &types.TopicData{}
	r.SubMsg(&queue.Message{Ty: types.EventReceiveSubData, Data: topic})

	topic.Topic = psValidatorSignTopic
	topic.Data = []byte("test")
	r.SubMsg(&queue.Message{Ty: types.EventReceiveSubData, Data: topic})

	select {
	case data := <-r.subChan:
		require.Equal(t, topic.GetData(), data.Data)
	default:
		t.Error("receive topic err")
	}
}

func Test_trySubTopic(t *testing.T) {

	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	q := queue.New("test")
	q.SetConfig(cfg)
	api, err := client.New(q.Client(), nil)
	require.Nil(t, err)

	r := RollUp{base: &consensus.BaseClient{}}
	r.base.SetAPI(api)
	r.client = q.Client()
	topic := "test-topic"
	go func() {
		cli := q.Client()
		cli.Sub("p2p")
		count := 0
		for msg := range cli.Recv() {

			if msg.Ty == types.EventSubTopic {

				data, ok := msg.Data.(*types.SubTopic)
				require.True(t, ok)
				require.Equal(t, topic, data.Topic)
				if count == 0 {
					msg.Reply(&queue.Message{})
				} else {
					msg.Reply(&queue.Message{Data: &types.Reply{IsOk: true}})
				}
			}
			count++
		}
	}()

	r.trySubTopic(topic)
}

func Test_tryPubMsg(t *testing.T) {

	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	q := queue.New("test")
	q.SetConfig(cfg)
	api, err := client.New(q.Client(), nil)
	require.Nil(t, err)

	r := RollUp{base: &consensus.BaseClient{}}
	r.base.SetAPI(api)
	r.client = q.Client()
	topic := "test-topic"
	pubMsg := []byte("test-msg")
	go func() {
		cli := q.Client()
		cli.Sub("p2p")
		count := 0
		for msg := range cli.Recv() {

			if msg.Ty == types.EventPubTopicMsg {

				data, ok := msg.Data.(*types.PublishTopicMsg)
				require.True(t, ok)
				require.Equal(t, topic, data.Topic)
				require.Equal(t, pubMsg, data.Msg)
				if count == 0 {
					msg.Reply(&queue.Message{})
				} else {
					msg.Reply(&queue.Message{Data: &types.Reply{IsOk: true}})
				}
			}
			count++
		}
	}()

	r.tryPubMsg(topic, pubMsg, 0)
}

func Test_handleSubMsg(t *testing.T) {

	r := &RollUp{subChan: make(chan *types.TopicData, 32)}
	val, _, _ := newTestVal()
	r.cache = newCommitCache(0)
	r.val = val
	r.ctx = context.Background()
	go r.handleSubMsg()

	signMsg := &rtypes.ValidatorSignMsg{CommitRound: 1}
	r.subChan <- &types.TopicData{Data: []byte("errData")}
	r.subChan <- &types.TopicData{Data: types.Encode(signMsg)}
	driver := bls.Driver{}
	priv, _ := driver.GenKey()
	signMsg.MsgHash = []byte("test-msg")
	signMsg.PubKey = priv.PubKey().Bytes()
	signMsg.Signature = priv.Sign(signMsg.MsgHash).Bytes()
	r.val.validators[hex.EncodeToString(priv.PubKey().Bytes())] = 0
	r.subChan <- &types.TopicData{Data: types.Encode(signMsg)}
	timeout := time.NewTimer(time.Second * 3)
	signCount := 0
	for signCount == 0 {
		select {
		case <-timeout.C:
			t.Error("handle sub msg timeout")
		default:
		}
		r.cache.lock.RLock()
		signCount = len(r.cache.signList)
		r.cache.lock.RUnlock()
	}

	signSet, ok := r.cache.signList[1]
	require.True(t, ok)
	require.Equal(t, 1, len(signSet.others))
	require.Equal(t, int64(1), signSet.others[0].CommitRound)
}
