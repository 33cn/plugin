package executor

import "errors"

var (
	ErrGetRollupStatus        = errors.New("ErrGetRollupStatus")
	ErrNullCommitData         = errors.New("ErrNullCommitData")
	ErrInvalidCommitRound     = errors.New("ErrInvalidCommitRound")
	ErrGetValPubs             = errors.New("ErrGetValPubs")
	ErrInvalidValidator       = errors.New("ErrInvalidValidator")
	ErrOutOfOrderCommit       = errors.New("ErrOutOfOrderCommit")
	ErrInvalidValidatorSign   = errors.New("ErrInvalidValidatorSign")
	ErrInvalidBlsPub          = errors.New("ErrInvalidBlsPub")
	ErrInvalidTxAggregateSign = errors.New("ErrInvalidTxAggregateSign")
)
