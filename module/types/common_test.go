package types_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/bep3/module/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func ts(minOffset int) int64                { return tmtime.Now().Add(time.Duration(minOffset) * time.Minute).Unix() }

func atomicSwaps(count int) types.AtomicSwaps {
	var swaps types.AtomicSwaps
	for i := 0; i < count; i++ {
		swap := atomicSwap(i)
		swaps = append(swaps, swap)
	}
	return swaps
}

func atomicSwap(index int) types.AtomicSwap {
	timestamp := ts(index) // One minute apart
	expireTimestamp := time.Unix(timestamp, 0).
		Add(time.Duration(index*15)*time.Minute + 360).Unix() // 375 minutes apart
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber[:], timestamp)

	swap := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHash, expireTimestamp, timestamp,
		kavaAddrs[0], kavaAddrs[1], binanceAddrs[0].String(), binanceAddrs[1].String(), 1, types.Open,
		true, types.Incoming)

	return swap
}
