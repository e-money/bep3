package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "bep3"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute is the querier route for bep3
	QuerierRoute = ModuleName

	// DefaultParamspace default namestore
	DefaultParamspace = ModuleName

	// DefaultLongtermStorageDuration is 1 week
	DefaultLongtermStorageDuration uint64 = 7 * 24 * 60 * 60
)

// Key prefixes
var (
	// ModulePermissionsUpgradeTime is the block time after which the bep3 module account's permissions are synced with the supply module.
	ModulePermissionsUpgradeTime time.Time = time.Date(2020, 11, 3, 10, 0, 0, 0, time.UTC)

	AtomicSwapKeyPrefix             = []byte{0x00} // prefix for keys that store AtomicSwaps
	AtomicSwapByBlockPrefix         = []byte{0x01} // prefix for keys of the AtomicSwapsByBlock index
	AtomicSwapLongtermStoragePrefix = []byte{0x02} // prefix for keys of the AtomicSwapLongtermStorage index
	AssetSupplyPrefix               = []byte{0x03}
	PreviousBlockTimeKey            = []byte{0x04}
)

// GetAtomicSwapByHeightKey is used by the AtomicSwapByBlock index and AtomicSwapLongtermStorage index
func GetAtomicSwapByHeightKey(height uint64, swapID []byte) []byte {
	return append(GetHeightSortableKey(height), swapID...)
}

func GetTimestampSortableKey(timestamp int64) []byte {
	t := time.Unix(timestamp, 0).UTC()
	return sdk.FormatTimeBytes(t)
}

func GetHeightSortableKey(height uint64) []byte {
	return sdk.Uint64ToBigEndian(height)
}

// GetAtomicSwapByTimestampKey is used by the AtomicSwapByTimestamp and
// AtomicSwapLongTermStorage index to generate sortable timestamp keys.
func GetAtomicSwapByTimestampKey(timestamp int64, swapID []byte) []byte {
	return append(GetTimestampSortableKey(timestamp), swapID...)
}
