package bep3

import (
	"encoding/json"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	sdksim "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/e-money/bep3/module/client/cli"
	"github.com/e-money/bep3/module/client/rest"
	"github.com/e-money/bep3/module/keeper"
	"github.com/e-money/bep3/module/simulation"
	bep3types "github.com/e-money/bep3/module/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
	// TODO Simulation
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the bep3 module.
type AppModuleBasic struct {
	cdc *codec.LegacyAmino
}

// Name returns the bep3 module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterLegacyAminoCodec registers the bep3 module's types for Amino.
func (ab AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	bep3types.RegisterLegacyAminoCodec(cdc)
	ab.cdc = cdc
}

func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	bep3types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the bep3
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the bep3 module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage) error {
	var gs GenesisState
	err := cdc.UnmarshalJSON(bz, &gs)
	if err != nil {
		return err
	}
	return gs.Validate()
}

// RegisterRESTRoutes registers the REST routes for the bep3 module.
func (AppModuleBasic) RegisterRESTRoutes(ctx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(ctx, rtr)
}

// TODO GRPC
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
}

// GetTxCmd returns the root tx command for the bep3 module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns no root query command for the bep3 module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

//____________________________________________________________________________

// AppModule implements the sdk.AppModule interface.
type AppModule struct {
	AppModuleBasic

	keeper        Keeper
	accountKeeper bep3types.AccountKeeper
	bankKeeper    bep3types.BankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper Keeper, accountKeeper bep3types.AccountKeeper, bankKeeper bep3types.BankKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
	}
}

// Name returns the bep3 module's name.
func (AppModule) Name() string {
	return ModuleName
}

// RegisterInvariants registers the bep3 module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the bep3 module.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(bep3types.RouterKey, NewHandler(am.keeper))
}

// NewHandler returns an sdk.Handler for the bep3 module.
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// QuerierRoute returns the bep3 module's querier route name.
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return keeper.NewQuerier(am.keeper)
}

// NewQuerierHandler returns the bep3 module sdk.Querier.
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// InitGenesis performs genesis initialization for the bep3 module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, am.accountKeeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the bep3
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
}

// BeginBlock returns the begin blocker for the bep3 module.
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	BeginBlocker(ctx, am.keeper)
}

// EndBlock returns the end blocker for the bep3 module. It returns no validator updates.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

//____________________________________________________________________________

// GenerateGenesisState creates a randomized GenState of the bep3 module
func (AppModuleBasic) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents doesn't return any content functions for governance proposals.
func (AppModuleBasic) ProposalContents(_ module.SimulationState) []sdksim.WeightedProposalContent {
	return nil
}

// RandomizedParams returns nil because bep3 has no params.
func (AppModuleBasic) RandomizedParams(r *rand.Rand) []sdksim.ParamChange {
	return simulation.ParamChanges(r)
}

// RegisterStoreDecoder registers a decoder for bep3 module's types
func (ab AppModuleBasic) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[StoreKey] = simulation.NewDecodeStore(ab.cdc)
}

// WeightedOperations returns the all the bep3 module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []sdksim.WeightedOperation {
	return simulation.WeightedOperations(simState.AppParams, simState.Cdc, am.accountKeeper, am.bankKeeper, am.keeper)
}
