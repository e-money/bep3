package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bep3 "github.com/e-money/bep3/module"
	"github.com/e-money/bep3/module/keeper"
	"github.com/e-money/bep3/module/types"
	app "github.com/e-money/bep3/testapp"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type AtomicSwapTestSuite struct {
	suite.Suite

	keeper             keeper.Keeper
	accountKeeper      types.AccountKeeper
	bankKeeper         types.BankKeeper
	ctx                sdk.Context
	randMacc           sdk.AccAddress
	deputy             sdk.AccAddress
	addrs              []sdk.AccAddress
	timestamps         []int64
	randomNumberHashes []tmbytes.HexBytes
	swapIDs            []tmbytes.HexBytes
	randomNumbers      []tmbytes.HexBytes
}

const (
	STARING_BNB_BALANCE   = int64(3000000000000)
	BNB_DENOM             = "bnb"
	OTHER_DENOM           = "inc"
	STARING_OTHER_BALANCE = int64(3000000000000)
)

func (suite *AtomicSwapTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	ctx, jsonMarshaller, bep3Keeper, accountKeeper, bankKeeper, appModule := app.CreateTestComponents(suite.T())

	// Generate a random moduleAccount for the test
	key := ed25519.GenPrivKey()
	suite.randMacc = sdk.AccAddress(key.PubKey().Address())
	bep3Keeper.Maccs[suite.randMacc.String()] = true

	// Create and load 20 accounts with bnb tokens
	coins := cs(c(BNB_DENOM, STARING_BNB_BALANCE), c(OTHER_DENOM, STARING_OTHER_BALANCE))
	_, addrs := app.GeneratePrivKeyAddressPairs(20)
	deputy := addrs[0]

	for _, addr := range addrs {
		account := accountKeeper.NewAccountWithAddress(ctx, addr)
		bankKeeper.SetBalances(ctx, addr, coins)
		accountKeeper.SetAccount(ctx, account)
	}

	appModule.InitGenesis(ctx, jsonMarshaller, NewBep3GenState(deputy))

	params := bep3Keeper.GetParams(ctx)
	params.AssetParams[1].Active = true
	bep3Keeper.SetParams(ctx, params)

	suite.ctx = ctx
	suite.deputy = deputy
	suite.addrs = addrs
	suite.keeper = bep3Keeper
	suite.accountKeeper = accountKeeper
	suite.bankKeeper = bankKeeper

	suite.GenerateSwapDetails()
}

func (suite *AtomicSwapTestSuite) GenerateSwapDetails() {
	var timestamps []int64
	var randomNumberHashes []tmbytes.HexBytes
	var randomNumbers []tmbytes.HexBytes
	for i := 0; i < 15; i++ {
		// Set up atomic swap details
		timestamp := ts(i)
		randomNumber, _ := types.GenerateSecureRandomNumber()
		randomNumberHash := types.CalculateRandomHash(randomNumber[:], timestamp)

		timestamps = append(timestamps, timestamp)
		randomNumberHashes = append(randomNumberHashes, randomNumberHash)
		randomNumbers = append(randomNumbers, randomNumber[:])
	}
	suite.timestamps = timestamps
	suite.randomNumberHashes = randomNumberHashes
	suite.randomNumbers = randomNumbers
}

