package types

const (
	// MaxAssetSymbolLength is the maximum byte length of an asset's symbol.
	// This byte length is equivalent to character count for single-byte
	// UTF-8 characters.
	MaxAssetSymbolLength = 16

	// MetaHashLen is the length of the metadata hash.
	MetaHashLen = 32
	// DefaultPrecision default Precision
	DefaultPrecision = int64(1e8)
	// MaxAssetAmount max amount
	MaxAssetAmount = 1e8 * 900 * DefaultPrecision // 900 亿
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
