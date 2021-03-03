package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

const (
	// QueryGetAssetSupply command for getting info about an asset's supply
	QueryGetAssetSupply = "supply"
	// QueryGetAssetSupplies command for getting a list of asset supplies
	QueryGetAssetSupplies = "supplies"
	// QueryGetAtomicSwap command for getting info about an atomic swap
	QueryGetAtomicSwap = "swap"
	// QueryGetAtomicSwaps command for getting a list of atomic swaps
	QueryGetAtomicSwaps = "swaps"
	// QueryGetParams command for getting module params
	QueryGetParams = "parameters"
)

// NewQueryAssetSupply creates a new QueryAssetSupply
func NewQueryAssetSupply(denom string) QueryAssetSupply {
	return QueryAssetSupply{
		Denom: denom,
	}
}

// NewQueryAssetSupplies creates a new QueryAssetSupplies
func NewQueryAssetSupplies(page int, limit int) QueryAssetSupplies {
	return QueryAssetSupplies{
		Page:  page,
		Limit: limit,
	}
}

// NewQueryAtomicSwapByID creates a new QueryAtomicSwapByID
func NewQueryAtomicSwapByID(swapBytes tmbytes.HexBytes) QueryAtomicSwapByID {
	return QueryAtomicSwapByID{
		SwapID: swapBytes,
	}
}

// NewQueryAtomicSwaps creates a new instance of QueryAtomicSwaps
func NewQueryAtomicSwaps(page, limit int, involve sdk.AccAddress, expiration int64, status SwapStatus,
	direction SwapDirection) QueryAtomicSwaps {
	return QueryAtomicSwaps{
		Page:       page,
		Limit:      limit,
		Involve:    involve.String(),
		Expiration: expiration,
		Status:     status,
		Direction:  direction,
	}
}
