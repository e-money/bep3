package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/bep3/module/types"
)

const (
	sixtySeconds = 60
)

// createAtomicSwap creates a new atomic swap.
func (k Keeper) createAtomicSwap(ctx sdk.Context, randomNumberHash []byte, timestamp, swapTimeSpan int64,
	sender, recipient sdk.AccAddress, senderOtherChain, recipientOtherChain string, amount sdk.Coins,
	crossChain bool) (*sdk.Result, error) {
	// Confirm that this is not a duplicate swap
	swapID := types.CalculateSwapID(randomNumberHash, sender, senderOtherChain)
	hexSwapID := hex.EncodeToString(swapID)
	 fmt.Println("***************** Entered CreateAtomic Swap, swapID:", hexSwapID)
	 fmt.Printf("**** Swap from sender: %s,\nTo: %s\nsenderOtherChain: %s\n",
		sender.String(), recipient.String(), senderOtherChain)

	_, found := k.GetAtomicSwap(ctx, swapID)
	if found {
		return nil, sdkerrors.Wrap(types.ErrAtomicSwapAlreadyExists, hex.EncodeToString(swapID))
	}

	// Cannot send coins to a module account
	if k.Maccs[recipient.String()] {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is a module account", recipient)
	}

	if len(amount) != 1 {
		return nil, fmt.Errorf("amount must contain exactly one coin")
	}
	asset, err := k.GetAsset(ctx, amount[0].Denom)
	if err != nil {
		return nil, err
	}

	err = k.ValidateLiveAsset(ctx, amount[0])
	if err != nil {
		return nil, err
	}

	fmt.Println("***************** Validated Asset in createAtomicSwap(), amount:", amount[0].Amount.String())

	// Swap amount must be within the specified swap amount limits
	if amount[0].Amount.LT(asset.MinSwapAmount) || amount[0].Amount.GT(asset.MaxSwapAmount) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAmount, "amount %d outside range [%s, %s]", amount[0].Amount, asset.MinSwapAmount, asset.MaxSwapAmount)
	}

	// Unix timestamp must be in range [-15 mins, 30 mins] of the current time
	pastTimestampLimit := ctx.BlockTime().Add(time.Duration(-15) * time.Minute).Unix()
	fmt.Println("***************** Amount within limit, pastTimestampLimit:", time.Unix(pastTimestampLimit, 0).String())
	futureTimestampLimit := ctx.BlockTime().Add(time.Duration(30) * time.Minute).Unix()
	fmt.Println("***************** futureTimestampLimit:", time.Unix(futureTimestampLimit, 0).String())
	fmt.Println("***************** Sent timestamp", time.Unix(timestamp, 0).String())
	if timestamp < pastTimestampLimit || timestamp >= futureTimestampLimit {
		return nil, sdkerrors.Wrap(types.ErrInvalidTimestamp, fmt.Sprintf("block time: %s, timestamp: %s", ctx.BlockTime().String(), time.Unix(timestamp, 0).UTC().String()))
	}

	// fmt.Println("***************** Amount within limit")

	var direction types.SwapDirection
	if sender.String() == asset.DeputyAddress {
		if recipient.String() == asset.DeputyAddress {
			fmt.Println("***************** ErrInvalidSwapAccount: deputy cannot be both sender and receiver", asset.DeputyAddress)
			return nil, sdkerrors.Wrapf(types.ErrInvalidSwapAccount, "deputy cannot be both sender and receiver: %s", asset.DeputyAddress)
		}
		direction = types.Incoming
	} else {
		if recipient.String() != asset.DeputyAddress {
			fmt.Printf("***************** ErrInvalidSwapAccount: sender: %s, deputy %s must be recipient %s for outgoing swap\n",
				sender.String(), asset.DeputyAddress, recipient.String())
			return nil, sdkerrors.Wrapf(types.ErrInvalidSwapAccount,
				fmt.Sprintf("***************** ErrInvalidSwapAccount: sender: %s, deputy %s must be recipient %s for outgoing swap\n",
					sender.String(), asset.DeputyAddress, recipient.String()))
			//return nil, sdkerrors.Wrapf(types.ErrInvalidSwapAccount,
			//	"asset: %s deputy %s must be recipient for outgoing account: %s",asset.Denom, asset.DeputyAddress, recipient)
		}
		direction = types.Outgoing
	}

	fmt.Println("***************** SwapID:", hexSwapID, "is outgoing:", direction == types.Outgoing)

	switch direction {
	case types.Incoming:
		// If recipient's account doesn't exist, register it in state so that the address can send
		// a claim swap tx without needing to be registered in state by receiving a coin transfer.
		recipientAcc := k.accountKeeper.GetAccount(ctx, recipient)
		if recipientAcc == nil {
			newAcc := k.accountKeeper.NewAccountWithAddress(ctx, recipient)
			k.accountKeeper.SetAccount(ctx, newAcc)
		}
		// Incoming swaps have already had their fees collected by the deputy during the relay process.
		err = k.IncrementIncomingAssetSupply(ctx, amount[0])
	case types.Outgoing:

		// Outgoing swaps must have a seconds time span within [60, 3 days]
		if swapTimeSpan < sixtySeconds || swapTimeSpan > types.ThreeDaySeconds {
			return nil, sdkerrors.Wrapf(types.ErrInvalidTimeSpan,
				"seconds span %d outside range of 1 min...1 day[%d, %d]",
				swapTimeSpan, sixtySeconds, types.ThreeDaySeconds)
		}
		// Amount in outgoing swaps must be able to pay the deputy's fixed fee.
		fee := asset.FixedFee.Int64()
		if amount[0].Amount.LTE(asset.FixedFee.Add(asset.MinSwapAmount)) {
			fmt.Printf("amount %s - fee %d must be <= min swap amount %s\n",
				amount[0].Amount.String(), fee, asset.MinSwapAmount.String())
			return nil, sdkerrors.Wrap(types.ErrInsufficientAmount, amount[0].String())
		}
		err = k.IncrementOutgoingAssetSupply(ctx, amount[0])
		if err != nil {
			return nil, err
		}

		fmt.Println("***************** About to send to module account in CreateAtomic Swap")

		// Transfer coins to module - only needed for outgoing swaps
		err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
	default:
		err = fmt.Errorf("invalid swap direction: %s", direction.String())
	}
	if err != nil {
		return nil, err
	}

	// Store the details of the swap
	expireTime := ctx.BlockTime().Add(time.Duration(swapTimeSpan) * time.Second)
	atomicSwap := types.NewAtomicSwap(amount, randomNumberHash, expireTime.Unix(), timestamp, sender, recipient,
		senderOtherChain, recipientOtherChain, 0, types.Open, crossChain, direction)

	// fmt.Println("***************** About to store the swap in keeper within createAtomicSwap")

	// Insert the atomic swap under both keys
	k.SetAtomicSwap(ctx, atomicSwap)
	k.InsertIntoByTimestamp(ctx, atomicSwap)

	fmt.Println("***************** Success swap created id:", hex.EncodeToString(atomicSwap.GetSwapID()),
		", height:", ctx.BlockHeight(), ", amount:", atomicSwap.Amount.String(), ", timestamp:",
		time.Unix(int64(atomicSwap.Timestamp), 0).String(),
		", expiration:", time.Unix(int64(atomicSwap.ExpireTimestamp), 0).String())
	// Emit 'create_atomic_swap' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateAtomicSwap,
			sdk.NewAttribute(types.AttributeKeySender, atomicSwap.Sender),
			sdk.NewAttribute(types.AttributeKeyRecipient, atomicSwap.Recipient),
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, hex.EncodeToString(atomicSwap.GetSwapID())),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, hex.EncodeToString(atomicSwap.RandomNumberHash)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", atomicSwap.Timestamp)),
			sdk.NewAttribute(types.AttributeKeySenderOtherChain, atomicSwap.SenderOtherChain),
			sdk.NewAttribute(types.AttributeKeyExpireTimestamp, fmt.Sprintf("%d", atomicSwap.ExpireTimestamp)),
			sdk.NewAttribute(types.AttributeKeyAmount, atomicSwap.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyDirection, atomicSwap.Direction.String()),
		),
	)

	fmt.Println("***************** created",
		"to eth:", atomicSwap.RecipientOtherChain,
		sdk.NewAttribute(types.AttributeKeyClaimSender, atomicSwap.Sender).String(),
		sdk.NewAttribute(types.AttributeKeyRecipient, atomicSwap.Recipient).String(),
		sdk.NewAttribute(types.AttributeKeyAtomicSwapID, hex.EncodeToString(atomicSwap.GetSwapID())).String(),
		sdk.NewAttribute(types.AttributeKeyRandomNumberHash, hex.EncodeToString(atomicSwap.RandomNumberHash)).String())

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

