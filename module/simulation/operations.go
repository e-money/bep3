package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/e-money/bep3/module/keeper"
	"github.com/e-money/bep3/module/types"
)

var (
	noOpMsg      = simtypes.NoOpMsg(types.ModuleName, "NoOpMsg", "")
	randomNumber = []byte{114, 21, 74, 180, 81, 92, 21, 91, 173, 164, 143, 111, 120, 58, 241, 58, 40, 22, 59, 133, 102, 233, 55, 149, 12, 199, 231, 63, 122, 23, 88, 9}
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateAtomicSwap = "op_weight_msg_create_atomic_swap"
)

func defaultWeightMsgCreateAtomicSwap(r *rand.Rand) int {
	return simtypes.RandIntBetween(r, 20, 100)
}

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightCreateAtomicSwap int

	// GenDepositParamsDepositPeriod randomized DepositParamsDepositPeriod
	appParams.GetOrGenerate(cdc, OpWeightMsgCreateAtomicSwap, &weightCreateAtomicSwap, nil,
		func(r *rand.Rand) {
			weightCreateAtomicSwap = defaultWeightMsgCreateAtomicSwap(r)
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightCreateAtomicSwap,
			SimulateMsgCreateAtomicSwap(ak, bk, k),
		),
	}
}

