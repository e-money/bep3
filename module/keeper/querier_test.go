package keeper_test

import (
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/bep3/module/keeper"
	"github.com/e-money/bep3/module/types"
	app "github.com/e-money/bep3/testapp"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"strings"
	"testing"
)

const (
	custom = "custom"
)

type QuerierTestSuite struct {
	suite.Suite
	keeper        keeper.Keeper
	ctx           sdk.Context
	querier       sdk.Querier
	addrs         []sdk.AccAddress
	isSupplyDenom map[string]bool
	swapIDs       []tmbytes.HexBytes
	isSwapID      map[string]bool
	marshaller    codec.JSONMarshaler
}

func (suite *QuerierTestSuite) SetupTest() {
	ctx, jsonMarshaller, bep3Keeper, accountKeeper, bankKeeper, appModule := app.CreateTestComponents(suite.T())

	_, addrs := app.GeneratePrivKeyAddressPairs(11)
	coins := cs(c("bnb", 10000000000), c("ukava", 10000000000))

	for _, addr := range addrs {
		account := accountKeeper.NewAccountWithAddress(ctx, addr)
		if err := bankKeeper.SetBalances(ctx, addr, coins); err != nil {
			panic(err)
		}
		accountKeeper.SetAccount(ctx, account)
	}

	appModule.InitGenesis(ctx, jsonMarshaller, NewBep3GenState(addrs[10]))

	suite.ctx = ctx
	suite.keeper = bep3Keeper
	suite.querier = keeper.NewQuerier(suite.keeper)
	suite.addrs = addrs
	suite.marshaller = jsonMarshaller

	// Create atomic swaps and save IDs
	var swapIDs []tmbytes.HexBytes
	isSwapID := make(map[string]bool)
	for i := 0; i < 10; i++ {
		// Set up atomic swap variables
		expireTimestamp := types.DefaultSwapBlockTimestamp + types.DefaultSwapTimeSpan
		amount := cs(c("bnb", 100))
		timestamp := ts(0)
		randomNumber, _ := types.GenerateSecureRandomNumber()
		randomNumberHash := types.CalculateRandomHash(randomNumber[:], timestamp)

		// Create atomic swap and check err
		err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, expireTimestamp,
			addrs[10], suite.addrs[i], TestSenderOtherChain, TestRecipientOtherChain, amount, true)
		suite.Nil(err)

		// Calculate swap ID and save
		swapID := types.CalculateSwapID(randomNumberHash, addrs[10], TestSenderOtherChain)
		swapIDs = append(swapIDs, swapID)
		isSwapID[hex.EncodeToString(swapID)] = true
	}
	suite.swapIDs = swapIDs
	suite.isSwapID = isSwapID
}

func (suite *QuerierTestSuite) TestQueryAssetSupply() {
	const denom = "bnb"
	ctx := suite.ctx.WithIsCheckTx(false)

	// Set up request query
	qAssetSupply := types.NewQueryAssetSupply(denom)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAssetSupply}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(&qAssetSupply),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryGetAssetSupply}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	// Unmarshal the bytes into type asset supply
	var supply types.AssetSupply
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &supply))

	expectedSupply := types.NewAssetSupply(c(denom, 1000),
		c(denom, 0), c(denom, 0), c(denom, 0), 0)
	suite.Equal(supply, expectedSupply)
}

func (suite *QuerierTestSuite) TestQueryAtomicSwap() {
	ctx := suite.ctx.WithIsCheckTx(false)

	// Set up request query
	qSwapByID := types.NewQueryAtomicSwapByID(suite.swapIDs[0])
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAtomicSwap}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(&qSwapByID),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryGetAtomicSwap}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	// Unmarshal the bytes into type atomic swap
	var swap types.AugmentedAtomicSwap
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &swap))

	// Check the returned atomic swap's ID
	suite.True(suite.isSwapID[swap.ID])
}

func (suite *QuerierTestSuite) TestQueryAssetSupplies() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	qAssetSupplies := types.NewQueryAssetSupplies(1, 100)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAssetSupplies}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(&qAssetSupplies),
	}

	bz, err := suite.querier(ctx, []string{types.QueryGetAssetSupplies}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	supplies := types.AssetSupplies{}
	err = suite.marshaller.UnmarshalJSON(bz, &supplies)
	suite.Nil(err)

	// Check that returned value matches asset supplies in state
	storeSupplies := suite.keeper.GetAllAssetSupplies(ctx)
	suite.Equal(len(storeSupplies.GetAssetSupplies()), len(supplies.GetAssetSupplies()))
	suite.Equal(supplies, storeSupplies)
}

func (suite *QuerierTestSuite) TestQueryAtomicSwaps() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	qSwaps := types.NewQueryAtomicSwaps(1, 100, sdk.AccAddress{}, 0, types.Open, types.Incoming)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAtomicSwaps}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(&qSwaps),
	}

	bz, err := suite.querier(ctx, []string{types.QueryGetAtomicSwaps}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	var swaps types.AugmentedAtomicSwaps
	suite.Nil(suite.marshaller.UnmarshalJSON(bz, &swaps))

	suite.Equal(len(suite.swapIDs), len(swaps.AugmentedAtomicSwaps))
	for _, swap := range swaps.AugmentedAtomicSwaps {
		suite.True(suite.isSwapID[swap.ID])
	}
}

func (suite *QuerierTestSuite) TestQueryParams() {
	ctx := suite.ctx.WithIsCheckTx(false)
	bz, err := suite.querier(ctx, []string{types.QueryGetParams}, abci.RequestQuery{})
	suite.Nil(err)
	suite.NotNil(bz)

	var p types.Params
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &p))

	bep3GenesisState := NewBep3GenState(suite.addrs[10])
	gs := types.GenesisState{}
	types.ModuleCdc.UnmarshalJSON(bep3GenesisState, &gs)
	// update asset supply to account for swaps that were created in setup
	suite.Equal(gs.Params, p)
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QuerierTestSuite))
}