func (suite *AtomicSwapTestSuite) TestCreateAtomicSwap() {
	currentTmTime := tmtime.Now()
	type args struct {
		randomNumberHash    []byte
		timestamp           int64
		timeSpan            int64
		sender              sdk.AccAddress
		recipient           sdk.AccAddress
		senderOtherChain    string
		recipientOtherChain string
		coins               sdk.Coins
		crossChain          bool
		direction           types.SwapDirection
	}
	testCases := []struct {
		name          string
		blockTime     time.Time
		args          args
		expectPass    bool
		shouldBeFound bool
	}{
		{
			"incoming swap",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[0],
				timestamp:           suite.timestamps[0],
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.deputy,
				recipient:           suite.addrs[1],
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 50000)),
				crossChain:          true,
				direction:           types.Incoming,
			},
			true,
			true,
		},
		{
			"incoming swap rate limited",
			currentTmTime.Add(time.Minute * 10),
			args{
				randomNumberHash:    suite.randomNumberHashes[12],
				timestamp:           suite.timestamps[12],
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.deputy,
				recipient:           suite.addrs[1],
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c("inc", 50000000000)),
				crossChain:          true,
				direction:           types.Incoming,
			},
			true,
			true,
		},
		{
			"incoming swap over rate limit",
			currentTmTime.Add(time.Minute * 10),
			args{
				randomNumberHash:    suite.randomNumberHashes[13],
				timestamp:           suite.timestamps[13],
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.deputy,
				recipient:           suite.addrs[1],
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c("inc", 50000000001)),
				crossChain:          true,
				direction:           types.Incoming,
			},
			false,
			false,
		},
		{
			"outgoing swap",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[0],
				timestamp:           suite.timestamps[0],
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.addrs[1],
				recipient:           suite.deputy,
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 50000)),
				crossChain:          true,
				direction:           types.Outgoing,
			},
			true,
			true,
		},
		{

			"outgoing swap amount not greater than fixed fee",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[1],
				timestamp:           suite.timestamps[1],
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.addrs[1],
				recipient:           suite.addrs[2],
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 1000)),
				crossChain:          true,
				direction:           types.Outgoing,
			},
			false,
			false,
		},
		{
			"unsupported asset",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[2],
				timestamp:           suite.timestamps[2],
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.deputy,
				recipient:           suite.addrs[2],
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c("xyz", 50000)),
				crossChain:          true,
				direction:           types.Incoming,
			},
			false,
			false,
		},
		{
			"outside timestamp range",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[3],
				timestamp:           suite.timestamps[3] - 2000,
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.deputy,
				recipient:           suite.addrs[3],
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 50000)),
				crossChain:          true,
				direction:           types.Incoming,
			},
			false,
			false,
		},
		{
			"future timestamp",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[4],
				timestamp:           suite.timestamps[4] + 5000,
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.deputy,
				recipient:           suite.addrs[4],
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 50000)),
				crossChain:          true,
				direction:           types.Incoming,
			},
			false,
			false,
		},
		{
			"small time span on outgoing swap",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[5],
				timestamp:           suite.timestamps[5],
				timeSpan:            types.DefaultSwapTimeSpan - 1,
				sender:              suite.addrs[5],
				recipient:           suite.deputy,
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 50000)),
				crossChain:          true,
				direction:           types.Outgoing,
			},
			false,
			false,
		},
		{
			"big height span on outgoing swap",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[6],
				timestamp:           suite.timestamps[6],
				timeSpan:            types.ThreeDaySeconds + 1,
				sender:              suite.addrs[6],
				recipient:           suite.deputy,
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 50000)),
				crossChain:          true,
				direction:           types.Outgoing,
			},
			false,
			false,
		},
		{
			"zero amount",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[7],
				timestamp:           suite.timestamps[7],
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.deputy,
				recipient:           suite.addrs[7],
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 0)),
				crossChain:          true,
				direction:           types.Incoming,
			},
			false,
			false,
		},
		{
			"duplicate swap",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[0],
				timestamp:           suite.timestamps[0],
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.deputy,
				recipient:           suite.addrs[1],
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 50000)),
				crossChain:          true,
				direction:           types.Incoming,
			},
			false,
			true,
		},
		{
			"recipient is module account",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[8],
				timestamp:           suite.timestamps[8],
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.deputy,
				recipient:           suite.randMacc,
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 5000)),
				crossChain:          true,
				direction:           types.Incoming,
			},
			false,
			false,
		},
		{
			"exactly at maximum amount",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[10],
				timestamp:           suite.timestamps[10],
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.deputy,
				recipient:           suite.addrs[4],
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 1000000000000)), // 10,000 BNB
				crossChain:          true,
				direction:           types.Incoming,
			},
			true,
			true,
		},
		{
			"above maximum amount",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[11],
				timestamp:           suite.timestamps[11],
				timeSpan:            types.DefaultSwapTimeSpan,
				sender:              suite.deputy,
				recipient:           suite.addrs[5],
				senderOtherChain:    TestSenderOtherChain,
				recipientOtherChain: TestRecipientOtherChain,
				coins:               cs(c(BNB_DENOM, 1000000000001)), // 10,001 BNB
				crossChain:          true,
				direction:           types.Incoming,
			},
			false,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Increment current asset supply to support outgoing swaps
			suite.ctx = suite.ctx.WithBlockTime(tc.blockTime)
			if tc.args.direction == types.Outgoing {
				err := suite.keeper.IncrementCurrentAssetSupply(suite.ctx, tc.args.coins[0])
				suite.Nil(err)
			}

			// Load asset denom (required for zero coins test case)
			var swapAssetDenom string
			if len(tc.args.coins) == 1 {
				swapAssetDenom = tc.args.coins[0].Denom
			} else {
				swapAssetDenom = BNB_DENOM
			}

			// Load sender's account prior to swap creation
			bk := suite.bankKeeper
			senderBalancePre := bk.GetBalance(suite.ctx, tc.args.sender, swapAssetDenom)
			assetSupplyPre, _ := suite.keeper.GetAssetSupply(suite.ctx, swapAssetDenom)

			// Create atomic swap
			err := suite.keeper.CreateAtomicSwap(suite.ctx, tc.args.randomNumberHash, tc.args.timestamp,
				tc.args.timeSpan, tc.args.sender, tc.args.recipient, tc.args.senderOtherChain,
				tc.args.recipientOtherChain, tc.args.coins, tc.args.crossChain)

			// Load sender's account after swap creation
			senderBalancePost := bk.GetAllBalances(suite.ctx, tc.args.sender)

			assetSupplyPost, _ := suite.keeper.GetAssetSupply(suite.ctx, swapAssetDenom)

			// Load expected swap ID
			expectedSwapID := types.CalculateSwapID(tc.args.randomNumberHash, tc.args.sender, tc.args.senderOtherChain)

			if tc.expectPass {
				suite.NoError(err)

				// Check incoming/outgoing asset supply increased
				switch tc.args.direction {
				case types.Incoming:
					suite.Equal(assetSupplyPre.IncomingSupply.Add(tc.args.coins[0]), assetSupplyPost.IncomingSupply)
				case types.Outgoing:
					// Check coins moved
					suite.Equal(senderBalancePre.Amount.Sub(tc.args.coins[0].Amount), senderBalancePost)
					suite.Equal(assetSupplyPre.OutgoingSupply.Add(tc.args.coins[0]), assetSupplyPost.OutgoingSupply)
				default:
					suite.Fail("should not have invalid direction")
				}

				// Check swap in store
				actualSwap, found := suite.keeper.GetAtomicSwap(suite.ctx, expectedSwapID)
				suite.True(found)
				suite.NotNil(actualSwap)

				// Confirm swap contents
				expectedSwap :=
					types.AtomicSwap{
						Amount:              tc.args.coins,
						RandomNumberHash:    tc.args.randomNumberHash,
						ExpireTimestamp:     suite.ctx.BlockTime().Unix() + tc.args.timeSpan,
						Timestamp:           tc.args.timestamp,
						Sender:              tc.args.sender.String(),
						Recipient:           tc.args.recipient.String(),
						SenderOtherChain:    tc.args.senderOtherChain,
						RecipientOtherChain: tc.args.recipientOtherChain,
						ClosedBlock:         0,
						Status:              types.Open,
						CrossChain:          tc.args.crossChain,
						Direction:           tc.args.direction,
					}
				suite.Equal(expectedSwap, actualSwap)
			} else {
				suite.Error(err)
				// Check coins not moved
				suite.Equal(senderBalancePre, senderBalancePost)

				// Check incoming/outgoing asset supply not increased
				switch tc.args.direction {
				case types.Incoming:
					suite.Equal(assetSupplyPre.IncomingSupply, assetSupplyPost.IncomingSupply)
				case types.Outgoing:
					suite.Equal(assetSupplyPre.OutgoingSupply, assetSupplyPost.OutgoingSupply)
				default:
					suite.Fail("should not have invalid direction")
				}

				// Check if swap found in store
				_, found := suite.keeper.GetAtomicSwap(suite.ctx, expectedSwapID)
				if !tc.shouldBeFound {
					suite.False(found)
				} else {
					suite.True(found)
				}
			}
		})
	}
}

