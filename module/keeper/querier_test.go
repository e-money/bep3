package keeper_test

import (
	"encoding/hex"
	app "github.com/e-money/bep3/testapp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/e-money/bep3/module/keeper"
	"github.com/e-money/bep3/module/types"
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
}

func (suite *QuerierTestSuite) SetupTest() {
	ctx, bep3Keeper, accountKeeper, _, appModule := app.CreateTestComponents(suite.T())

	_, addrs := app.GeneratePrivKeyAddressPairs(11)
	coins := cs(c("bnb", 10000000000), c("ukava", 10000000000))

	for _, addr := range addrs {
		account := accountKeeper.NewAccountWithAddress(ctx, addr)
		account.SetCoins(coins)
		accountKeeper.SetAccount(ctx, account)
	}

	appModule.InitGenesis(ctx, NewBep3GenState(addrs[10]))

	suite.ctx = ctx
	suite.keeper = bep3Keeper
	suite.querier = keeper.NewQuerier(suite.keeper)
	suite.addrs = addrs

	// Create atomic swaps and save IDs
	var swapIDs []tmbytes.HexBytes
	isSwapID := make(map[string]bool)
	for i := 0; i < 10; i++ {
		// Set up atomic swap variables
		expireHeight := types.DefaultMinBlockLock
		amount := cs(c("bnb", 100))
		timestamp := ts(0)
		randomNumber, _ := types.GenerateSecureRandomNumber()
		randomNumberHash := types.CalculateRandomHash(randomNumber[:], timestamp)

		// Create atomic swap and check err
		err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, expireHeight,
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
	ctx := suite.ctx.WithIsCheckTx(false)

	// Set up request query
	denom := "bnb"
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAssetSupply}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryAssetSupply(denom)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryGetAssetSupply}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	// Unmarshal the bytes into type asset supply
	var supply types.AssetSupply
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &supply))

	expectedSupply := types.NewAssetSupply(c(denom, 1000),
		c(denom, 0), c(denom, 0), c(denom, 0), time.Duration(0))
	suite.Equal(supply, expectedSupply)
}

func (suite *QuerierTestSuite) TestQueryAtomicSwap() {
	ctx := suite.ctx.WithIsCheckTx(false)

	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAtomicSwap}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryAtomicSwapByID(suite.swapIDs[0])),
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
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAssetSupplies}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryAssetSupplies(1, 100)),
	}

	bz, err := suite.querier(ctx, []string{types.QueryGetAssetSupplies}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	var supplies types.AssetSupplies
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &supplies))

	// Check that returned value matches asset supplies in state
	storeSupplies := suite.keeper.GetAllAssetSupplies(ctx)
	suite.Equal(len(storeSupplies), len(supplies))
	suite.Equal(supplies, storeSupplies)
}

func (suite *QuerierTestSuite) TestQueryAtomicSwaps() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAtomicSwaps}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryAtomicSwaps(1, 100, sdk.AccAddress{}, 0, types.Open, types.Incoming)),
	}

	bz, err := suite.querier(ctx, []string{types.QueryGetAtomicSwaps}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	var swaps types.AugmentedAtomicSwaps
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &swaps))

	suite.Equal(len(suite.swapIDs), len(swaps))
	for _, swap := range swaps {
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
