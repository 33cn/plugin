// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	// ErrNoTicket error type
	ErrNoTicket = errors.New("ErrNoTicket")
	// ErrTicketCount error type
	ErrTicketCount = errors.New("ErrTicketCount")
	// ErrTime error type
	ErrTime = errors.New("ErrTime")
	// ErrTicketClosed err type
	ErrTicketClosed = errors.New("ErrTicketClosed")
	// ErrEmptyMinerTx err type
	ErrEmptyMinerTx = errors.New("ErrEmptyMinerTx")
	// ErrMinerNotPermit err type
	ErrMinerNotPermit = errors.New("ErrMinerNotPermit")
	// ErrMinerAddr err type
	ErrMinerAddr = errors.New("ErrMinerAddr")
	// ErrModify err type
	ErrModify = errors.New("ErrModify")
	// ErrMinerTx err type
	ErrMinerTx = errors.New("ErrMinerTx")
)