// SimulateMsgCreateAtomicSwap generates a MsgCreateAtomicSwap with random values
func SimulateMsgCreateAtomicSwap(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		const msgType = "MsgCreateAtomicSwap"

		// Get asset supplies and shuffle them
		assets, found := k.GetAssets(ctx)
		if !found {
			return noOpMsg, nil, nil
		}
		r.Shuffle(len(assets), func(i, j int) {
			assets[i], assets[j] = assets[j], assets[i]
		})
		senderOutgoing, selectedAsset, found := findValidAccountAssetPair(accs, assets, func(simAcc simtypes.Account, asset types.AssetParam) bool {
			supply, found := k.GetAssetSupply(ctx, asset.Denom)
			if !found {
				return false
			}
			if supply.CurrentSupply.Amount.IsPositive() {
				authAcc := ak.GetAccount(ctx, simAcc.Address)
				// deputy cannot be sender of outgoing swap
				if authAcc.GetAddress().String() == asset.DeputyAddress {
					return false
				}
				// Search for an account that holds coins received by an atomic swap
				minAmountPlusFee := asset.MinSwapAmount.Add(asset.FixedFee)
				if bk.SpendableCoins(ctx.WithBlockTime(ctx.BlockTime()), simAcc.Address).
					AmountOf(asset.Denom).GT(minAmountPlusFee) {
					return true
				}
			}
			return false
		})
		var sender simtypes.Account
		var recipient simtypes.Account
		var asset types.AssetParam

		depAddr, err := sdk.AccAddressFromBech32(selectedAsset.DeputyAddress)
		if err != nil {
			return simtypes.NewOperationMsg(&types.MsgCreateAtomicSwap{}, false, fmt.Sprintf("%+v", err)), nil, err
		}

		// If an outgoing swap can be created, it's chosen 50% of the time.
		if found && r.Intn(100) < 50 {
			deputy, found := simtypes.FindAccount(accs, depAddr)
			if !found {
				return noOpMsg, nil, nil
			}
			sender = senderOutgoing
			recipient = deputy
			asset = selectedAsset
		} else {
			// if an outgoing swap cannot be created or was not selected, simulate an incoming swap
			assets, _ := k.GetAssets(ctx)
			asset = assets[r.Intn(len(assets))]
			var eligibleAccs []simtypes.Account
			for _, simAcc := range accs {
				// don't allow recipient of incoming swap to be the deputy
				if simAcc.Address.Equals(depAddr) {
					continue
				}
				eligibleAccs = append(eligibleAccs, simAcc)
			}
			recipient, _ = simtypes.RandomAcc(r, eligibleAccs)
			deputy, found := simtypes.FindAccount(accs, depAddr)
			if !found {
				return noOpMsg, nil, nil
			}
			sender = deputy

		}

		recipientOtherChain := simtypes.RandStringOfLength(r, 43)
		senderOtherChain := simtypes.RandStringOfLength(r, 43)

		// Use same random number for determinism
		timestamp := ctx.BlockTime().Unix()
		randomNumberHash := types.CalculateRandomHash(randomNumber, timestamp)

		// Check that the sender has coins for fee
		senderAcc := ak.GetAccount(ctx, sender.Address)
		fees, err := simtypes.RandomFees(r, ctx, bk.
			SpendableCoins(ctx.WithBlockTime(ctx.BlockTime()), sender.Address))
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, ""), nil, err
		}

		// Get maximum valid amount
		maximumAmount := bk.SpendableCoins(ctx.WithBlockTime(ctx.BlockTime()), sender.Address).
			Sub(fees).AmountOf(asset.Denom)
		assetSupply, foundAssetSupply := k.GetAssetSupply(ctx, asset.Denom)
		if !foundAssetSupply {
			return noOpMsg, nil, fmt.Errorf("no asset supply found for %s", asset.Denom)
		}
		// The maximum amount for outgoing swaps is limited by the asset's current supply
		if recipient.Address.Equals(depAddr) {
			if maximumAmount.GT(assetSupply.CurrentSupply.Amount.Sub(assetSupply.OutgoingSupply.Amount)) {
				maximumAmount = assetSupply.CurrentSupply.Amount.Sub(assetSupply.OutgoingSupply.Amount)
			}
		} else {
			// the maximum amount for incoming swaps in limited by the asset's incoming supply + current supply (rate-limited if applicable)  + swap amount being less than the supply limit
			var currentRemainingSupply sdk.Int
			if asset.SupplyLimit.TimeLimited {
				currentRemainingSupply = asset.SupplyLimit.Limit.Sub(assetSupply.IncomingSupply.Amount).Sub(assetSupply.TimeLimitedCurrentSupply.Amount)
			} else {
				currentRemainingSupply = asset.SupplyLimit.Limit.Sub(assetSupply.IncomingSupply.Amount).Sub(assetSupply.CurrentSupply.Amount)
			}
			if currentRemainingSupply.LT(maximumAmount) {
				maximumAmount = currentRemainingSupply
			}
		}

		// The maximum amount for all swaps is limited by the total max limit
		if maximumAmount.GT(asset.MaxSwapAmount) {
			maximumAmount = asset.MaxSwapAmount
		}

		// Get an amount of coins between 0.1 and 2% of total coins
		amount := maximumAmount.Quo(sdk.NewInt(int64(simtypes.RandIntBetween(r, 50, 1000))))
		minAmountPlusFee := asset.MinSwapAmount.Add(asset.FixedFee)
		if amount.LT(minAmountPlusFee) {
			return simtypes.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (account funds exhausted for asset %s)", asset.Denom), "", false, nil), nil, nil
		}
		coins := sdk.NewCoins(sdk.NewCoin(asset.Denom, amount))

		msg := types.NewMsgCreateAtomicSwap(
			sender.Address, recipient.Address, recipientOtherChain, senderOtherChain,
			randomNumberHash, timestamp, coins, asset.SwapTimeSpan,
		)

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{senderAcc.GetAccountNumber()},
			[]uint64{senderAcc.GetSequence()},
			sender.PrivKey,
		)
		if err != nil {
			return simtypes.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}

		_, result, err := app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}

		// Construct a MsgClaimAtomicSwap or MsgRefundAtomicSwap future operation
		var futureOp simtypes.FutureOperation

		fromAddr, err := sdk.AccAddressFromBech32(selectedAsset.DeputyAddress)
		if err != nil {
			return simtypes.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}
		swapID := types.CalculateSwapID(msg.RandomNumberHash, fromAddr, msg.SenderOtherChain)
		if r.Intn(100) < 50 {
			// Claim future operation - choose between next block and the block before time span
			executionTime :=
				time.Unix(ctx.BlockTime().Unix()+1+int64(r.Intn(int(asset.SwapTimeSpan-1))), 0)

			futureOp = simtypes.FutureOperation{
				BlockTime: executionTime,
				Op:        operationClaimAtomicSwap(ak, bk, k, swapID, randomNumber),
			}
		} else {
			// Refund future operation
			executionTime := time.Unix(ctx.BlockTime().Unix()+msg.TimeSpan, 0)
			futureOp = simtypes.FutureOperation{
				BlockTime: executionTime,
				Op:        operationRefundAtomicSwap(ak, bk, k, swapID),
			}
		}

		return simtypes.NewOperationMsg(msg, true, result.Log), []simtypes.FutureOperation{futureOp}, nil
	}
}