// claimAtomicSwap validates a claim attempt, and if successful, sends the escrowed amount and closes the AtomicSwap.
func (k Keeper) claimAtomicSwap(ctx sdk.Context, from sdk.AccAddress, swapID []byte, randomNumber []byte) (*sdk.Result, error) {
	fmt.Println("***************** attempting to Claim Swap", hex.EncodeToString(swapID))
	atomicSwap, found := k.GetAtomicSwap(ctx, swapID)
	if !found {
		fmt.Println("***************** not found Swap", hex.EncodeToString(swapID))
		return nil, sdkerrors.Wrapf(types.ErrAtomicSwapNotFound, "%s", hex.EncodeToString(swapID))
	}

	// Only open atomic swaps can be claimed
	if atomicSwap.Status != types.Open {
		fmt.Println("***************** not open Swap", hex.EncodeToString(swapID))
		return nil, sdkerrors.Wrapf(types.ErrSwapNotClaimable, "status %s", atomicSwap.Status.String())
	}

	//  Calculate hashed secret using submitted number
	hashedSubmittedNumber := types.CalculateRandomHash(randomNumber, atomicSwap.Timestamp)

	swapSender, errBech := sdk.AccAddressFromBech32(atomicSwap.Sender)
	if errBech != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidSwapAccount, "ClaimSwap sender:%s, error:%s",
			atomicSwap.Sender, errBech)
	}

	hashedSecret := types.CalculateSwapID(hashedSubmittedNumber, swapSender, atomicSwap.SenderOtherChain)

	// Confirm that secret unlocks the atomic swap
	if !bytes.Equal(hashedSecret, atomicSwap.GetSwapID()) {
		fmt.Println("***************** incorrect random number:", hex.EncodeToString(randomNumber))
		return nil, sdkerrors.Wrapf(types.ErrInvalidClaimSecret, "the submitted random number is incorrect")
	}

	var err error
	switch atomicSwap.Direction {
	case types.Incoming:
		err = k.DecrementIncomingAssetSupply(ctx, atomicSwap.Amount[0])
		if err != nil {
			return nil, err
		}
		err = k.IncrementCurrentAssetSupply(ctx, atomicSwap.Amount[0])
		if err != nil {
			return nil, err
		}
		// incoming case - coins should be MINTED, then sent to user
		err = k.bankKeeper.MintCoins(ctx, types.ModuleName, atomicSwap.Amount)
		if err != nil {
			return nil, err
		}

		// Send intended recipient coins
		swapRecipient, errBech := sdk.AccAddressFromBech32(atomicSwap.Recipient)
		if errBech != nil {
			return nil, sdkerrors.Wrapf(types.ErrInvalidSwapAccount, "ClaimSwap sender:%s, error:%s", atomicSwap.Recipient, errBech)
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, swapRecipient, atomicSwap.Amount)
		if err != nil {
			return nil, err
		}
	case types.Outgoing:
		err = k.DecrementOutgoingAssetSupply(ctx, atomicSwap.Amount[0])
		if err != nil {
			return nil, err
		}
		err = k.DecrementCurrentAssetSupply(ctx, atomicSwap.Amount[0])
		if err != nil {
			return nil, err
		}
		// outgoing case  - coins should be burned
		err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, atomicSwap.Amount)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid swap direction: %s", atomicSwap.Direction.String())
	}

	fmt.Println("***************** Claim completed:", atomicSwap.GetSwapID())
	// Complete swap
	atomicSwap.Status = types.Completed
	atomicSwap.ClosedBlock = ctx.BlockHeight()
	k.SetAtomicSwap(ctx, atomicSwap)

	// Remove from byTimestamp key and transition to long term storage
	k.RemoveFromByTimestamp(ctx, atomicSwap)
	k.InsertIntoLongtermStorage(ctx, atomicSwap)

	// Emit 'claim_atomic_swap' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyClaimSender, from.String()),
			sdk.NewAttribute(types.AttributeKeyRecipient, atomicSwap.Recipient),
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, hex.EncodeToString(atomicSwap.GetSwapID())),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, hex.EncodeToString(atomicSwap.RandomNumberHash)),
			sdk.NewAttribute(types.AttributeKeyRandomNumber, hex.EncodeToString(randomNumber)),
		),
	)

	fmt.Println("***************** claimed, from:", from.String(),
		"to eth:", atomicSwap.RecipientOtherChain,
		sdk.NewAttribute(types.AttributeKeyRecipient, atomicSwap.Recipient).String(),
		sdk.NewAttribute(types.AttributeKeyAtomicSwapID, hex.EncodeToString(atomicSwap.GetSwapID())).String(),
		sdk.NewAttribute(types.AttributeKeyRandomNumberHash, hex.EncodeToString(atomicSwap.RandomNumberHash)).String(),
		sdk.NewAttribute(types.AttributeKeyRandomNumber, hex.EncodeToString(randomNumber)).String())

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

