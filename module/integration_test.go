package bep3_test

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bep3 "github.com/e-money/bep3/module"
	"github.com/e-money/bep3/module/types"
	app "github.com/e-money/bep3/testapp"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	TestSenderOtherChain    = "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7"
	TestRecipientOtherChain = "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7"
	TestDeputy              = "kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj"
	TestUser                = "kava1vry5lhegzlulehuutcr7nmdlmktw88awp0a39p"
)

var (
	StandardSupplyLimit = i(100000000000)
	DenomMap            = map[int]string{0: "bnb", 1: "inc"}
)

func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func ts(minOffset int) int64                { return tmtime.Now().Add(time.Duration(minOffset) * time.Minute).Unix() }

func NewBep3GenState(deputy sdk.AccAddress) json.RawMessage {
	bep3Genesis := baseGenState(deputy)
	return bep3.ModuleCdc.MustMarshalJSON(&bep3Genesis)
}

func NewBep3GenStateMulti(deputy sdk.AccAddress) app.GenesisState {
	bep3Genesis := baseGenState(deputy)
	return app.GenesisState{bep3.ModuleName: bep3.ModuleCdc.MustMarshalJSON(&bep3Genesis)}
}

func getBep3Coins() ([]string, sdk.Coins) {
	// bep3 genesis for supported coins
	bep3Denoms := []string{"bnb", "inc", "echf", "edkk", "eeur", "enok", "esek", "ungm"}
	coins := make(sdk.Coins, len(bep3Denoms))
	amount := sdk.NewInt(int64(bep3.MaxSupplyLimit))

	for idx, denom := range bep3Denoms {
		coins[idx] = sdk.NewCoin(denom, amount)
	}

	return bep3Denoms, coins
}

func baseGenState(deputy sdk.AccAddress) bep3.GenesisState {
	_ = &bep3.MsgCreateAtomicSwap{}
	bep3Genesis := bep3.GenesisState{
		Params: bep3.Params{
			AssetParams: bep3.AssetParams{
				bep3.AssetParam{
					Denom:  "bnb",
					CoinID: 714,
					SupplyLimit: bep3.SupplyLimit{
						Limit:          sdk.NewInt(350000000000000),
						TimeLimited:    false,
						TimeBasedLimit: sdk.ZeroInt(),
						TimePeriod:     int64(time.Hour),
					},
					Active:          true,
					DeputyAddress:   deputy.String(),
					FixedFee:        sdk.NewInt(1000),
					MinSwapAmount:   sdk.OneInt(),
					MaxSwapAmount:   sdk.NewInt(1000000000000),
					SwapTimeSpanMin: bep3.DefaultSwapTimeSpan,
					SwapTimestamp:   bep3.DefaultSwapBlockTimestamp,
				},
				bep3.AssetParam{
					Denom:  "inc",
					CoinID: 9999,
					SupplyLimit: bep3.SupplyLimit{
						Limit:          sdk.NewInt(100000000000),
						TimeLimited:    false,
						TimeBasedLimit: sdk.ZeroInt(),
						TimePeriod:     int64(time.Hour),
					},
					Active:          true,
					DeputyAddress:   deputy.String(),
					FixedFee:        sdk.NewInt(1000),
					MinSwapAmount:   sdk.OneInt(),
					MaxSwapAmount:   sdk.NewInt(1000000000000),
					SwapTimeSpanMin: bep3.DefaultSwapTimeSpan,
					SwapTimestamp:   bep3.DefaultSwapBlockTimestamp,
				},
			},
		},
		Supplies: bep3.AssetSupplies{
			AssetSupplies: []types.AssetSupply{
				{
					IncomingSupply:           sdk.NewCoin("bnb", sdk.ZeroInt()),
					OutgoingSupply:           sdk.NewCoin("bnb", sdk.ZeroInt()),
					CurrentSupply:            sdk.NewCoin("bnb", sdk.ZeroInt()),
					TimeLimitedCurrentSupply: sdk.NewCoin("bnb", sdk.ZeroInt()),
					TimeElapsed:              0,
				},
				{
					IncomingSupply:           sdk.NewCoin("inc", sdk.ZeroInt()),
					OutgoingSupply:           sdk.NewCoin("inc", sdk.ZeroInt()),
					CurrentSupply:            sdk.NewCoin("inc", sdk.ZeroInt()),
					TimeLimitedCurrentSupply: sdk.NewCoin("inc", sdk.ZeroInt()),
					TimeElapsed:              0,
				},
			},
		},
		PreviousBlockTime: bep3.DefaultPreviousBlockTime,
	}
	return bep3Genesis
}

func loadSwapAndSupply(addr sdk.AccAddress, index int) (bep3.AtomicSwap, bep3.AssetSupply) {
	coin := c(DenomMap[index], 50000)
	expireOffset := bep3.DefaultSwapBlockTimestamp + bep3.DefaultSwapTimeSpan // Default expiration seconds + offset to match timestamp
	timestamp := ts(index)                                                    // One minute apart
	randomNumber, _ := bep3.GenerateSecureRandomNumber()
	randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)
	swap := bep3.NewAtomicSwap(cs(coin), randomNumberHash,
		expireOffset, timestamp, addr, addr, TestSenderOtherChain,
		TestRecipientOtherChain, 1, bep3.Open, true, bep3.Incoming)

	supply := bep3.NewAssetSupply(coin, c(coin.Denom, 0),
		c(coin.Denom, 0), c(coin.Denom, 0), 0)

	return swap, supply
}
