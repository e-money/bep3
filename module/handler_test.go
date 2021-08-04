package bep3_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bep3 "github.com/e-money/bep3/module"
	app "github.com/e-money/bep3/testapp"
	"github.com/stretchr/testify/suite"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

type HandlerTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	handler sdk.Handler
	keeper  bep3.Keeper
	addrs   []sdk.AccAddress
}

func (suite *HandlerTestSuite) SetupTest() {
	ctx, jsonMarshaller, bep3Keeper, accountKeeper, bankKeeper, appModule := app.CreateTestComponents(suite.T())

	// Set up genesis state and initialize
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	coins := cs(c("bnb", 10000000000), c("ukava", 10000000000))

	for _, addr := range addrs {
		account := accountKeeper.NewAccountWithAddress(ctx, addr)
		if err := bep3.FundAccount(ctx, bankKeeper, addr, coins); err != nil {
			panic(err)
		}
		accountKeeper.SetAccount(ctx, account)
	}

	appModule.InitGenesis(ctx, jsonMarshaller, NewBep3GenState(addrs[0]))

	suite.addrs = addrs
	suite.handler = bep3.NewHandler(bep3Keeper)
	suite.keeper = bep3Keeper
	suite.ctx = ctx
}

func (suite *HandlerTestSuite) AddAtomicSwap() (tmbytes.HexBytes, tmbytes.HexBytes) {
	expireTimeSpan := bep3.DefaultSwapTimeSpanMinutes
	amount := cs(c("bnb", int64(50000)))
	timestamp := ts(0)
	randomNumber, _ := bep3.GenerateSecureRandomNumber()
	randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)

	// Create atomic swap and check err to confirm creation

	_, err := suite.keeper.CreateAtomicSwapState(
		suite.ctx, randomNumberHash, timestamp, expireTimeSpan,
		suite.addrs[0], suite.addrs[1], TestSenderOtherChain,
		TestRecipientOtherChain,
		amount, true,
	)
	suite.Nil(err)

	swapID := bep3.CalculateSwapID(
		randomNumberHash, suite.addrs[0], TestSenderOtherChain,
	)
	return swapID, randomNumber[:]
}

func (suite *HandlerTestSuite) TestMsgCreateAtomicSwap() {
	amount := cs(c("bnb", int64(10000)))
	timestamp := ts(0)
	randomNumber, _ := bep3.GenerateSecureRandomNumber()
	randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)

	msg := bep3.NewMsgCreateAtomicSwap(
		suite.addrs[0].String(),
		suite.addrs[2].String(),
		TestRecipientOtherChain,
		TestSenderOtherChain,
		randomNumberHash,
		timestamp,
		amount,
		bep3.DefaultSwapTimeSpanMinutes,
	)

	res, err := suite.handler(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
}

func (suite *HandlerTestSuite) TestMsgClaimAtomicSwap() {
	// Attempt claim msg on fake atomic swap
	badRandomNumber, _ := bep3.GenerateSecureRandomNumber()
	badRandomNumberHash := bep3.CalculateRandomHash(badRandomNumber[:], ts(0))
	badSwapID := bep3.CalculateSwapID(
		badRandomNumberHash, suite.addrs[0], TestSenderOtherChain,
	)
	badMsg := bep3.NewMsgClaimAtomicSwap(
		suite.addrs[0], badSwapID, badRandomNumber[:],
	)
	badRes, err := suite.handler(suite.ctx, badMsg)
	suite.Require().Error(err)
	suite.Require().Nil(badRes)

	// Add an atomic swap before attempting new claim msg
	swapID, randomNumber := suite.AddAtomicSwap()
	msg := bep3.NewMsgClaimAtomicSwap(suite.addrs[0], swapID, randomNumber)
	res, err := suite.handler(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
}

// getContextPlusMinutes returns a context forward or backward in time and block
// index. Assuming 1 second finality.
func (suite *HandlerTestSuite) getContextPlusMinutes(plusMinutes int64) sdk.Context {
	offset := plusMinutes
	ctx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Duration(offset) * time.Minute))
	ctx = ctx.WithBlockTime(time.Unix(ctx.BlockTime().Unix()+offset, 0))

	return ctx
}

func (suite *HandlerTestSuite) TestMsgRefundAtomicSwap() {
	// Attempt refund msg on fake atomic swap
	badRandomNumber, _ := bep3.GenerateSecureRandomNumber()
	badRandomNumberHash := bep3.CalculateRandomHash(badRandomNumber[:], ts(0))
	badSwapID := bep3.CalculateSwapID(
		badRandomNumberHash, suite.addrs[0], TestSenderOtherChain,
	)
	badMsg := bep3.NewMsgRefundAtomicSwap(suite.addrs[0], badSwapID)
	badRes, err := suite.handler(suite.ctx, badMsg)
	suite.Require().Error(err)
	suite.Require().Nil(badRes)

	// Add an atomic swap and build refund msg
	swapID, _ := suite.AddAtomicSwap()
	msg := bep3.NewMsgRefundAtomicSwap(suite.addrs[0], swapID)

	// Attempt to refund active atomic swap
	res1, err := suite.handler(suite.ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res1)

	// Expire the atomic swap with begin blocker and attempt refund
	laterCtx := suite.getContextPlusMinutes(bep3.DefaultSwapTimeSpanMinutes)
	bep3.BeginBlocker(laterCtx, suite.keeper)
	res2, err := suite.handler(laterCtx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res2)
}

func (suite *HandlerTestSuite) TestInvalidMsg() {
	res, err := suite.handler(suite.ctx, testdata.NewTestMsg())
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
