package rest

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// REST Variable names
// nolint
const (
	RestExpiration = "expiration"
	RestInvolve    = "involve"
	RestStatus     = "status"
	RestDirection  = "direction"
)

// RegisterRoutes registers bep3-related REST handlers to a router
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

// PostCreateSwapReq defines the properties of a swap create request's body
type PostCreateSwapReq struct {
	BaseReq             rest.BaseReq     `json:"base_req" yaml:"base_req"`
	From                string           `json:"from" yaml:"from"`
	To                  string           `json:"to" yaml:"to"`
	RecipientOtherChain string           `json:"recipient_other_chain" yaml:"recipient_other_chain"`
	SenderOtherChain    string           `json:"sender_other_chain" yaml:"sender_other_chain"`
	RandomNumberHash    tmbytes.HexBytes `json:"random_number_hash" yaml:"random_number_hash"`
	Timestamp           int64            `json:"timestamp" yaml:"timestamp"`
	Amount              sdk.Coins        `json:"amount" yaml:"amount"`
	TimeSpan            int64            `json:"time_span" yaml:"time_span"`
	CrossChain          bool             `json:"cross_chain" yaml:"cross_chain"`
}

// PostClaimSwapReq defines the properties of a swap claim request's body
type PostClaimSwapReq struct {
	BaseReq      rest.BaseReq     `json:"base_req" yaml:"base_req"`
	From         sdk.AccAddress   `json:"from" yaml:"from"`
	SwapID       tmbytes.HexBytes `json:"swap_id" yaml:"swap_id"`
	RandomNumber tmbytes.HexBytes `json:"random_number" yaml:"random_number"`
}

// PostRefundSwapReq defines the properties of swap refund request's body
type PostRefundSwapReq struct {
	BaseReq rest.BaseReq     `json:"base_req" yaml:"base_req"`
	From    sdk.AccAddress   `json:"from" yaml:"from"`
	SwapID  tmbytes.HexBytes `json:"swap_id" yaml:"swap_id"`
}
