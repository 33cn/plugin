package types

import "errors"

var (
	ErrContractExist       = errors.New("contract exist")
	ErrInvalidWasm         = errors.New("invalid wasm code")
	ErrCodeOversize        = errors.New("code oversize")
	ErrInvalidMethod       = errors.New("invalid method")
	ErrInvalidContractName = errors.New("invalid contract name")
	ErrInvalidParam        = errors.New("invalid parameters")
	ErrUnknown             = errors.New("unknown error")
)
