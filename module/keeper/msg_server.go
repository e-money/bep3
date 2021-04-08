package keeper

import (
	"context"
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/bep3/module/types"
	"strconv"
)

var _ types.MsgServer = msgServer{}

type bep3Keeper interface {
	CreateAtomicSwapState(ctx sdk.Context, randomNumberHash []byte, timestamp, swapTimeSpanMin int64,
		sender, recipient sdk.AccAddress, senderOtherChain, recipientOtherChain string, amount sdk.Coins,
		crossChain bool) (*sdk.Result, error)
	ClaimAtomicSwapState(ctx sdk.Context, from sdk.AccAddress, swapID []byte, randomNumber []byte) (*sdk.Result, error)
	RefundAtomicSwapState(ctx sdk.Context, from sdk.AccAddress, swapID []byte) (*sdk.Result, error)
}

type msgServer struct {
	k bep3Keeper
}

func NewMsgServerImpl(keeper bep3Keeper) types.MsgServer {
	return &msgServer{k: keeper}
}

func (m msgServer)CreateAtomicSwap(goCtx context.Context, msg *types.MsgCreateAtomicSwap)(*types.MsgCreateAtomicSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	fromAcc, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender")
	}
	toAcc, err := sdk.AccAddressFromBech32(msg.To)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "to")
	}
	res, err := m.k.CreateAtomicSwapState(ctx, msg.RandomNumberHash, msg.Timestamp,
		msg.TimeSpanMin, fromAcc, toAcc, msg.SenderOtherChain, msg.RecipientOtherChain, msg.Amount, true)
	if err != nil {
		return nil, err
	}

	for _, e := range res.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}

	return &types.MsgCreateAtomicSwapResponse{
		RandomNumberHash: res.Log,
		SwapID:           hex.EncodeToString(res.Data),
	}, nil
}

func (m msgServer)ClaimAtomicSwap(goCtx context.Context, msg *types.MsgClaimAtomicSwap)(*types.MsgClaimAtomicSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	fromAcc, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender")
	}

	res, err := m.k.ClaimAtomicSwapState(ctx, fromAcc, msg.SwapID, msg.RandomNumber)
	if err != nil {
		return nil, err
	}

	for _, e := range res.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}

	timestamp, _ := strconv.Atoi(res.Log)

	return &types.MsgClaimAtomicSwapResponse{
		RandomNumberHash: hex.EncodeToString(res.Data),
		Timestamp:        int64(timestamp),
	}, nil
}

func (m msgServer)RefundAtomicSwap(goCtx context.Context, msg *types.MsgRefundAtomicSwap)(*types.MsgRefundAtomicSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	fromAcc, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender")
	}

	res, err := m.k.RefundAtomicSwapState(ctx, fromAcc, msg.SwapID)
	if err != nil {
		return nil, err
	}

	for _, e := range res.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}

	timestamp, _ := strconv.Atoi(res.Log)

	return &types.MsgRefundAtomicSwapResponse{
		RandomNumberHash: hex.EncodeToString(res.Data),
		Timestamp:        int64(timestamp),
	}, nil
}