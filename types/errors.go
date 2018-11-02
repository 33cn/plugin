package types

import "errors"

var (
	ErrUnfreezeBeforeDue = errors.New("ErrUnfreezeBeforeDue")
	ErrUnfreezeEmptied   = errors.New("ErrUnfreezeEmptied")
	ErrUnfreezeMeans     = errors.New("ErrUnfreezeMeans")
	ErrUnfreezeID        = errors.New("ErrUnfreezeID")
	ErrNoUnfreezeItem    = errors.New("ErrNoUnfreezeItem")
)