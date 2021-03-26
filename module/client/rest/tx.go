package rest

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/e-money/bep3/module/types"
	"github.com/gorilla/mux"
)

func registerTxRoutes(cliCtx client.Context, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/txs", types.ModuleName), BroadcastTxRequest(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/swap/create", types.ModuleName), postCreateHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/swap/claim", types.ModuleName), postClaimHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/swap/refund", types.ModuleName), postRefundHandlerFn(cliCtx)).Methods("POST")
}

// BroadcastReq defines a tx broadcasting request.
type BroadcastReq struct {
	Tx   legacytx.StdTx `json:"tx" yaml:"tx"`
	Mode string         `json:"mode" yaml:"mode"`
}

// BroadcastTxRequest implements a tx broadcasting handler that is responsible
// for broadcasting a valid and signed tx to a tendermint RCP port. The tx can
// be broadcasted via a sync|async|block mechanism via the tendermint RPC Api.
func BroadcastTxRequest(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req BroadcastReq

		body, err := ioutil.ReadAll(r.Body)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		err = clientCtx.LegacyAmino.UnmarshalJSON(body, &req)
		if err != nil {
			err = fmt.Errorf("Unmarshalling error of req: %w", err)
			if rest.CheckBadRequestError(w, err) {
				return
			}
		}

		txBytes, err := tx.ConvertAndEncodeStdTx(clientCtx.TxConfig, req.Tx)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithBroadcastMode(req.Mode)

		res, err := clientCtx.BroadcastTx(txBytes)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		rest.PostProcessResponseBare(w, clientCtx, res)
	}
}

func postCreateHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Decode PUT request body
		var req PostCreateSwapReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		from, err := sdk.AccAddressFromBech32(req.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, fmt.Sprintf("invalid from address: %s, bech32 format error:%s", req.From, err))

			return
		}

		to, err := sdk.AccAddressFromBech32(req.To)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, fmt.Sprintf("invalid to address: %s, bech32 format error:%s", req.To, err))

			return
		}

		// Create and return msg
		msg := types.NewMsgCreateAtomicSwap(
			from,
			to,
			req.RecipientOtherChain,
			req.SenderOtherChain,
			req.RandomNumberHash,
			req.Timestamp,
			req.Amount,
			req.TimeSpan,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

func postClaimHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode PUT request body
		var req PostClaimSwapReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// Create and return msg
		msg := types.NewMsgClaimAtomicSwap(
			req.From,
			req.SwapID,
			req.RandomNumber,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

func postRefundHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode PUT request body
		var req PostRefundSwapReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		senderAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Create and return msg
		msg := types.NewMsgRefundAtomicSwap(
			senderAddr,
			req.SwapID,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}
