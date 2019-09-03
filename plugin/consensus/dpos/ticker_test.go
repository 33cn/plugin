package dpos

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	ticker := NewTimeoutTicker()
	ticker.Start()

	ti := timeoutInfo{
		Duration: time.Second * time.Duration(2),
		State:    InitStateType,
	}

	now := time.Now().Unix()
	ticker.ScheduleTimeout(ti)
	<-ticker.Chan()
	end := time.Now().Unix()

	ticker.Stop()
	assert.True(t, end-now == 2)

}