func operationClaimAtomicSwap(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, swapID []byte, randomNumber []byte) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		acc := ak.GetAccount(ctx, simAccount.Address)

		swap, found := k.GetAtomicSwap(ctx, swapID)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, types.ClaimAtomicSwap, "swap ID not found"), nil, fmt.Errorf("cannot claim: swap with ID %s not found", swapID)
		}
		// check that asset supply supports claiming (it could have changed due to a param change proposal)
		// use CacheContext so changes don't take effect
		cacheCtx, _ := ctx.CacheContext()
		switch swap.Direction {
		case types.Incoming:
			err := k.DecrementIncomingAssetSupply(cacheCtx, swap.Amount[0])
			if err != nil {
				return simtypes.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - unable to decrement incoming asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
			err = k.IncrementCurrentAssetSupply(cacheCtx, swap.Amount[0])
			if err != nil {
				return simtypes.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - unable to increment current asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
		case types.Outgoing:
			err := k.DecrementOutgoingAssetSupply(cacheCtx, swap.Amount[0])
			if err != nil {
				return simtypes.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - unable to decrement outgoing asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
			err = k.DecrementCurrentAssetSupply(cacheCtx, swap.Amount[0])
			if err != nil {
				return simtypes.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - unable to decrement current asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
		}

		asset, err := k.GetAsset(ctx, swap.Amount[0].Denom)
		if err != nil {
			return simtypes.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - asset not found %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
		}
		supply, found := k.GetAssetSupply(ctx, asset.Denom)
		if !found {
			return simtypes.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - asset supply not found %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
		}
		if asset.SupplyLimit.Limit.LT(supply.CurrentSupply.Amount.Add(swap.Amount[0].Amount)) {
			return simtypes.NoOpMsg(types.ModuleName, types.ClaimAtomicSwap,
				fmt.Sprintf("supplyLimit %s less than current supply %s + swap amount %s",
					asset.SupplyLimit.Limit.String(),
					supply.CurrentSupply.Amount.String(),
					swap.Amount[0].Amount.String())), nil, nil
		}

		msg := types.NewMsgClaimAtomicSwap(acc.GetAddress(), swapID, randomNumber)
		fees, err := simtypes.RandomFees(r, ctx, bk.SpendableCoins(ctx.WithBlockTime(ctx.BlockTime()), acc.GetAddress()))
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "MsgClaimAtomicSwap", "RandomFees error"), nil, err
		}
		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{acc.GetAccountNumber()},
			[]uint64{acc.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}

		_, result, err := app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}
		return simtypes.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

func operationRefundAtomicSwap(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, swapID []byte) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		acc := ak.GetAccount(ctx, simAccount.Address)

		swap, found := k.GetAtomicSwap(ctx, swapID)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "AtomicSwap", "Atomic Swap not found during refund attempt"), nil, fmt.Errorf("cannot refund: swap with ID %s not found", swapID)
		}
		cacheCtx, _ := ctx.CacheContext()
		switch swap.Direction {
		case types.Incoming:
			if err := k.DecrementIncomingAssetSupply(cacheCtx, swap.Amount[0]); err != nil {
				return simtypes.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not refund - unable to decrement incoming asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
		case types.Outgoing:
			if err := k.DecrementOutgoingAssetSupply(cacheCtx, swap.Amount[0]); err != nil {
				return simtypes.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not refund - unable to decrement outgoing asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
		}

		msg := types.NewMsgRefundAtomicSwap(acc.GetAddress(), swapID)

		fees, err := simtypes.RandomFees(r, ctx, bk.SpendableCoins(ctx.WithBlockTime(ctx.BlockTime()), acc.GetAddress()))
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName,
				"MsgRefundAtomicSwap", "RandomFees error during refund"), nil, err
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{acc.GetAccountNumber()},
			[]uint64{acc.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}

		_, result, err := app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}
		return simtypes.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

// findValidAccountAssetSupplyPair finds an account for which the callback func returns true
func findValidAccountAssetPair(accounts []simtypes.Account, assets types.AssetParams,
	cb func(simtypes.Account, types.AssetParam) bool) (simtypes.Account, types.AssetParam, bool) {
	for _, asset := range assets {
		for _, acc := range accounts {
			if isValid := cb(acc, asset); isValid {
				return acc, asset, true
			}
		}
	}
	return simtypes.Account{}, types.AssetParam{}, false
}
