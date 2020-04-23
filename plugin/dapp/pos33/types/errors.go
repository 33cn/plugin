// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	// ErrNoPos33Ticket error type
	ErrNoPos33Ticket = errors.New("ErrNoPos33Ticket")
	// ErrPos33TicketCount error type
	ErrPos33TicketCount = errors.New("ErrPos33TicketCount")
	// ErrTime error type
	ErrTime = errors.New("ErrTime")
	// ErrPos33TicketClosed err type
	ErrPos33TicketClosed = errors.New("ErrPos33TicketClosed")
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
	// ErrNoVrf err type
	ErrNoVrf = errors.New("ErrNoVrf")
	// ErrVrfVerify err type
	ErrVrfVerify = errors.New("ErrVrfVerify")
)
