package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/bep3/module/types"
	app "github.com/e-money/bep3/testapp"
	"github.com/stretchr/testify/suite"
)

type ParamsTestSuite struct {
	suite.Suite
	addr   sdk.AccAddress
	supply []types.SupplyLimit
}

func (suite *ParamsTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	suite.addr = addrs[0]
	supply1 := types.SupplyLimit{
		Limit:          sdk.NewInt(10000000000000),
		TimeLimited:    false,
		TimeBasedLimit: sdk.ZeroInt(),
		TimePeriod:     time.Hour,
	}
	supply2 := types.SupplyLimit{
		Limit:          sdk.NewInt(10000000000000),
		TimeLimited:    true,
		TimeBasedLimit: sdk.NewInt(100000000000),
		TimePeriod:     time.Hour * 24,
	}
	suite.supply = append(suite.supply, supply1, supply2)
	return
}

func (suite *ParamsTestSuite) TestParamValidation() {
	type args struct {
		assetParams types.AssetParams
	}

	testCases := []struct {
		name        string
		args        args
		expectPass  bool
		expectedErr string
	}{
		{
			name: "default",
			args: args{
				assetParams: types.AssetParams{},
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "valid single asset",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"bnb", 714, suite.supply[0], true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
					types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan)},
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "valid single asset time limited",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"bnb", 714, suite.supply[1], true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
					types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan)},
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "valid multi asset",
			args: args{
				assetParams: types.AssetParams{
					types.NewAssetParam(
						"bnb", 714, suite.supply[0], true,
						suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
						types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan),
					types.NewAssetParam(
						"btcb", 0, suite.supply[1], true,
						suite.addr, sdk.NewInt(1000), sdk.NewInt(10000000), sdk.NewInt(100000000000),
						types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan),
				},
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "invalid denom - empty",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"", 714, suite.supply[0], true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
					types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan)},
			},
			expectPass:  false,
			expectedErr: "denom invalid",
		},
		{
			name: "invalid denom - bad format",
			args: args{
				// note updated SDK denom regex mask  = `[a-zA-Z][a-zA-Z0-9/]{2,127}`
				assetParams: types.AssetParams{types.NewAssetParam(
					"1BNB", 714, suite.supply[0], true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
					types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan)},
			},
			expectPass:  false,
			expectedErr: "denom invalid",
		},
		{
			name: "min block lock equal max block lock",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"bnb", 714, suite.supply[0], true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
					243, 243)},
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "Swap time span < acceptable minimum value",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"bnb", 714, suite.supply[0], true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
					244, types.DefaultSwapTimeSpan-1)},
			},
			expectPass:  false,
			expectedErr: "asset bnb swap time span be within [60, 3 days in seconds] 59",
		},
		{
			name: "min swap not positive",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"bnb", 714, suite.supply[0], true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(0), sdk.NewInt(10000000000),
					types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan)},
			},
			expectPass:  false,
			expectedErr: "must have a positive minimum swap",
		},
		{
			name: "max swap not positive",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"bnb", 714, suite.supply[0], true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(10000), sdk.NewInt(0),
					types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan)},
			},
			expectPass:  false,
			expectedErr: "must have a positive maximum swap",
		},
		{
			name: "min swap greater max swap",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"bnb", 714, suite.supply[0], true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000000), sdk.NewInt(10000000000),
					types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan)},
			},
			expectPass:  false,
			expectedErr: "minimum swap amount > maximum swap amount",
		},
		{
			name: "negative coin id",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"bnb", -714, suite.supply[0], true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
					types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan)},
			},
			expectPass:  false,
			expectedErr: "coin id must be a non negative",
		},
		{
			name: "negative asset limit",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"bnb", 714,
					types.SupplyLimit{sdk.NewInt(-10000000000000), false, time.Hour, sdk.ZeroInt()}, true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
					types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan)},
			},
			expectPass:  false,
			expectedErr: "invalid (negative) supply limit",
		},
		{
			name: "negative asset time limit",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"bnb", 714,
					types.SupplyLimit{sdk.NewInt(10000000000000), false, time.Hour, sdk.NewInt(-10000000000000)}, true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
					types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan)},
			},
			expectPass:  false,
			expectedErr: "invalid (negative) supply time limit",
		},
		{
			name: "asset time limit greater than overall limit",
			args: args{
				assetParams: types.AssetParams{types.NewAssetParam(
					"bnb", 714,
					types.SupplyLimit{sdk.NewInt(10000000000000), true, time.Hour, sdk.NewInt(100000000000000)},
					true,
					suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
					types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan)},
			},
			expectPass:  false,
			expectedErr: "supply time limit > supply limit",
		},
		{
			name: "duplicate denom",
			args: args{
				assetParams: types.AssetParams{
					types.NewAssetParam(
						"bnb", 714, suite.supply[0], true,
						suite.addr, sdk.NewInt(1000), sdk.NewInt(100000000), sdk.NewInt(100000000000),
						types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan),
					types.NewAssetParam(
						"bnb", 0, suite.supply[0], true,
						suite.addr, sdk.NewInt(1000), sdk.NewInt(10000000), sdk.NewInt(100000000000),
						types.DefaultSwapBlockTimestamp, types.DefaultSwapTimeSpan),
				},
			},
			expectPass:  false,
			expectedErr: "duplicate denom",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			params := types.NewParams(tc.args.assetParams)
			err := params.Validate()
			if tc.expectPass {
				suite.Require().NoError(err, tc.name)
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().Contains(err.Error(), tc.expectedErr)
			}
		})
	}
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}
