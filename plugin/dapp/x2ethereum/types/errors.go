package types

import "errors"

//err ...
var (
	ErrInvalidClaimType              = errors.New("invalid claim type provided")
	ErrInvalidEthSymbol              = errors.New("invalid symbol provided, symbol \"eth\" must have null address set as token contract address")
	ErrInvalidChainID                = errors.New("invalid ethereum chain id")
	ErrInvalidEthAddress             = errors.New("invalid ethereum address provided, must be a valid hex-encoded Ethereum address")
	ErrInvalidEthNonce               = errors.New("invalid ethereum nonce provided, must be >= 0")
	ErrInvalidAddress                = errors.New("invalid Chain33 address")
	ErrInvalidIdentifier             = errors.New("invalid identifier provided, must be a nonempty string")
	ErrProphecyNotFound              = errors.New("prophecy with given id not found")
	ErrProphecyGet                   = errors.New("prophecy with given id find error")
	ErrinternalDB                    = errors.New("internal error serializing/deserializing prophecy")
	ErrNoClaims                      = errors.New("cannot create prophecy without initial claim")
	ErrInvalidClaim                  = errors.New("claim cannot be empty string")
	ErrProphecyFinalized             = errors.New("prophecy already finalized")
	ErrProphecyFailed                = errors.New("prophecy failed so you can't burn this prophecy")
	ErrDuplicateMessage              = errors.New("already processed message from validator for this id")
	ErrMinimumConsensusNeededInvalid = errors.New("minimum consensus proportion of validator staking power must be > 0 and <= 1")
	ErrInvalidValidator              = errors.New("validator is invalid")
	ErrUnknownAddress                = errors.New("module account does not exist")
	ErrPowerIsNotEnough              = errors.New("remove power is more than which this address saves")
	ErrAddressNotExist               = errors.New("this address doesn't exist in DB")
	ErrInvalidProphecyID             = errors.New("Prophecy ID is invalid")
	ErrAddressExists                 = errors.New("This address already exists")
	ErrInvalidAdminAddress           = errors.New("This address is not admin address")
	ErrClaimInconsist                = errors.New("This claim does not consist with others")
	ErrInvalidPower                  = errors.New("This power is invalid")
)

//common
var (
	ErrSetKV = errors.New("Set KV error")
)