func (suite *AtomicSwapTestSuite) TestClaimAtomicSwap() {
	suite.SetupTest()
	currentTmTime := tmtime.Now()
	invalidRandomNumber, _ := types.GenerateSecureRandomNumber()
	type args struct {
		coins        sdk.Coins
		swapID       []byte
		randomNumber []byte
		direction    types.SwapDirection
	}
	testCases := []struct {
		name       string
		claimCtx   sdk.Context
		args       args
		expectPass bool
	}{
		{
			"normal incoming swap",
			suite.ctx,
			args{
				coins:        cs(c(BNB_DENOM, 50000)),
				swapID:       []byte{},
				randomNumber: []byte{},
				direction:    types.Incoming,
			},
			true,
		},
		{
			"normal incoming swap rate-limited",
			suite.ctx.WithBlockTime(currentTmTime.Add(time.Second * 10)),
			args{
				coins:        cs(c(OTHER_DENOM, 50000)),
				swapID:       []byte{},
				randomNumber: []byte{},
				direction:    types.Incoming,
			},
			true,
		},
		{
			"normal outgoing swap",
			suite.ctx,
			args{
				coins:        cs(c(BNB_DENOM, 50000)),
				swapID:       []byte{},
				randomNumber: []byte{},
				direction:    types.Outgoing,
			},
			true,
		},
		{
			"invalid random number",
			suite.ctx,
			args{
				coins:        cs(c(BNB_DENOM, 50000)),
				swapID:       []byte{},
				randomNumber: invalidRandomNumber[:],
				direction:    types.Incoming,
			},
			false,
		},
		{
			"wrong swap ID",
			suite.ctx,
			args{
				coins:        cs(c(BNB_DENOM, 50000)),
				swapID:       types.CalculateSwapID(suite.randomNumberHashes[3], suite.addrs[6], TestRecipientOtherChain),
				randomNumber: []byte{},
				direction:    types.Outgoing,
			},
			false,
		},
		{
			"past expiration",
			suite.ctx.WithBlockTime(currentTmTime.Add(time.Minute * 10)),
			args{
				coins:        cs(c(BNB_DENOM, 50000)),
				swapID:       []byte{},
				randomNumber: []byte{},
				direction:    types.Incoming,
			},
			false,
		},
	}

	for i, tc := range testCases {
		suite.GenerateSwapDetails()
		suite.Run(tc.name, func() {
			expectedRecipient := suite.addrs[5]
			sender := suite.deputy

			// Set sender to other and increment current asset supply for outgoing swap
			if tc.args.direction == types.Outgoing {
				sender = suite.addrs[6]
				expectedRecipient = suite.deputy
				err := suite.keeper.IncrementCurrentAssetSupply(suite.ctx, tc.args.coins[0])
				suite.Nil(err)
			}

			// Create atomic swap
			err := suite.keeper.CreateAtomicSwap(suite.ctx, suite.randomNumberHashes[i], suite.timestamps[i],
				types.DefaultSwapTimeSpan, sender, expectedRecipient, TestSenderOtherChain, TestRecipientOtherChain,
				tc.args.coins, true)
			suite.NoError(err)

			realSwapID := types.CalculateSwapID(suite.randomNumberHashes[i], sender, TestSenderOtherChain)

			// If args contains an invalid swap ID claim attempt will use it instead of the real swap ID
			var claimSwapID []byte
			if len(tc.args.swapID) == 0 {
				claimSwapID = realSwapID
			} else {
				claimSwapID = tc.args.swapID
			}

			// If args contains an invalid random number claim attempt will use it instead of the real random number
			var claimRandomNumber []byte
			if len(tc.args.randomNumber) == 0 {
				claimRandomNumber = suite.randomNumbers[i]
			} else {
				claimRandomNumber = tc.args.randomNumber
			}

			// Run the beginblocker before attempting claim
			bep3.BeginBlocker(tc.claimCtx, suite.keeper)

			// Load expected recipient's account prior to claim attempt
			bk := suite.bankKeeper
			expectedRecipientBalancePre := bk.GetBalance(tc.claimCtx, expectedRecipient, tc.args.coins[0].Denom).Amount
			// Load asset supplies prior to claim attempt
			assetSupplyPre, _ := suite.keeper.GetAssetSupply(tc.claimCtx, tc.args.coins[0].Denom)

			// Attempt to claim atomic swap
			err = suite.keeper.ClaimAtomicSwap(tc.claimCtx, expectedRecipient, claimSwapID, claimRandomNumber)

			// Load expected recipient's account after the claim attempt
			expectedRecipientBalancePost := bk.GetBalance(tc.claimCtx, expectedRecipient, tc.args.coins[0].Denom)
			// Load asset supplies after the claim attempt
			assetSupplyPost, _ := suite.keeper.GetAssetSupply(tc.claimCtx, tc.args.coins[0].Denom)

			if tc.expectPass {
				suite.NoError(err)

				// Check asset supply changes
				switch tc.args.direction {
				case types.Incoming:
					// Check coins moved
					suite.Equal(expectedRecipientBalancePre.Add(tc.args.coins[0].Amount), expectedRecipientBalancePost)
					// Check incoming supply decreased
					suite.True(assetSupplyPre.IncomingSupply.Amount.Sub(tc.args.coins[0].Amount).Equal(assetSupplyPost.IncomingSupply.Amount))
					// Check current supply increased
					suite.Equal(assetSupplyPre.CurrentSupply.Add(tc.args.coins[0]), assetSupplyPost.CurrentSupply)
					// Check outgoing supply not changed
					suite.Equal(assetSupplyPre.OutgoingSupply, assetSupplyPost.OutgoingSupply)
				case types.Outgoing:
					// Check incoming supply not changed
					suite.Equal(assetSupplyPre.IncomingSupply, assetSupplyPost.IncomingSupply)
					// Check current supply decreased
					suite.Equal(assetSupplyPre.CurrentSupply.Sub(tc.args.coins[0]), assetSupplyPost.CurrentSupply)
					// Check outgoing supply decreased
					suite.True(assetSupplyPre.OutgoingSupply.Sub(tc.args.coins[0]).IsEqual(assetSupplyPost.OutgoingSupply))
				default:
					suite.Fail("should not have invalid direction")
				}
			} else {
				suite.Error(err)
				// Check coins not moved
				suite.Equal(expectedRecipientBalancePre, expectedRecipientBalancePost)

				// Check asset supply has not changed
				switch tc.args.direction {
				case types.Incoming:
					suite.Equal(assetSupplyPre.IncomingSupply, assetSupplyPost.IncomingSupply)
					suite.Equal(assetSupplyPre.CurrentSupply, assetSupplyPost.CurrentSupply)
					suite.Equal(assetSupplyPre.OutgoingSupply, assetSupplyPost.OutgoingSupply)
				case types.Outgoing:
					suite.Equal(assetSupplyPre.IncomingSupply, assetSupplyPost.IncomingSupply)
					suite.Equal(assetSupplyPre.CurrentSupply, assetSupplyPost.CurrentSupply)
					suite.Equal(assetSupplyPre.OutgoingSupply, assetSupplyPost.OutgoingSupply)
				default:
					suite.Fail("should not have invalid direction")
				}
			}
		})
	}
}

