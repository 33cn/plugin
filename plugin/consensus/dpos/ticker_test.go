package dpos

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTicker(t *testing.T) {
	ticker := NewTimeoutTicker()
	ticker.Start()

	ti := timeoutInfo{
		Duration: time.Millisecond * time.Duration(2000),
		State:    InitStateType,
	}
	fmt.Println("timeoutInfo:", ti.String())

	now := time.Now().Unix()
	ticker.ScheduleTimeout(ti)
	ti2 := <-ticker.Chan()
	end := time.Now().Unix()
	fmt.Println("timeoutInfo2:", ti2.String())

	time.Sleep(time.Second * 3)
	ticker.Stop()
	assert.True(t, end-now >= 2)
	fmt.Println("TestTicker ok", end-now)
}
