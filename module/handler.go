package bep3

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/bep3/module/types"
)

// NewHandler creates an sdk.Handler for all the bep3 type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *MsgCreateAtomicSwap:
			return handleMsgCreateAtomicSwap(ctx, k, msg)
		case *MsgClaimAtomicSwap:
			return handleMsgClaimAtomicSwap(ctx, k, msg)
		case *MsgRefundAtomicSwap:
			return handleMsgRefundAtomicSwap(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

// handleMsgCreateAtomicSwap handles requests to create a new AtomicSwap
func handleMsgCreateAtomicSwap(ctx sdk.Context, k Keeper, msg *MsgCreateAtomicSwap) (*sdk.Result, error) {
	fmt.Println("***************** Entered handleMsgCreateAtomic Swap")
	from, errBech := sdk.AccAddressFromBech32(msg.From)
	if errBech != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidSwapAccount, "handleMsgCreateAtomicSwap from:%s, error:%s",
			msg.From, errBech)
	}

	to, errBech := sdk.AccAddressFromBech32(msg.To)
	if errBech != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidSwapAccount, "handleMsgCreateAtomicSwap to:%s, error:%s",
			msg.From, errBech)
	}

	err := k.createAtomicSwap(ctx, msg.RandomNumberHash, msg.Timestamp, msg.TimeSpan,
		from, to, msg.SenderOtherChain, msg.RecipientOtherChain, msg.Amount, true)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

// handleMsgClaimAtomicSwap handles requests to claim funds in an active AtomicSwap
func handleMsgClaimAtomicSwap(ctx sdk.Context, k Keeper, msg *MsgClaimAtomicSwap) (*sdk.Result, error) {
	fmt.Println("***************** Entered handleMsgClaimAtomicSwap")
	from, errBech := sdk.AccAddressFromBech32(msg.From)
	if errBech != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidSwapAccount, "handleMsgClaimAtomicSwap from:%s, error:%s",
			msg.From, errBech)
	}

	err := k.claimAtomicSwap(ctx, from, msg.SwapID, msg.RandomNumber)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

// handleMsgRefundAtomicSwap handles requests to refund an active AtomicSwap
func handleMsgRefundAtomicSwap(ctx sdk.Context, k Keeper, msg *MsgRefundAtomicSwap) (*sdk.Result, error) {
	fmt.Println("***************** Entered handleMsgRefundAtomicSwap")
	from, errBech := sdk.AccAddressFromBech32(msg.From)
	if errBech != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidSwapAccount, "handleMsgRefundAtomicSwap from:%s, error:%s",
			msg.From, errBech)
	}

	err := k.refundAtomicSwap(ctx, from, msg.SwapID)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}
