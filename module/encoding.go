package bep3

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Marshaler         codec.Marshaler
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
	marshaler := getAminoMarshaller(cdc)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          legacytx.StdTxConfig{Cdc: cdc},
		Amino:             cdc,
	}
}

func getAminoMarshaller(cdc *codec.LegacyAmino) *codec.AminoCodec {
	marshaler := codec.NewAminoCodec(cdc)
	return marshaler
}

// MakeProtoEncodingConfig creates an EncodingConfig for a protobuf based configuration.
func MakeProtoEncodingConfig() EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := getProtoMarshaller(interfaceRegistry)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             cdc,
	}
}

func getProtoMarshaller(interfaceRegistry types.InterfaceRegistry) *codec.ProtoCodec {
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	return marshaler
}