// refundAtomicSwap refunds an AtomicSwap, sending assets to the original sender and closing the AtomicSwap.
func (k Keeper) refundAtomicSwap(ctx sdk.Context, from sdk.AccAddress, swapID []byte) (*sdk.Result, error) {
	fmt.Println(
		"***************** attempting to refund Swap",
		hex.EncodeToString(swapID))
	atomicSwap, found := k.GetAtomicSwap(ctx, swapID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAtomicSwapNotFound, "%s", swapID)
	}
	// Only expired swaps may be refunded
	if atomicSwap.Status != types.Expired {
		return nil, sdkerrors.Wrapf(
			types.ErrSwapNotRefundable, "status %s", atomicSwap.Status.String(),
		)
	}

	var err error
	switch atomicSwap.Direction {
	case types.Incoming:
		err = k.DecrementIncomingAssetSupply(ctx, atomicSwap.Amount[0])
	case types.Outgoing:
		err = k.DecrementOutgoingAssetSupply(ctx, atomicSwap.Amount[0])
		if err != nil {
			return nil, err
		}

		// Refund coins to original swap sender for outgoing swaps
		swapSender, errBech := sdk.AccAddressFromBech32(atomicSwap.Sender)
		if errBech != nil {
			return nil, sdkerrors.Wrapf(
				types.ErrInvalidSwapAccount, "RefundSwap sender:%s, error:%s",
				atomicSwap.Sender, errBech,
			)
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx, types.ModuleName, swapSender, atomicSwap.Amount,
		)
	default:
		err = fmt.Errorf(
			"invalid swap direction: %s", atomicSwap.Direction.String(),
		)
	}

	if err != nil {
		return nil, err
	}

	// Complete swap
	atomicSwap.Status = types.Completed
	atomicSwap.ClosedBlock = ctx.BlockHeight()
	k.SetAtomicSwap(ctx, atomicSwap)

	// Transition to longterm storage
	k.InsertIntoLongtermStorage(ctx, atomicSwap)

	// Emit 'refund_atomic_swap' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRefundAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyRefundSender, from.String()),
			sdk.NewAttribute(types.AttributeKeySender, atomicSwap.Sender),
			sdk.NewAttribute(
				types.AttributeKeyAtomicSwapID,
				hex.EncodeToString(atomicSwap.GetSwapID()),
			),
			sdk.NewAttribute(
				types.AttributeKeyRandomNumberHash,
				hex.EncodeToString(atomicSwap.RandomNumberHash),
			),
		),
	)

	fmt.Println(
		"***************** refunded, from:", from.String(),
		"to eth:", atomicSwap.RecipientOtherChain,
		sdk.NewAttribute(
			types.AttributeKeyRecipient, atomicSwap.Recipient,
		).String(),
		sdk.NewAttribute(
			types.AttributeKeyAtomicSwapID,
			hex.EncodeToString(atomicSwap.GetSwapID()),
		).String(),
		sdk.NewAttribute(
			types.AttributeKeyRandomNumberHash,
			hex.EncodeToString(atomicSwap.RandomNumberHash),
		).String(),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

// UpdateExpiredAtomicSwaps finds all AtomicSwaps that are past (or at) their ending times and expires them.
func (k Keeper) UpdateExpiredAtomicSwaps(ctx sdk.Context) {
	var expiredSwapIDs []string
	k.IterateAtomicSwapsByBlock(ctx, ctx.BlockTime().Unix(), func(id []byte) bool {
		atomicSwap, found := k.GetAtomicSwap(ctx, id)
		if !found {
			// NOTE: shouldn't happen. Continue to next item.
			return false
		}
		// Expire the uncompleted swap and update both indexes
		atomicSwap.Status = types.Expired
		// Note: claimed swaps have already been removed from byBlock index.
		k.RemoveFromByTimestamp(ctx, atomicSwap)
		k.SetAtomicSwap(ctx, atomicSwap)
		expiredSwapIDs = append(expiredSwapIDs, hex.EncodeToString(atomicSwap.GetSwapID()))
		return false
	})

	// Emit 'swaps_expired' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapsExpired,
			sdk.NewAttribute(types.AttributeKeyAtomicSwapIDs, fmt.Sprintf("%s", expiredSwapIDs)),
			sdk.NewAttribute(types.AttributeExpirationBlock, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

// DeleteClosedAtomicSwapsFromLongtermStorage removes swaps one week after completion.
func (k Keeper) DeleteClosedAtomicSwapsFromLongtermStorage(ctx sdk.Context) {
	k.IterateAtomicSwapsLongtermStorage(ctx, uint64(ctx.BlockHeight()), func(id []byte) bool {
		swap, found := k.GetAtomicSwap(ctx, id)
		if !found {
			// NOTE: shouldn't happen. Continue to next item.
			return false
		}
		k.RemoveAtomicSwap(ctx, swap.GetSwapID())
		k.RemoveFromLongtermStorage(ctx, swap)
		return false
	})
}
