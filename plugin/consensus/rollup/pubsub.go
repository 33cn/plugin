package rollup

import (
	"time"

	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/system/consensus"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

const (
	psValidatorSignTopic = "rollup/valSignMsg/1.0"
)

func (r *RollUp) SubMsg(msg *queue.Message) {

	if msg == nil {
		rlog.Error("SubMsg", "err", "receive nil msg")
		return
	}
	data, ok := msg.Data.(*types.TopicData)
	if msg.Ty != types.EventReceiveSubData || !ok {
		rlog.Error("SubMsg", "ty", msg.Ty, "ok", ok, "err", "receive invalid msg")
		return
	}

	if data.Topic != psValidatorSignTopic {
		rlog.Error("SubMsg", "topic", data.Topic, "receive invalid topic")
		return
	}

	r.subChan <- data
}

func (r *RollUp) trySubTopic(topic string) {

	data := &types.SubTopic{Topic: topic, Module: consensus.ModuleName}

	for {
		err := r.sendP2PMsg(types.EventSubTopic, data)
		if err == nil {
			break
		}
		rlog.Debug("trySubTopic", "err", err)
		time.Sleep(time.Second)
	}
}

func (r *RollUp) tryPubMsg(topic string, msg []byte, commitRound int64) {

	data := &types.PublishTopicMsg{Topic: topic, Msg: msg}
	tryCount := 0
	for {
		tryCount++
		err := r.sendP2PMsg(types.EventPubTopicMsg, data)
		if err == nil || tryCount >= 3 {
			break
		}
		rlog.Error("tryPubMsg", "topic", topic, "round", commitRound, "err", err)
		time.Sleep(time.Second)
	}
}

func (r *RollUp) handleSubMsg() {

	for {

		select {

		case <-r.ctx.Done():
			return

		case data := <-r.subChan:

			signMsg := &rtypes.ValidatorSignMsg{}
			err := types.Decode(data.GetData(), signMsg)
			if err != nil {
				rlog.Error("handleSubMsg", "from", data.From, "decode err", err)
				break
			}

			if !r.val.validateSignMsg(signMsg) {
				rlog.Error("handleSubMsg", "err", "validate sign err")
				break
			}

			r.cache.addValidatorSign(false, signMsg)
		}

	}
}
