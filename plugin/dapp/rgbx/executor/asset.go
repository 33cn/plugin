package executor

import (
	"errors"
	"strings"
)

var (
	ErrInvalidSymbolLength         = errors.New("invalid asset symbol length")
	ErrInvalidAssetAmount          = errors.New("invalid asset amount")
	ErrInvalidMetaHashLength       = errors.New("invalid meta hash length")
	ErrNilGenesisOut               = errors.New("nil genesis output")
	ErrDuplicateAssetSymbol        = errors.New("duplicate asset symbol")
	ErrAssetNotExist               = errors.New("asset not exist")
	ErrDecodeBtcTx                 = errors.New("decode btc tx error")
	ErrPendingTxNotExist           = errors.New("pending tx not exist")
	ErrTxAlreadyConfirmed          = errors.New("tx already confirmed")
	ErrConfirmedHashNotEqual       = errors.New("confirmed hash not equal")
	ErrSpendingInputNotEqual       = errors.New("spending input not equal")
	ErrOpRetOutputPkScriptNotEqual = errors.New("ErrOpRetOutputPkScriptNotEqual")
	ErrInvalidCommitAddress        = errors.New("ErrInvalidCommitAddress")
)

const (
	// MaxAssetSymbolLength is the maximum byte length of an asset's symbol.
	// This byte length is equivalent to character count for single-byte
	// UTF-8 characters.
	MaxAssetSymbolLength = 16

	// MetaHashLen is the length of the metadata hash.
	MetaHashLen = 32
	// MaxAssetAmount max amount
	MaxAssetAmount = 1e8 * 10000
)

// Type denotes the asset types supported by the Taproot Asset protocol.
type Type uint8

const (
	// Normal is an asset that can be represented in multiple units,
	// resembling a divisible asset.
	Normal Type = 0

	// Collectible is a unique asset, one that cannot be represented in
	// multiple units.
	Collectible Type = 1
)

// String returns a human-readable description of the type.
func (t Type) String() string {
	switch t {
	case Normal:
		return "Normal"
	case Collectible:
		return "Collectible"
	default:
		return "<Unknown>"
	}
}

func formatSymbol(symbol string) string {
	return strings.ToUpper(symbol)
}
