package executor

import "errors"

var (
	ErrDecodeAction         = errors.New("ErrDecodeAction")
	ErrNilBtcHeader         = errors.New("ErrNilBtcHeader")
	ErrBtcGetLastHeader     = errors.New("ErrBtcGetLastHeader")
	ErrIllegalCommitAddress = errors.New("ErrIllegalCommitAddress")
	ErrBtcHeaderDisorder    = errors.New("ErrBtcHeaderDisorder")

	ErrBtcTargetBits        = errors.New("ErrBtcTargetBits")
	ErrBtcHeaderTimeTooOld  = errors.New("ErrBtcHeaderTimeTooOld")
	ErrBtcHeaderTimeTooNew  = errors.New("ErrBtcHeaderTimeTooNew")
	ErrBtcHeaderInvalidTime = errors.New("ErrBtcHeaderInvalidTime")
	ErrToBtcWireHeader      = errors.New("ErrToBtcWireHeader")
	ErrInvalidBtcBlockHash  = errors.New("ErrInvalidBtcBlockHash")
	ErrBtcHeaderVerify      = errors.New("ErrBtcHeaderVerify")
)
