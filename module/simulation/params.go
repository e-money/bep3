package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const (
	keyAssetParams = "AssetParams"
)

// ParamChanges defines the parameters that can be modified by param change proposals
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	return []simulation.ParamChange{
		// TODO
		//simulation.NewSimParamChange(types.ModuleName, keyAssetParams,
		//	func(r *rand.Rand) string {
		//		return fmt.Sprintf("\"%s\"", GenSupportedAssets(r))
		//	},
		//),
	}
}
