package runtime

import (
	"errors"
	"fmt"
)

// List evm execution errors
var (
	// ErrInvalidSubroutineEntry means that a BEGINSUB was reached via iteration,
	// as opposed to from a JUMPSUB instruction
	ErrOutOfGas          = errors.New("out of gas")
	ErrExecutionReverted = errors.New("execution reverted")
	ErrGasUintOverflow   = errors.New("gas uint64 overflow")
)

// ErrStackUnderflow wraps an evm error when the items on the stack less
// than the minimal requirement.
type ErrStackUnderflow struct {
	stackLen int
	required int
}

func (e *ErrStackUnderflow) Error() string {
	return fmt.Sprintf("stack underflow (%d <=> %d)", e.stackLen, e.required)
}

// ErrStackOverflow wraps an evm error when the items on the stack exceeds
// the maximum allowance.
type ErrStackOverflow struct {
	stackLen int
	limit    int
}

func (e *ErrStackOverflow) Error() string {
	return fmt.Sprintf("stack limit reached %d (%d)", e.stackLen, e.limit)
}

// ErrInvalidOpCode wraps an evm error when an invalid opcode is encountered.
type ErrInvalidOpCode struct {
	opcode OpCode
}

func (e *ErrInvalidOpCode) Error() string { return fmt.Sprintf("invalid opcode: %s", e.opcode) }
