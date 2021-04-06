package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/bep3/module/types"
)

// GetParams returns the total set of bep3 parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSubspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the bep3 parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// ------------------------------------------
//				Asset
// ------------------------------------------

// GetAsset returns the asset param associated with the input denom
func (k Keeper) GetAsset(ctx sdk.Context, denom string) (types.AssetParam, error) {
	params := k.GetParams(ctx)
	for _, asset := range params.AssetParams {
		if denom == asset.Denom {
			return asset, nil
		}
	}
	return types.AssetParam{}, sdkerrors.Wrap(types.ErrAssetNotSupported, denom)
}

// SetAsset sets an asset in the params
func (k Keeper) SetAsset(ctx sdk.Context, asset types.AssetParam) {
	params := k.GetParams(ctx)
	for i := range params.AssetParams {
		if params.AssetParams[i].Denom == asset.Denom {
			params.AssetParams[i] = asset
		}
	}
	k.SetParams(ctx, params)
}

// GetAssets returns a list containing all supported assets
func (k Keeper) GetAssets(ctx sdk.Context) (types.AssetParams, bool) {
	params := k.GetParams(ctx)
	return params.AssetParams, len(params.AssetParams) > 0
}

// ------------------------------------------
//				Asset-specific getters
// ------------------------------------------

// GetDeputyAddress returns the deputy address for the input denom
func (k Keeper) GetDeputyAddress(ctx sdk.Context, denom string) (sdk.AccAddress, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return sdk.AccAddress{}, err
	}
	depAddr, err := sdk.AccAddressFromBech32(asset.DeputyAddress)
	if err != nil {
		return sdk.AccAddress{}, err
	}
	return depAddr, nil
}

// GetFixedFee returns the fixed fee for incoming swaps
func (k Keeper) GetFixedFee(ctx sdk.Context, denom string) (sdk.Int, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return sdk.Int{}, err
	}
	return asset.FixedFee, nil
}

// GetMinSwapAmount returns the minimum swap amount
func (k Keeper) GetMinSwapAmount(ctx sdk.Context, denom string) (sdk.Int, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return sdk.Int{}, err
	}
	return asset.MinSwapAmount, nil
}

// GetMaxSwapAmount returns the maximum swap amount
func (k Keeper) GetMaxSwapAmount(ctx sdk.Context, denom string) (sdk.Int, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return sdk.Int{}, err
	}
	return asset.MaxSwapAmount, nil
}

// GetSwapTime returns the swap creation block Unix seconds timestamp
func (k Keeper) GetSwapTime(ctx sdk.Context, denom string) (int64, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return int64(0), err
	}
	return asset.SwapTimestamp, nil
}

// GetTimeSpanMin returns the swap minutes allowance
func (k Keeper) GetTimeSpanMin(ctx sdk.Context, denom string) (int64, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return int64(0), err
	}
	return asset.SwapTimeSpanMin, nil
}

// GetAssetByCoinID returns an asset by its denom
func (k Keeper) GetAssetByCoinID(ctx sdk.Context, coinID int64) (types.AssetParam, bool) {
	params := k.GetParams(ctx)
	for _, asset := range params.AssetParams {
		if asset.CoinID == coinID {
			return asset, true
		}
	}
	return types.AssetParam{}, false
}

// ValidateLiveAsset checks if an asset is both supported and active
func (k Keeper) ValidateLiveAsset(ctx sdk.Context, coin sdk.Coin) error {
	asset, err := k.GetAsset(ctx, coin.Denom)
	if err != nil {
		return err
	}
	if !asset.Active {
		return sdkerrors.Wrap(types.ErrAssetNotActive, asset.Denom)
	}
	return nil
}

// GetSupplyLimit returns the supply limit for the input denom
func (k Keeper) GetSupplyLimit(ctx sdk.Context, denom string) (types.SupplyLimit, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return types.SupplyLimit{}, err
	}
	return asset.SupplyLimit, nil
}
