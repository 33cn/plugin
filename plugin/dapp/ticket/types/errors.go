// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	ErrNoTicket       = errors.New("ErrNoTicket")
	ErrTicketCount    = errors.New("ErrTicketCount")
	ErrTime           = errors.New("ErrTime")
	ErrTicketClosed   = errors.New("ErrTicketClosed")
	ErrEmptyMinerTx   = errors.New("ErrEmptyMinerTx")
	ErrMinerNotPermit = errors.New("ErrMinerNotPermit")
	ErrMinerAddr      = errors.New("ErrMinerAddr")
	ErrModify         = errors.New("ErrModify")
	ErrMinerTx        = errors.New("ErrMinerTx")
)
