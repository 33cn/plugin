package executor

import "errors"

var (
	ErrGetRollupStatus = errors.New("ErrGetRollupStatus")

	ErrInvalidCommitRound   = errors.New("ErrInvalidCommitRound")
	ErrGetValPubs           = errors.New("ErrGetValPubs")
	ErrInvalidValidator     = errors.New("ErrInvalidValidator")
	ErrParentHashNotEqual   = errors.New("ErrParentHashNotEqual")
	ErrInvalidValidatorSign = errors.New("ErrInvalidValidatorSign")
	ErrInvalidBlsPub        = errors.New("ErrInvalidBlsPub")
)
