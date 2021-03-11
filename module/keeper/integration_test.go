package keeper_test

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bep3 "github.com/e-money/bep3/module"
	"github.com/e-money/bep3/module/types"
	app "github.com/e-money/bep3/testapp"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	TestSenderOtherChain    = "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7"
	TestRecipientOtherChain = "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7"
	TestDeputy              = "app1xy7hrjy9r0algz9w3gzm8u6mrpq97kwtwewktj"
)

var (
	DenomMap  = map[int]string{0: "btc", 1: "eth", 2: "bnb", 3: "xrp", 4: "dai"}
	TestUser1 = sdk.AccAddress(crypto.AddressHash([]byte("KavaTestUser1")))
	TestUser2 = sdk.AccAddress(crypto.AddressHash([]byte("KavaTestUser2")))
)

func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func ts(minOffset int) int64                { return tmtime.Now().Add(time.Duration(minOffset) * time.Minute).Unix() }

func NewAuthGenStateFromAccs(accounts ...authtypes.GenesisAccount) app.GenesisState {
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), accounts)
	return app.GenesisState{authtypes.ModuleName: authtypes.ModuleCdc.MustMarshalJSON(authGenesis)}
}

func NewBep3GenState(deputyAddress sdk.AccAddress) json.RawMessage {
	bep3Genesis := types.GenesisState{
		Params: types.Params{
			AssetParams: types.AssetParams{
				types.AssetParam{
					Denom:  "bnb",
					CoinID: 714,
					SupplyLimit: types.SupplyLimit{
						Limit:          sdk.NewInt(350000000000000),
						TimeLimited:    false,
						TimeBasedLimit: sdk.ZeroInt(),
						TimePeriod:     int64(time.Hour),
					},
					Active:        true,
					DeputyAddress: deputyAddress.String(),
					FixedFee:      sdk.NewInt(1000),
					MinSwapAmount: sdk.OneInt(),
					MaxSwapAmount: sdk.NewInt(1000000000000),
					SwapTimeSpan:  bep3.DefaultSwapTimeSpan,
					SwapTimestamp: bep3.DefaultSwapBlockTimestamp,
				},
				types.AssetParam{
					Denom:  "inc",
					CoinID: 9999,
					SupplyLimit: types.SupplyLimit{
						Limit:          sdk.NewInt(100000000000000),
						TimeLimited:    true,
						TimeBasedLimit: sdk.NewInt(50000000000),
						TimePeriod:     int64(time.Hour),
					},
					Active:        false,
					DeputyAddress: deputyAddress.String(),
					FixedFee:      sdk.NewInt(1000),
					MinSwapAmount: sdk.OneInt(),
					MaxSwapAmount: sdk.NewInt(100000000000),
					SwapTimeSpan:  bep3.DefaultSwapTimeSpan,
					SwapTimestamp: bep3.DefaultSwapBlockTimestamp,
				},
			},
		},
		Supplies: types.AssetSupplies{
			AssetSupplies: []types.AssetSupply{
				{
					sdk.NewCoin("bnb", sdk.ZeroInt()),
					sdk.NewCoin("bnb", sdk.ZeroInt()),
					sdk.NewCoin("bnb", sdk.ZeroInt()),
					sdk.NewCoin("bnb", sdk.ZeroInt()),
					0,
				},
				{
					sdk.NewCoin("inc", sdk.ZeroInt()),
					sdk.NewCoin("inc", sdk.ZeroInt()),
					sdk.NewCoin("inc", sdk.ZeroInt()),
					sdk.NewCoin("inc", sdk.ZeroInt()),
					0,
				},
			},
		},
		PreviousBlockTime: types.DefaultPreviousBlockTime,
	}

	return types.ModuleCdc.MustMarshalJSON(&bep3Genesis)
}

func atomicSwaps(ctx sdk.Context, count int) types.AtomicSwaps {
	var swaps types.AtomicSwaps
	for i := 0; i < count; i++ {
		swap := atomicSwap(ctx, i)
		swaps = append(swaps, swap)
	}
	return swaps
}

func atomicSwap(ctx sdk.Context, index int) types.AtomicSwap {
	expireOffset := int64(200) // Default expire height + offet to match timestamp
	timestamp := ts(index)     // One minute apart
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber[:], timestamp)

	return types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHash,
		ctx.BlockTime().Unix()+expireOffset, timestamp, TestUser1, TestUser2,
		TestSenderOtherChain, TestRecipientOtherChain, 0, types.Open,
		true, types.Incoming)
}

func assetSupplies(count int) types.AssetSupplies {
	if count > 5 { // Max 5 asset supplies
		return types.AssetSupplies{}
	}

	var supplies types.AssetSupplies

	for i := 0; i < count; i++ {
		supply := assetSupply(DenomMap[i])
		supplies.AssetSupplies = append(supplies.AssetSupplies, supply)
	}
	return supplies
}

func assetSupply(denom string) types.AssetSupply {
	return types.NewAssetSupply(c(denom, 0), c(denom, 0), c(denom, 0), c(denom, 0), 0)
}
