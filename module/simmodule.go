package bep3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	simmod "github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/e-money/bep3/module/keeper"
	"github.com/e-money/bep3/module/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

const (
	keyAssetParams = "AssetParams"
	// Simulation operation weights constants
	OpWeightMsgCreateAtomicSwap = "op_weight_msg_create_atomic_swap"
)

var (
	noOpMsg            = simtypes.NoOpMsg(types.ModuleName, "NoOpMsg", "")
	randomNumber       = []byte{114, 21, 74, 180, 81, 92, 21, 91, 173, 164, 143, 111, 120, 58, 241, 58, 40, 22, 59, 133, 102, 233, 55, 149, 12, 199, 231, 63, 122, 23, 88, 9}
	ConsistentDenoms   = [3]string{"bnb", "xrp", "btc"}
	MaxSupplyLimit     = 1000000000000
	MinSupplyLimit     = 100000000
	MinSwapAmountLimit = 999
	accs               []simtypes.Account
	MinBlockLock       = uint64(5)
)

func mustMarshalJSONIndent(o interface{}) []byte {
	bz, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("failed to JSON encode: %s", err))
	}

	return bz
}

// RandomizedGenState generates a random GenesisState
// https://github.com/cosmos/cosmos-sdk/blob/1c6e2679641d0892a3f35d778d7c2316a2937a7c/x/bank/simulation/genesis.go#L54
func RandomizedGenState(simState *module.SimulationState) {
	accs = simState.Accounts

	bep3Genesis := loadRandomBep3GenState(simState)
	simState.GenState[types.ModuleName] = mustMarshalJSONIndent(bep3Genesis)

	// Update bank supply to match amount of coins in auth
	bankGenesis, totalCoins := loadBankGenState(simState, bep3Genesis)

	for _, deputyCoin := range totalCoins {
		bankGenesis.Supply = bankGenesis.Supply.Add(deputyCoin...)
	}

	simState.GenState[banktypes.ModuleName] = mustMarshalJSONIndent(bankGenesis)
}

func loadBankGenState(simState *module.SimulationState, bep3Genesis types.GenesisState) (banktypes.GenesisState, []sdk.Coins) {
	var bankGenesis banktypes.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[banktypes.ModuleName], &bankGenesis)

	var (
		totalCoins      []sdk.Coins
		genesisBalances = []banktypes.Balance{}
	)

	// Load total limit of each supported asset to deputy's account
	for _, asset := range bep3Genesis.Params.AssetParams {

		assetCoin := sdk.NewCoins(sdk.NewCoin(asset.Denom, asset.SupplyLimit.Limit))
		genesisBalances = append(genesisBalances, banktypes.Balance{
			Address: asset.DeputyAddress,
			Coins:   assetCoin,
		})

		totalCoins = append(totalCoins, assetCoin)
	}

	bankGenesis.Balances = genesisBalances

	return bankGenesis, totalCoins
}

// GenSupportedAssets gets randomized SupportedAssets
func GenSupportedAssets(r *rand.Rand) types.AssetParams {
	numAssets := r.Intn(10) + 1
	assets := make(types.AssetParams, numAssets+1)
	for i := 0; i < numAssets; i++ {
		denom := strings.ToLower(simtypes.RandStringOfLength(r, r.Intn(3)+3))
		asset := genSupportedAsset(r, denom)
		assets[i] = asset
	}
	// Add bnb, btc, or xrp as a supported asset for interactions with other modules
	assets[len(assets)-1] = genSupportedAsset(r, ConsistentDenoms[r.Intn(3)])

	return assets
}

func genSupportedAsset(r *rand.Rand, denom string) types.AssetParam {
	coinID, _ := simtypes.RandPositiveInt(r, sdk.NewInt(100000))
	limit := GenSupplyLimit(r, MaxSupplyLimit)

	minSwapAmount := GenMinSwapAmount(r)
	timeLimited := r.Float32() < 0.5
	timeBasedLimit := sdk.ZeroInt()
	if timeLimited {
		// set time-based limit to between 10 and 25% of the total limit
		min := int(limit.Quo(sdk.NewInt(10)).Int64())
		max := int(limit.Quo(sdk.NewInt(4)).Int64())
		timeBasedLimit = sdk.NewInt(int64(simtypes.RandIntBetween(r, min, max)))
	}
	return types.AssetParam{
		Denom:  denom,
		CoinID: coinID.Int64(),
		SupplyLimit: types.SupplyLimit{
			Limit:          limit,
			TimeLimited:    timeLimited,
			TimePeriod:     int64(time.Hour * 24),
			TimeBasedLimit: timeBasedLimit,
		},
		Active:          true,
		DeputyAddress:   GenRandBnbDeputy(r).Address.String(),
		FixedFee:        GenRandFixedFee(r),
		MinSwapAmount:   minSwapAmount,
		MaxSwapAmount:   GenMaxSwapAmount(r, minSwapAmount, limit),
		SwapTimestamp:   time.Now().Unix(),
		SwapTimeSpanMin: limit.Int64(),
	}
}

func loadRandomBep3GenState(simState *module.SimulationState) types.GenesisState {
	supportedAssets := GenSupportedAssets(simState.Rand)
	supplies := types.AssetSupplies{}
	for _, asset := range supportedAssets {
		supply := GenAssetSupply(simState.Rand, asset.Denom)
		supplies.AssetSupplies = append(supplies.AssetSupplies, supply)
	}

	bep3Genesis := types.GenesisState{
		Params: types.Params{
			AssetParams: supportedAssets,
		},
		Supplies:          supplies,
		PreviousBlockTime: types.DefaultPreviousBlockTime,
	}

	return bep3Genesis
}

