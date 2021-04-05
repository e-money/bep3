package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	ThreeDayMinutes = 60 * 24 * 3

	// Todo set this to a meaningful value
	DeputyFee = 5000
)

// Parameter keys
var (
	KeyAssetParams = []byte("AssetParams")

	DefaultMinAmount          sdk.Int = sdk.ZeroInt()
	DefaultMaxAmount          sdk.Int = sdk.NewInt(1000000000000) // 10,000 BNB
	DefaultPreviousBlockTime          = tmtime.Canonical(time.Unix(0, 0))
	DefaultSwapBlockTimestamp int64   = 10 // At 10th second.
	DefaultSwapTimeSpan       int64   = 180 // 5 minutes
)

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	AssetParams: %s`,
		p.AssetParams)
}

// NewParams returns a new params object
func NewParams(ap AssetParams,
) Params {
	return Params{
		AssetParams: ap,
	}
}

// DefaultParams returns default params for bep3 module
func DefaultParams() Params {
	return NewParams(AssetParams{})
}

// NewAssetParam returns a new AssetParam
func NewAssetParam(denom string, coinID int64, limit SupplyLimit, active bool,
	deputyAddr sdk.AccAddress, fixedFee sdk.Int, minSwapAmount sdk.Int, maxSwapAmount sdk.Int,
	swapTimestamp int64, swapTimeSpanMin int64) AssetParam {

	fmt.Printf("*** NewAssetParam Fee:%s%s\n", denom, fixedFee.String())
	if strings.Contains(denom, "ngm") || strings.Contains(denom, "NGM") {
		panic("*** NewAssetParam stop -> ***")
	}

	return AssetParam{
		Denom:           denom,
		CoinID:          coinID,
		SupplyLimit:     limit,
		Active:          active,
		DeputyAddress:   deputyAddr.String(),
		FixedFee:        fixedFee,
		MinSwapAmount:   minSwapAmount,
		MaxSwapAmount:   maxSwapAmount,
		SwapTimestamp:   swapTimestamp,
		SwapTimeSpanMin: swapTimeSpanMin,
	}
}

// String implements fmt.Stringer
func (ap AssetParam) String() string {
	return fmt.Sprintf(`Asset:
	Denom: %s
	Coin ID: %d
	Limit: %s
	Active: %t
	Deputy Address: %s
	Fixed Fee: %s
	Min Swap Amount: %s
	Max Swap Amount: %s
	Swap Time in Seconds: %d
	Time Span in Minutes: %d`,
		ap.Denom, ap.CoinID, ap.SupplyLimit, ap.Active, ap.DeputyAddress, ap.FixedFee,
		ap.MinSwapAmount, ap.MaxSwapAmount, ap.SwapTimestamp, ap.SwapTimeSpanMin)
}

// AssetParams array of AssetParam
type AssetParams []AssetParam

// String implements fmt.Stringer
func (aps AssetParams) String() string {
	out := "Asset Params\n"
	for _, ap := range aps {
		out += fmt.Sprintf("%s\n", ap)
	}
	return out
}

// String implements fmt.Stringer
func (sl SupplyLimit) String() string {
	return fmt.Sprintf(`%s
	%t
	%s
	%s
	`, sl.Limit, sl.TimeLimited, time.Duration(sl.TimePeriod), sl.TimeBasedLimit)
}

// Equals returns true if two supply limits are equal
func (sl SupplyLimit) Equals(sl2 SupplyLimit) bool {
	return sl.Limit.Equal(sl2.Limit) && sl.TimeLimited == sl2.TimeLimited && sl.TimePeriod == sl2.TimePeriod && sl.TimeBasedLimit.Equal(sl2.TimeBasedLimit)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of bep3 module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAssetParams, &p.AssetParams, validateAssetParams),
	}
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	return validateAssetParams(p.AssetParams)
}

func validateAssetParams(i interface{}) error {
	assetParams, ok := i.([]AssetParam)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	coinDenoms := make(map[string]bool)
	for _, asset := range assetParams {
		if err := sdk.ValidateDenom(asset.Denom); err != nil {
			return fmt.Errorf(fmt.Sprintf("asset denom invalid: %s", asset.Denom))
		}

		if asset.CoinID < 0 {
			return fmt.Errorf(fmt.Sprintf("asset %s coin id must be a non negative integer", asset.Denom))
		}

		if asset.SupplyLimit.Limit.IsNegative() {
			return fmt.Errorf(fmt.Sprintf("asset %s has invalid (negative) supply limit: %s", asset.Denom, asset.SupplyLimit.Limit))
		}

		if asset.SupplyLimit.TimeBasedLimit.IsNegative() {
			return fmt.Errorf(fmt.Sprintf("asset %s has invalid (negative) supply time limit: %s", asset.Denom, asset.SupplyLimit.TimeBasedLimit))
		}

		if asset.SupplyLimit.TimeBasedLimit.GT(asset.SupplyLimit.Limit) {
			return fmt.Errorf(fmt.Sprintf("asset %s cannot have supply time limit > supply limit: %s>%s", asset.Denom, asset.SupplyLimit.TimeBasedLimit, asset.SupplyLimit.Limit))
		}

		_, found := coinDenoms[asset.Denom]
		if found {
			return fmt.Errorf(fmt.Sprintf("asset %s cannot have duplicate denom", asset.Denom))
		}

		coinDenoms[asset.Denom] = true

		if len(asset.DeputyAddress) == 0 {
			return fmt.Errorf("deputy address cannot be empty for %s", asset.Denom)
		}

		depAddr, err := sdk.AccAddressFromBech32(asset.DeputyAddress)
		if err != nil {
			return err
		}
		if len(depAddr.Bytes()) != sdk.AddrLen {
			return fmt.Errorf("%s deputy address invalid bytes length got %d, want %d", asset.Denom, len([]byte(asset.DeputyAddress)), sdk.AddrLen)
		}

		if asset.FixedFee.IsNegative() {
			return fmt.Errorf("asset %s cannot have a negative fixed fee %s", asset.Denom, asset.FixedFee)
		}

		if asset.SwapTimeSpanMin < 1 || asset.SwapTimeSpanMin > ThreeDayMinutes {
			return fmt.Errorf("asset %s swap time span be within [1, 3 days in minutes(4320)] %d", asset.Denom, asset.SwapTimeSpanMin)
		}

		if !asset.MinSwapAmount.IsPositive() {
			return fmt.Errorf(fmt.Sprintf("asset %s must have a positive minimum swap amount, got %s", asset.Denom, asset.MinSwapAmount))
		}

		if !asset.MaxSwapAmount.IsPositive() {
			return fmt.Errorf(fmt.Sprintf("asset %s must have a positive maximum swap amount, got %s", asset.Denom, asset.MaxSwapAmount))
		}

		if asset.MinSwapAmount.GT(asset.MaxSwapAmount) {
			return fmt.Errorf("asset %s has minimum swap amount > maximum swap amount %s > %s", asset.Denom, asset.MinSwapAmount, asset.MaxSwapAmount)
		}
	}

	return nil
}
