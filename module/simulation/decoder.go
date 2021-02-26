package simulation

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/e-money/bep3/module/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

type CodecUnmarshaler interface {
	MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{})
}

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding type.
func NewDecodeStore(cdc CodecUnmarshaler) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.AtomicSwapKeyPrefix):
			var swapA, swapB types.AtomicSwap
			cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &swapA)
			cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &swapB)
			return fmt.Sprintf("%v\n%v", swapA, swapB)

		case bytes.Equal(kvA.Key[:1], types.AtomicSwapByBlockPrefix),
			bytes.Equal(kvA.Key[:1], types.AtomicSwapLongtermStoragePrefix):
			var bytesA tmbytes.HexBytes = kvA.Value
			var bytesB tmbytes.HexBytes = kvA.Value
			return fmt.Sprintf("%s\n%s", bytesA.String(), bytesB.String())
		case bytes.Equal(kvA.Key[:1], types.AssetSupplyPrefix):
			var supplyA, supplyB types.AssetSupply
			cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &supplyA)
			cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &supplyB)
			return fmt.Sprintf("%s\n%s", supplyA, supplyB)
		case bytes.Equal(kvA.Key[:1], types.PreviousBlockTimeKey):
			var timeA, timeB time.Time
			cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &timeA)
			cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &timeB)
			return fmt.Sprintf("%s\n%s", timeA, timeB)

		default:
			panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
		}
	}
}
