package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	bep3 "github.com/e-money/bep3/module"
	"github.com/e-money/bep3/module/types"
	"github.com/stretchr/testify/require"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

func makeTestCodec() (cdc *codec.LegacyAmino) {
	encConfig := bep3.MakeProtoEncodingConfig()
	// encConfig := bep3.MakeAminoEncodingConfig()

	return encConfig.Amino
}

func TestDecodeBep3Store(t *testing.T) {
	cdc := makeTestCodec()
	prevBlockTime := time.Now().UTC()

	oneCoin := sdk.NewCoin("coin", sdk.OneInt())
	swap := types.NewAtomicSwap(sdk.Coins{oneCoin}, nil, 10, 100,
		nil, nil, "otherChainSender", "otherChainRec",
		200, types.Completed, true, types.Outgoing)
	supply := types.AssetSupply{
		IncomingSupply: oneCoin, OutgoingSupply: oneCoin, CurrentSupply: oneCoin,
		TimeLimitedCurrentSupply: oneCoin, TimeElapsed: 0,
	}
	bz := tmbytes.HexBytes([]byte{1, 2})

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.AtomicSwapKeyPrefix, Value: cdc.MustMarshalLengthPrefixed(swap)},
			{Key: types.AssetSupplyPrefix, Value: cdc.MustMarshalLengthPrefixed(supply)},
			{Key: types.AtomicSwapByBlockPrefix, Value: bz},
			{Key: types.AtomicSwapByBlockPrefix, Value: bz},
			{Key: types.PreviousBlockTimeKey, Value: cdc.MustMarshalLengthPrefixed(prevBlockTime)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"AtomicSwap", fmt.Sprintf("%v\n%v", swap, swap)},
		{"AssetSupply", fmt.Sprintf("%v\n%v", supply, supply)},
		{"AtomicSwapByBlock", fmt.Sprintf("%s\n%s", bz, bz)},
		{"AtomicSwapLongtermStorage", fmt.Sprintf("%s\n%s", bz, bz)},
		{"PreviousBlockTime", fmt.Sprintf("%s\n%s", prevBlockTime, prevBlockTime)},
		{"other", ""},
	}

	decodeStore := bep3.NewDecodeStore(cdc)

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { decodeStore(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, decodeStore(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
