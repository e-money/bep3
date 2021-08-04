package testapp

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	bep3 "github.com/e-money/bep3/module"
	bep3types "github.com/e-money/bep3/module/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func CreateTestComponents(t *testing.T) (
	sdk.Context,
	codec.JSONCodec,
	bep3.Keeper,
	bep3types.AccountKeeper,
	bep3types.BankKeeper,
	bep3.AppModule) {
	encoding := bep3.MakeProtoEncodingConfig()

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	keys := sdk.NewKVStoreKeys(bep3.StoreKey, authtypes.StoreKey, banktypes.StoreKey, paramstypes.StoreKey)
	for _, k := range keys {
		ms.MountStoreWithDB(k, sdk.StoreTypeIAVL, db)
	}

	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	for _, k := range tkeys {
		ms.MountStoreWithDB(k, sdk.StoreTypeTransient, db)
	}

	err := ms.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(ms, tmproto.Header{
		ChainID: "test-chain",
	}, true, log.NewNopLogger())
	ctx = ctx.WithBlockTime(time.Now())

	mAccPerms := map[string][]string{
		bep3.ModuleName: {authtypes.Minter, authtypes.Burner},
	}

	paramsKeeper := paramskeeper.NewKeeper(encoding.Marshaller, encoding.Amino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])
	var (
		authSubspace = paramsKeeper.Subspace(authtypes.ModuleName)
		bankSubspace = paramsKeeper.Subspace(banktypes.ModuleName)
		bep3Subspace = paramsKeeper.Subspace(bep3.DefaultParamspace)
	)

	var (
		accountKeeper = authkeeper.NewAccountKeeper(encoding.Marshaller, keys[authtypes.StoreKey], authSubspace, authtypes.ProtoBaseAccount, mAccPerms)
		bankKeeper    = bankkeeper.NewBaseKeeper(encoding.Marshaller, keys[banktypes.ModuleName], accountKeeper, bankSubspace, make(map[string]bool))
		bep3Keeper    = bep3.NewKeeper(encoding.Marshaller, keys[bep3.StoreKey], bankKeeper, accountKeeper, bep3Subspace, make(map[string]bool))
	)

	err = bankKeeper.MintCoins(ctx, bep3.ModuleName, sdk.NewCoins())
	require.NoError(t, err)

	return ctx,
		codec.NewProtoCodec(nil),
		bep3Keeper,
		accountKeeper,
		bankKeeper,
		bep3.NewAppModule(bep3Keeper, accountKeeper, bankKeeper)
}
