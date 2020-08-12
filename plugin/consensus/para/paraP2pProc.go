// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"
)

//新增需要保证顺序
const (
	P2pSubCommitTx      = 1
	P2pSubLeaderSyncMsg = 2
	moduleName          = "consensus"
)

func (client *client) sendP2PMsg(ty int64, data interface{}) ([]byte, error) {
	msg := client.GetQueueClient().NewMessage("p2p", ty, data)
	err := client.GetQueueClient().Send(msg, true)
	if err != nil {
		return nil, errors.Wrapf(err, "ty=%d", ty)
	}
	resp, err := client.GetQueueClient().Wait(msg)
	if err != nil {
		return nil, errors.Wrapf(err, "wait ty=%d", ty)
	}

	if resp.GetData().(*types.Reply).IsOk {
		return resp.GetData().(*types.Reply).Msg, nil
	}
	return nil, errors.Wrapf(types.ErrInvalidParam, "resp msg=%s", string(resp.GetData().(*types.Reply).GetMsg()))
}

// p2p订阅消息
func (client *client) SendPubP2PMsg(topic string, msg []byte) error {
	data := &types.PublishTopicMsg{Topic: topic, Msg: msg}
	_, err := client.sendP2PMsg(types.EventPubTopicMsg, data)
	return err
}

func (client *client) SendSubP2PTopic(topic string) error {
	data := &types.SubTopic{Topic: topic, Module: moduleName}
	_, err := client.sendP2PMsg(types.EventSubTopic, data)
	return err
}

func (client *client) SendRmvP2PTopic(topic string) error {
	data := &types.RemoveTopic{Topic: topic, Module: moduleName}
	_, err := client.sendP2PMsg(types.EventRemoveTopic, data)
	return err

}

func (client *client) SendFetchP2PTopic() (*types.TopicList, error) {
	data := &types.FetchTopicList{Module: moduleName}
	msg, err := client.sendP2PMsg(types.EventFetchTopics, data)
	if err != nil {
		return nil, errors.Wrap(err, "reply fail")
	}
	var reply types.TopicList
	err = types.Decode(msg, &reply)
	if err != nil {
		return nil, errors.Wrap(err, "decode fail")
	}
	return &reply, err

}