// getContextPlusSec returns a context forward or backward in time and block
// index. Assuming 1 second finality.
func (suite *AtomicSwapTestSuite) getContextPlusSec(plusSeconds int64) sdk.Context {
	offset := int64(plusSeconds)
	ctx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Duration(offset) * time.Second))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + offset)

	return ctx
}

func (suite *AtomicSwapTestSuite) TestRefundAtomicSwap() {
	suite.SetupTest()

	type args struct {
		swapID    []byte
		direction types.SwapDirection
	}
	testCases := []struct {
		name       string
		refundCtx  sdk.Context
		args       args
		expectPass bool
	}{
		{
			"expired incoming swap",
			suite.getContextPlusSec(bep3.DefaultSwapTimeSpan + 1),
			args{
				swapID:    []byte{},
				direction: types.Incoming,
			},
			true,
		},
		{
			"normal outgoing swap",
			suite.getContextPlusSec(bep3.DefaultSwapTimeSpan + 1),
			args{
				swapID:    []byte{},
				direction: types.Outgoing,
			},
			true,
		},
		{
			"before expiration",
			suite.ctx,
			args{
				swapID:    []byte{},
				direction: types.Incoming,
			},
			false,
		},
		{
			"wrong swapID",
			suite.ctx,
			args{
				swapID:    types.CalculateSwapID(suite.randomNumberHashes[6], suite.addrs[1], TestRecipientOtherChain),
				direction: types.Incoming,
			},
			false,
		},
	}

	for i, tc := range testCases {
		suite.GenerateSwapDetails()
		suite.Run(tc.name, func() {
			// Create atomic swap
			expectedRefundAmount := cs(c(BNB_DENOM, 50000))
			sender := suite.deputy
			expectedRecipient := suite.addrs[9]

			// Set sender to other and increment current asset supply for outgoing swap
			if tc.args.direction == types.Outgoing {
				sender = suite.addrs[6]
				expectedRecipient = suite.deputy
				err := suite.keeper.IncrementCurrentAssetSupply(suite.ctx, expectedRefundAmount[0])
				suite.Nil(err)
			}

			err := suite.keeper.CreateAtomicSwap(suite.ctx, suite.randomNumberHashes[i], suite.timestamps[i],
				types.DefaultSwapTimeSpan, sender, expectedRecipient, TestSenderOtherChain, TestRecipientOtherChain,
				expectedRefundAmount, true)
			suite.NoError(err)

			realSwapID := types.CalculateSwapID(suite.randomNumberHashes[i], sender, TestSenderOtherChain)

			// If args contains an invalid swap ID refund attempt will use it instead of the real swap ID
			var refundSwapID []byte
			if len(tc.args.swapID) == 0 {
				refundSwapID = realSwapID
			} else {
				refundSwapID = tc.args.swapID
			}

			// Run the beginblocker before attempting refund
			bep3.BeginBlocker(tc.refundCtx, suite.keeper)

			// Load sender's account prior to swap refund
			bk := suite.bankKeeper
			originalSenderBalancePre := bk.GetBalance(tc.refundCtx, sender, expectedRefundAmount[0].Denom).Amount
			// Load asset supply prior to swap refund
			assetSupplyPre, _ := suite.keeper.GetAssetSupply(tc.refundCtx, expectedRefundAmount[0].Denom)

			// Attempt to refund atomic swap
			err = suite.keeper.RefundAtomicSwap(tc.refundCtx, sender, refundSwapID)

			// Load sender's account after refund
			originalSenderBalancePost := bk.GetBalance(tc.refundCtx, sender, expectedRefundAmount[0].Denom)
			// Load asset supply after to swap refund
			assetSupplyPost, _ := suite.keeper.GetAssetSupply(tc.refundCtx, expectedRefundAmount[0].Denom)

			if tc.expectPass {
				suite.NoError(err)

				// Check asset supply changes
				switch tc.args.direction {
				case types.Incoming:
					// Check incoming supply decreased
					suite.True(assetSupplyPre.IncomingSupply.Sub(expectedRefundAmount[0]).IsEqual(assetSupplyPost.IncomingSupply))
					// Check current, outgoing supply not changed
					suite.Equal(assetSupplyPre.CurrentSupply, assetSupplyPost.CurrentSupply)
					suite.Equal(assetSupplyPre.OutgoingSupply, assetSupplyPost.OutgoingSupply)
				case types.Outgoing:
					// Check coins moved
					suite.Equal(originalSenderBalancePre.Add(expectedRefundAmount[0].Amount), originalSenderBalancePost)
					// Check incoming, current supply not changed
					suite.Equal(assetSupplyPre.IncomingSupply, assetSupplyPost.IncomingSupply)
					suite.Equal(assetSupplyPre.CurrentSupply, assetSupplyPost.CurrentSupply)
					// Check outgoing supply decreased
					suite.True(assetSupplyPre.OutgoingSupply.Sub(expectedRefundAmount[0]).IsEqual(assetSupplyPost.OutgoingSupply))
				default:
					suite.Fail("should not have invalid direction")
				}
			} else {
				suite.Error(err)
				// Check coins not moved
				suite.Equal(originalSenderBalancePre, originalSenderBalancePost)

				// Check asset supply has not changed
				switch tc.args.direction {
				case types.Incoming:
					suite.Equal(assetSupplyPre.IncomingSupply, assetSupplyPost.IncomingSupply)
					suite.Equal(assetSupplyPre.CurrentSupply, assetSupplyPost.CurrentSupply)
					suite.Equal(assetSupplyPre.OutgoingSupply, assetSupplyPost.OutgoingSupply)
				case types.Outgoing:
					suite.Equal(assetSupplyPre.IncomingSupply, assetSupplyPost.IncomingSupply)
					suite.Equal(assetSupplyPre.CurrentSupply, assetSupplyPost.CurrentSupply)
					suite.Equal(assetSupplyPre.OutgoingSupply, assetSupplyPost.OutgoingSupply)
				default:
					suite.Fail("should not have invalid direction")
				}
			}
		})
	}
}

func TestAtomicSwapTestSuite(t *testing.T) {
	suite.Run(t, new(AtomicSwapTestSuite))
}
