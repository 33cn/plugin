package types

import "errors"

var (
	ErrInvalidEthContractAddress = errors.New("ErrInvalidEthContractAddress")
	ErrUnpack                    = errors.New("ErrUnpackResultWithABI")
	ErrPack                      = errors.New("ErrPackParamererWithABI")
	ErrEmptyAddress              = errors.New("ErrEmptyAddress")
	ErrAddress4Eth               = errors.New("symbol \"eth\" must have null address set as token address")
	ErrPublicKeyType             = errors.New("ErrPublicKeyType")
	ErrInvalidContractAddress    = errors.New("ErrInvalidTargetContractAddress")
	ErrNoValidatorConfigured     = errors.New("ErrNoValidatorConfigured")
	ErrMulSignNotDeployed        = errors.New("Contract multi-sign haven't been deployed yet")
	ErrGetTransactionHash        = errors.New("Failed to get tx hash from mulsign")
)