type CodecUnmarshaler interface {
	MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{})
}

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding type.
func NewDecodeStore(cdc CodecUnmarshaler) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.AtomicSwapKeyPrefix):
			var swapA, swapB types.AtomicSwap
			cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &swapA)
			cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &swapB)
			return fmt.Sprintf("%v\n%v", swapA, swapB)

		case bytes.Equal(kvA.Key[:1], types.AtomicSwapByBlockPrefix),
			bytes.Equal(kvA.Key[:1], types.AtomicSwapLongtermStoragePrefix):
			var bytesA tmbytes.HexBytes = kvA.Value
			var bytesB tmbytes.HexBytes = kvA.Value
			return fmt.Sprintf("%s\n%s", bytesA.String(), bytesB.String())
		case bytes.Equal(kvA.Key[:1], types.AssetSupplyPrefix):
			var supplyA, supplyB types.AssetSupply
			cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &supplyA)
			cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &supplyB)
			return fmt.Sprintf("%s\n%s", supplyA, supplyB)
		case bytes.Equal(kvA.Key[:1], types.PreviousBlockTimeKey):
			var timeA, timeB time.Time
			cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &timeA)
			cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &timeB)
			return fmt.Sprintf("%s\n%s", timeA, timeB)

		default:
			panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
		}
	}
}

func defaultWeightMsgCreateAtomicSwap(r *rand.Rand) int {
	return simtypes.RandIntBetween(r, 20, 100)
}

// ParamChanges defines the parameters that can be modified by param change proposals
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simmod.NewSimParamChange(types.ModuleName, keyAssetParams,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenSupportedAssets(r))
			},
		),
	}
}

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simmod.WeightedOperations {
	var weightCreateAtomicSwap int

	// GenDepositParamsDepositPeriod randomized DepositParamsDepositPeriod
	appParams.GetOrGenerate(cdc, OpWeightMsgCreateAtomicSwap, &weightCreateAtomicSwap, nil,
		func(r *rand.Rand) {
			weightCreateAtomicSwap = defaultWeightMsgCreateAtomicSwap(r)
		},
	)

	return simmod.WeightedOperations{
		simmod.NewWeightedOperation(
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
			sender.Address.String(),
			recipient.Address.String(),
			recipientOtherChain,
			senderOtherChain,
			randomNumberHash,
			timestamp,
			coins,
			asset.SwapTimeSpanMin,
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
				time.Unix(ctx.BlockTime().Unix()+1+int64(r.Intn(int(asset.SwapTimeSpanMin-1))), 0)

			futureOp = simtypes.FutureOperation{
				BlockTime: executionTime,
				Op:        operationClaimAtomicSwap(ak, bk, k, swapID, randomNumber),
			}
		} else {
			// Refund future operation
			executionTime := time.Unix(ctx.BlockTime().Unix()+msg.TimeSpanMin, 0)
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

// GenSupplyLimit generates a random SupplyLimit
func GenSupplyLimit(r *rand.Rand, max int) sdk.Int {
	max = simtypes.RandIntBetween(r, MinSupplyLimit, max)
	return sdk.NewInt(int64(max))
}

// GenSupplyLimit generates a random SupplyLimit
func GenAssetSupply(r *rand.Rand, denom string) types.AssetSupply {
	return types.NewAssetSupply(
		sdk.NewCoin(denom, sdk.ZeroInt()), sdk.NewCoin(denom, sdk.ZeroInt()),
		sdk.NewCoin(denom, sdk.ZeroInt()), sdk.NewCoin(denom, sdk.ZeroInt()), 0)
}

// GenMinBlockLock randomized MinBlockLock
func GenMinBlockLock(r *rand.Rand) uint64 {
	return MinBlockLock
}

// GenMaxBlockLock randomized MaxBlockLock
func GenMaxBlockLock(r *rand.Rand, minBlockLock uint64) uint64 {
	max := int(50)
	return uint64(r.Intn(max-int(MinBlockLock)) + int(MinBlockLock+1))
}

// GenRandBnbDeputy randomized BnbDeputyAddress
func GenRandBnbDeputy(r *rand.Rand) simtypes.Account {
	acc, _ := simtypes.RandomAcc(r, accs)
	return acc
}

// GenRandFixedFee randomized FixedFee in range [1, 10000]
func GenRandFixedFee(r *rand.Rand) sdk.Int {
	min := int(1)
	max := types.DeputyFee
	return sdk.NewInt(int64(r.Intn(int(max)-min) + min))
}

// GenMinSwapAmount randomized MinAmount in range [1, 1000]
func GenMinSwapAmount(r *rand.Rand) sdk.Int {
	return sdk.OneInt().Add(simtypes.RandomAmount(r, sdk.NewInt(int64(MinSwapAmountLimit))))
}

// GenMaxSwapAmount randomized MaxAmount
func GenMaxSwapAmount(r *rand.Rand, minAmount sdk.Int, supplyMax sdk.Int) sdk.Int {
	min := minAmount.Int64()
	max := supplyMax.Quo(sdk.NewInt(100)).Int64()

	return sdk.NewInt((int64(r.Intn(int(max-min))) + min))
}
