package bep3

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Marshaller        codec.Marshaler
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// MakeEncodingConfig creates an EncodingConfig for a selected codec.
func MakeEncodingConfig() EncodingConfig {
	// default to Amino
	return MakeAminoEncodingConfig()
}

// MakeAminoEncodingConfig creates an EncodingConfig for an amino based configuration.
func MakeAminoEncodingConfig() EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaller := getAminoMarshaller(cdc)

	cfg := EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaller:        marshaller,
		TxConfig:          legacytx.StdTxConfig{Cdc: cdc},
		Amino:             cdc,
	}

	std.RegisterLegacyAminoCodec(cfg.Amino)
	std.RegisterInterfaces(cfg.InterfaceRegistry)
	ModuleBasics := module.NewBasicManager(
		bank.AppModuleBasic{},
		auth.AppModuleBasic{},
	)

	ModuleBasics.RegisterLegacyAminoCodec(cfg.Amino)
	ModuleBasics.RegisterInterfaces(cfg.InterfaceRegistry)

	return cfg
}

func getAminoMarshaller(cdc *codec.LegacyAmino) *codec.AminoCodec {
	marshaller := codec.NewAminoCodec(cdc)
	return marshaller
}

// MakeProtoEncodingConfig creates an EncodingConfig for a protobuf based configuration.
func MakeProtoEncodingConfig() EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaller := getProtoMarshaller(interfaceRegistry)

	cfg := EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaller:        marshaller,
		TxConfig:          tx.NewTxConfig(marshaller, tx.DefaultSignModes),
		Amino:             cdc,
	}

	std.RegisterLegacyAminoCodec(cfg.Amino)
	std.RegisterInterfaces(cfg.InterfaceRegistry)
	ModuleBasics := module.NewBasicManager(
		bank.AppModuleBasic{},
		auth.AppModuleBasic{},
	)

	ModuleBasics.RegisterLegacyAminoCodec(cfg.Amino)
	ModuleBasics.RegisterInterfaces(cfg.InterfaceRegistry)

	return cfg
}

func getProtoMarshaller(interfaceRegistry types.InterfaceRegistry) *codec.ProtoCodec {
	marshaller := codec.NewProtoCodec(interfaceRegistry)
	return marshaller
}
