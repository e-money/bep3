package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/bep3/module/types"
	app "github.com/e-money/bep3/testapp"
	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite
	swaps    types.AtomicSwaps
	supplies types.AssetSupplies
}

func (suite *GenesisTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	coin := sdk.NewCoin("kava", sdk.OneInt())
	suite.swaps = atomicSwaps(10)

	supply := types.NewAssetSupply(coin, coin, coin, coin, 0)
	suite.supplies = types.AssetSupplies{AssetSupplies: []types.AssetSupply{supply}}
}

func (suite *GenesisTestSuite) TestValidate() {
	type args struct {
		swaps             types.AtomicSwaps
		supplies          types.AssetSupplies
		previousBlockTime time.Time
	}
	testCases := []struct {
		name       string
		args       args
		expectPass bool
	}{
		{
			"default",
			args{
				swaps:             types.AtomicSwaps{},
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			true,
		},
		{
			"with swaps",
			args{
				swaps:             suite.swaps,
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			true,
		},
		{
			"with supplies",
			args{
				swaps:             types.AtomicSwaps{},
				supplies:          suite.supplies,
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			true,
		},
		{
			"invalid supply",
			args{
				swaps:             types.AtomicSwaps{},
				supplies:          types.AssetSupplies{
					AssetSupplies: []types.AssetSupply{
						{
							IncomingSupply: sdk.Coin{
								Denom: "Invalid", Amount: sdk.ZeroInt(),
							},
						},
					},
				},
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			false,
		},
		{
			"duplicate swaps",
			args{
				swaps:             types.AtomicSwaps{suite.swaps[2], suite.swaps[2]},
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			false,
		},
		{
			"invalid swap",
			args{
				swaps:             types.AtomicSwaps{types.AtomicSwap{Amount: sdk.Coins{sdk.Coin{Denom: "Invalid Denom", Amount: sdk.NewInt(-1)}}}},
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			false,
		},
		{
			"blocktime not set",
			args{
				swaps: types.AtomicSwaps{},
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			var gs *types.GenesisState
			if tc.name == "default" {
				gs = types.DefaultGenesisState()
			} else {
				gs = types.NewGenesisState(types.DefaultParams(), tc.args.swaps, tc.args.supplies, tc.args.previousBlockTime)
			}

			err := gs.Validate()
			if tc.expectPass {
				suite.Require().NoError(err, tc.name)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
