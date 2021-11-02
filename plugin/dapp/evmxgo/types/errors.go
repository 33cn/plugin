package types

import "errors"

var (
	// ErrEvmxgoSymbolNotExist error evmxgo symbol not exist
	ErrEvmxgoSymbolNotExist = errors.New("ErrEvmxgoSymbolNotExist")
	// ErrEvmxgoSymbolNotAllowedMint error evmxgo symbol not allowed mint
	ErrEvmxgoSymbolNotAllowedMint = errors.New("ErrEvmxgoSymbolNotAllowedMint")
	// ErrEvmxgoSymbolNotConfigValue error evmxgo symbol not config value
	ErrEvmxgoSymbolNotConfigValue = errors.New("ErrEvmxgoSymbolNotConfigValue")
	// ErrNotCorrectBridgeTokenAddress error Not Correct BridgeToken Address
	ErrNotCorrectBridgeTokenAddress = errors.New("ErrNotCorrectBridgeTokenAddress")
)
