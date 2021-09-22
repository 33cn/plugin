package test

//
//import (
//	"encoding/hex"
//	"fmt"
//	"math/big"
//	"strings"
//	"testing"
//
//	"github.com/ethereum/go-ethereum/accounts/abi"
//	"github.com/ethereum/go-ethereum/common"
//)
//
//func TestUnpackEvent(t *testing.T) {
//	const abiJSON = `[{"constant":false,"inputs":[{"name":"memo","type":"bytes"}],"name":"receive","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"name":"sender","type":"address"},{"indexed":false,"name":"amount","type":"uint256"},{"indexed":false,"name":"memo","type":"bytes"}],"name":"received","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"sender","type":"address"}],"name":"receivedAddr","type":"event"}]`
//	abi, err := abi.JSON(strings.NewReader(abiJSON))
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	const hexdata = `000000000000000000000000376c47978271565f56deb45495afa69e59c16ab200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000000158`
//	data, err := hex.DecodeString(hexdata)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if len(data)%32 == 0 {
//		t.Errorf("len(data) is %d, want a non-multiple of 32", len(data))
//	}
//
//	type ReceivedEvent struct {
//		Sender common.Address
//		Amount *big.Int
//		Memo   []byte
//	}
//	var ev ReceivedEvent
//
//	err = abi.Unpack(&ev, "received", data)
//	if err != nil {
//		t.Error(err)
//	}
//
//	fmt.Printf("\nReceivedEvent sender:%s", ev.Sender.String())
//
//	type ReceivedAddrEvent struct {
//		Sender common.Address
//	}
//	var receivedAddrEv ReceivedAddrEvent
//	err = abi.Unpack(&receivedAddrEv, "receivedAddr", data)
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Printf("\nreceivedAddrEv=%s\n\n\n", receivedAddrEv.Sender.String())
//}
