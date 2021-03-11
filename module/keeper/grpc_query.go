package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/bep3/module/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) AssetSupply(c context.Context, req *types.QueryAssetSupplyRequest) (*types.QueryAssetSupplyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	assetSupply, found := k.GetAssetSupply(ctx, req.Denom)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrAssetSupplyNotFound, string(req.Denom))
	}

	return &types.QueryAssetSupplyResponse{
		Supply: assetSupply,
	}, nil
}

func (k Keeper) AssetSupplies(c context.Context, req *types.QueryAssetSuppliesRequest) (*types.QueryAssetSuppliesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	supplies := k.GetAllAssetSupplies(ctx)

	return &types.QueryAssetSuppliesResponse{Supplies: supplies}, nil
}

func (k Keeper) Swap(c context.Context, req *types.QuerySwapRequest) (*types.QuerySwapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	swap, found := k.GetAtomicSwap(ctx, req.SwapID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAtomicSwapNotFound, "%d", req.SwapID)
	}

	return &types.QuerySwapResponse{Swap: swap}, nil
}

func (k Keeper) Swaps(c context.Context, req *types.QuerySwapsRequest) (*types.QuerySwapsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	swaps := k.GetAllAtomicSwaps(ctx)
	if params := req.GetParams(); params != nil {
		swaps = filterAtomicSwaps(ctx, swaps, *params)
		if swaps == nil {
			swaps = types.AtomicSwaps{}
		}
	}

	augmentedSwaps := types.AugmentedAtomicSwaps{}

	for _, swap := range swaps {
		augmentedSwaps.AugmentedAtomicSwaps = append(augmentedSwaps.AugmentedAtomicSwaps, types.NewAugmentedAtomicSwap(swap))
	}

	return &types.QuerySwapsResponse{Swaps: augmentedSwaps}, nil
}

