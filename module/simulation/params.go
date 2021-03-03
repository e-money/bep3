package simulation

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/e-money/bep3/module/types"
)

const (
	keyAssetParams = "AssetParams"
)

// ParamChanges defines the parameters that can be modified by param change proposals
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyAssetParams,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenSupportedAssets(r))
			},
		),
	}
}
