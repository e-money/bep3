package bep3_test

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bep3 "github.com/e-money/bep3/module"
	bep3types "github.com/e-money/bep3/module/types"
	app "github.com/e-money/bep3/testapp"
	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite

	appModule     bep3.AppModule
	ctx           sdk.Context
	keeper        bep3.Keeper
	addrs         []sdk.AccAddress
	jsonMarshaler codec.JSONMarshaler
}

func (suite *GenesisTestSuite) SetupTest() {
	ctx, jsonMarshaller, bep3Keeper, _, _, appModule := app.CreateTestComponents(suite.T())

	suite.ctx = ctx
	suite.keeper = bep3Keeper
	suite.appModule = appModule
	suite.jsonMarshaler = jsonMarshaller

	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	suite.addrs = addrs
}

func (suite *GenesisTestSuite) TestGenesisState() {
	type GenState func() app.GenesisState

	testCases := []struct {
		name       string
		genState   GenState
		expectPass bool
	}{
		{
			name: "default",
			genState: func() app.GenesisState {
				return NewBep3GenStateMulti(suite.addrs[0])
			},
			expectPass: true,
		},
		{
			name: "import atomic swaps and asset supplies",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				var swaps bep3.AtomicSwaps
				var supplies bep3.AssetSupplies
				for i := 0; i < 2; i++ {
					swap, supply := loadSwapAndSupply(addrs[i], i)
					swaps = append(swaps, swap)
					supplies.AssetSupplies = append(supplies.AssetSupplies, supply)
				}
				gs.AtomicSwaps = swaps
				gs.Supplies = supplies
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: true,
		},
		{
			name: "0 deputy fees",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.AssetParams[0].FixedFee = sdk.ZeroInt()
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: true,
		},
		{
			name: "incoming supply doesn't match amount in incoming atomic swaps",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				_, addrs := app.GeneratePrivKeyAddressPairs(1)
				swap, _ := loadSwapAndSupply(addrs[0], 1)
				gs.AtomicSwaps = bep3.AtomicSwaps{swap}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "current supply above limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				assetParam, _ := suite.keeper.GetAsset(suite.ctx, "bnb")
				gs.Supplies = bep3types.AssetSupplies{
					AssetSupplies: []bep3types.AssetSupply{
						{
							IncomingSupply: c("bnb", 0),
							OutgoingSupply: c("bnb", 0),
							CurrentSupply:  c("bnb", assetParam.SupplyLimit.Limit.Add(i(1)).Int64()),
						},
					},
				}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "incoming supply above limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				// Set up overlimit amount
				assetParam, _ := suite.keeper.GetAsset(suite.ctx, "bnb")
				overLimitAmount := assetParam.SupplyLimit.Limit.Add(i(1))

				// Set up an atomic swap with amount equal to the currently asset supply
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := bep3.GenerateSecureRandomNumber()
				randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)
				swap := bep3.NewAtomicSwap(cs(c("bnb", overLimitAmount.Int64())), randomNumberHash,
					bep3.DefaultSwapTimeSpan, timestamp, suite.addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, bep3.Open, true, bep3.Incoming)
				gs.AtomicSwaps = bep3.AtomicSwaps{swap}

				// Set up asset supply with overlimit current supply
				gs.Supplies = bep3types.AssetSupplies{
					AssetSupplies: []bep3types.AssetSupply{
						{
							IncomingSupply: c("bnb", assetParam.SupplyLimit.Limit.Add(i(1)).Int64()),
							OutgoingSupply: c("bnb", 0),
							CurrentSupply:  c("bnb", 0),
						},
					},
				}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "emtest",
			genState: func() app.GenesisState {
				gs := bep3types.DefaultGenesisState()
				bep3Denoms, bep3Coins := getBep3Coins()
				gs.Params.AssetParams = make([]bep3types.AssetParam, len(bep3Denoms))
				gs.Supplies.AssetSupplies = make([]bep3types.AssetSupply, len(bep3Denoms))

				// Deterministic randomizer
				r := rand.New(rand.NewSource(1))
				limit := sdk.NewInt(int64(bep3.MaxSupplyLimit))
				for idx, denom := range bep3Denoms {
					bep3Coins[idx] = sdk.NewCoin(denom, limit)

					gs.Supplies.AssetSupplies[idx] = bep3types.AssetSupply{
						IncomingSupply:           sdk.NewCoin(denom, sdk.ZeroInt()),
						OutgoingSupply:           sdk.NewCoin(denom, sdk.ZeroInt()),
						CurrentSupply:            sdk.NewCoin(denom, limit),
						TimeLimitedCurrentSupply: sdk.NewCoin(denom, sdk.ZeroInt()),
						TimeElapsed:              0,
					}
					gs.Params.AssetParams[idx] =
						bep3types.AssetParam{
							Denom:  denom,
							CoinID: int64(idx) + 1,
							SupplyLimit: bep3types.SupplyLimit{
								Limit:          limit,
								TimeLimited:    false,
								TimePeriod:     int64(time.Hour * 24),
								TimeBasedLimit: sdk.ZeroInt(),
							},
							Active:          true,
							DeputyAddress:   suite.addrs[0].String(),
							FixedFee:        bep3.GenRandFixedFee(r),
							MinSwapAmount:   sdk.OneInt(),
							MaxSwapAmount:   limit,
							SwapTimestamp:   time.Now().Unix(),
							SwapTimeSpanMin: 60 * 24 * 3, // 3 days
						}
				}

				// test 2-way marshalling
				gsBytes := bep3.ModuleCdc.MustMarshalJSON(gs)
				var gsCp bep3.GenesisState
				bep3.ModuleCdc.MustUnmarshalJSON(gsBytes, &gsCp)
				if !gs.Equal(gsCp) {
					panic("unwound genesis not equal")
				}
				return app.GenesisState{"bep3": gsBytes}
			},
			expectPass: true,
		},
		{
			name: "incoming supply + current supply above limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				// Set up overlimit amount
				assetParam, _ := suite.keeper.GetAsset(suite.ctx, "bnb")
				halfLimit := assetParam.SupplyLimit.Limit.Int64() / 2
				overHalfLimit := halfLimit + 1

				// Set up an atomic swap with amount equal to the currently asset supply
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := bep3.GenerateSecureRandomNumber()
				randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)
				swap := bep3.NewAtomicSwap(cs(c("bnb", halfLimit)), randomNumberHash,
					360, timestamp, suite.addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, bep3.Open, true, bep3.Incoming)
				gs.AtomicSwaps = bep3.AtomicSwaps{swap}

				// Set up asset supply with overlimit current supply
				gs.Supplies = bep3types.AssetSupplies{
					AssetSupplies: []bep3types.AssetSupply{
						{
							IncomingSupply: c("bnb", halfLimit),
							OutgoingSupply: c("bnb", 0),
							CurrentSupply:  c("bnb", overHalfLimit),
						},
					},
				}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "outgoing supply above limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				// Set up overlimit amount
				assetParam, _ := suite.keeper.GetAsset(suite.ctx, "bnb")
				overLimitAmount := assetParam.SupplyLimit.Limit.Add(i(1))

				// Set up an atomic swap with amount equal to the currently asset supply
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := bep3.GenerateSecureRandomNumber()
				randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)
				swap := bep3.NewAtomicSwap(cs(c("bnb", overLimitAmount.Int64())), randomNumberHash,
					bep3.DefaultSwapTimeSpan, timestamp, addrs[1], suite.addrs[0], TestSenderOtherChain,
					TestRecipientOtherChain, 0, bep3.Open, true, bep3.Outgoing)
				gs.AtomicSwaps = bep3.AtomicSwaps{swap}

				// Set up asset supply with overlimit current supply
				gs.Supplies = bep3.AssetSupplies{
					AssetSupplies: []bep3types.AssetSupply{
						{
							IncomingSupply: c("bnb", 0),
							OutgoingSupply: c("bnb", 0),
							CurrentSupply:  c("bnb", assetParam.SupplyLimit.Limit.Add(i(1)).Int64()),
						},
					},
				}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "asset supply denom is not a supported asset",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Supplies = bep3types.AssetSupplies{
					AssetSupplies: []bep3types.AssetSupply{
						{
							IncomingSupply: c("fake", 0),
							OutgoingSupply: c("fake", 0),
							CurrentSupply:  c("fake", 0),
						},
					},
				}

				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "atomic swap asset type is unsupported",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := bep3.GenerateSecureRandomNumber()
				randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)
				swap := bep3.NewAtomicSwap(cs(c("fake", 500000)), randomNumberHash,
					360, timestamp, suite.addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, bep3.Open, true, bep3.Incoming)

				gs.AtomicSwaps = bep3.AtomicSwaps{swap}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "atomic swap status is invalid",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := bep3.GenerateSecureRandomNumber()
				randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)
				swap := bep3.NewAtomicSwap(cs(c("bnb", 5000)), randomNumberHash,
					360, timestamp, suite.addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, bep3.NULL, true, bep3.Incoming)

				gs.AtomicSwaps = bep3.AtomicSwaps{swap}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "time lock cannot be < 1 minute",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.AssetParams[0].SwapTimeSpanMin = 59
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "time lock cannot be > 3 days",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.AssetParams[0].SwapTimeSpanMin = (60 * 24 * 3) + 1 // 1 week + 1 minute
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "empty supported asset denom",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.AssetParams[0].Denom = ""
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "negative supported asset limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.AssetParams[0].SupplyLimit.Limit = i(-100)
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
		{
			name: "duplicate supported asset denom",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.AssetParams[1].Denom = "bnb"
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(&gs)}
			},
			expectPass: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			var genState json.RawMessage
			suite.NotPanics(func() {
				genState = tc.genState()[bep3.ModuleName]
			}, "Error in test case setup: %v", tc.name)

			if tc.expectPass {
				suite.NotPanics(func() {
					suite.appModule.InitGenesis(suite.ctx, suite.jsonMarshaler, genState)
				}, tc.name)
			} else {
				suite.Panics(func() {
					suite.appModule.InitGenesis(suite.ctx, suite.jsonMarshaler, genState)
				}, tc.name)
			}
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
