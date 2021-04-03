<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [bep3/swap.proto](#bep3/swap.proto)
    - [AtomicSwap](#bep3.AtomicSwap)
    - [AugmentedAtomicSwap](#bep3.AugmentedAtomicSwap)
    - [AugmentedAtomicSwaps](#bep3.AugmentedAtomicSwaps)
    - [MsgClaimAtomicSwap](#bep3.MsgClaimAtomicSwap)
    - [MsgClaimAtomicSwapResponse](#bep3.MsgClaimAtomicSwapResponse)
    - [MsgCreateAtomicSwap](#bep3.MsgCreateAtomicSwap)
    - [MsgCreateAtomicSwapResponse](#bep3.MsgCreateAtomicSwapResponse)
    - [MsgRefundAtomicSwap](#bep3.MsgRefundAtomicSwap)
    - [MsgRefundAtomicSwapResponse](#bep3.MsgRefundAtomicSwapResponse)
    - [PrevBlockTime](#bep3.PrevBlockTime)
  
- [bep3/genesis.proto](#bep3/genesis.proto)
    - [AssetParam](#bep3.AssetParam)
    - [AssetSupplies](#bep3.AssetSupplies)
    - [AssetSupply](#bep3.AssetSupply)
    - [GenesisState](#bep3.GenesisState)
    - [Params](#bep3.Params)
    - [SupplyLimit](#bep3.SupplyLimit)
  
- [bep3/query.proto](#bep3/query.proto)
    - [QueryAssetSupplies](#bep3.QueryAssetSupplies)
    - [QueryAssetSuppliesRequest](#bep3.QueryAssetSuppliesRequest)
    - [QueryAssetSuppliesResponse](#bep3.QueryAssetSuppliesResponse)
    - [QueryAssetSupply](#bep3.QueryAssetSupply)
    - [QueryAssetSupplyRequest](#bep3.QueryAssetSupplyRequest)
    - [QueryAssetSupplyResponse](#bep3.QueryAssetSupplyResponse)
    - [QueryAtomicSwapByID](#bep3.QueryAtomicSwapByID)
    - [QueryAtomicSwaps](#bep3.QueryAtomicSwaps)
    - [QuerySwapRequest](#bep3.QuerySwapRequest)
    - [QuerySwapResponse](#bep3.QuerySwapResponse)
    - [QuerySwapsRequest](#bep3.QuerySwapsRequest)
    - [QuerySwapsResponse](#bep3.QuerySwapsResponse)
  
    - [Query](#bep3.Query)
  
- [bep3/tx.proto](#bep3/tx.proto)
    - [Msg](#bep3.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="bep3/swap.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bep3/swap.proto



<a name="bep3.AtomicSwap"></a>

### AtomicSwap
AtomicSwap contains the information for an atomic swap


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `random_number_hash` | [bytes](#bytes) |  |  |
| `expire_timestamp` | [int64](#int64) |  |  |
| `timestamp` | [int64](#int64) |  |  |
| `sender` | [string](#string) |  |  |
| `recipient` | [string](#string) |  |  |
| `sender_other_chain` | [string](#string) |  |  |
| `recipient_other_chain` | [string](#string) |  |  |
| `closed_block` | [int64](#int64) |  |  |
| `status` | [uint32](#uint32) |  |  |
| `cross_chain` | [bool](#bool) |  |  |
| `direction` | [uint32](#uint32) |  |  |






<a name="bep3.AugmentedAtomicSwap"></a>

### AugmentedAtomicSwap
AtomicSwap with an ID


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `random_number_hash` | [bytes](#bytes) |  |  |
| `expire_timestamp` | [int64](#int64) |  |  |
| `timestamp` | [int64](#int64) |  |  |
| `sender` | [string](#string) |  |  |
| `recipient` | [string](#string) |  |  |
| `sender_other_chain` | [string](#string) |  |  |
| `recipient_other_chain` | [string](#string) |  |  |
| `closed_block` | [int64](#int64) |  |  |
| `status` | [uint32](#uint32) |  |  |
| `cross_chain` | [bool](#bool) |  |  |
| `direction` | [uint32](#uint32) |  |  |






<a name="bep3.AugmentedAtomicSwaps"></a>

### AugmentedAtomicSwaps
Slice of Augmented Atomic Swaps


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `augmented_atomic_swaps` | [AugmentedAtomicSwap](#bep3.AugmentedAtomicSwap) | repeated |  |






<a name="bep3.MsgClaimAtomicSwap"></a>

### MsgClaimAtomicSwap
MsgClaimAtomicSwap defines a AtomicSwap claim


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from` | [string](#string) |  |  |
| `swap_id` | [bytes](#bytes) |  |  |
| `random_number` | [bytes](#bytes) |  |  |






<a name="bep3.MsgClaimAtomicSwapResponse"></a>

### MsgClaimAtomicSwapResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `random_number_hash` | [string](#string) |  |  |
| `timestamp` | [int64](#int64) |  |  |






<a name="bep3.MsgCreateAtomicSwap"></a>

### MsgCreateAtomicSwap
MsgCreateAtomicSwap contains an AtomicSwap struct


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from` | [string](#string) |  |  |
| `to` | [string](#string) |  |  |
| `recipient_other_chain` | [string](#string) |  |  |
| `sender_other_chain` | [string](#string) |  |  |
| `random_number_hash` | [bytes](#bytes) |  |  |
| `timestamp` | [int64](#int64) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `time_span` | [int64](#int64) |  |  |






<a name="bep3.MsgCreateAtomicSwapResponse"></a>

### MsgCreateAtomicSwapResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `random_number_hash` | [string](#string) |  |  |
| `swap_id` | [string](#string) |  |  |






<a name="bep3.MsgRefundAtomicSwap"></a>

### MsgRefundAtomicSwap
MsgRefundAtomicSwap defines a refund msg


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from` | [string](#string) |  |  |
| `swap_id` | [bytes](#bytes) |  |  |






<a name="bep3.MsgRefundAtomicSwapResponse"></a>

### MsgRefundAtomicSwapResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `random_number_hash` | [string](#string) |  |  |
| `timestamp` | [int64](#int64) |  |  |






<a name="bep3.PrevBlockTime"></a>

### PrevBlockTime
Proto type required for serializing the previous block time to manage supply
expirations.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `val` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="bep3/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bep3/genesis.proto



<a name="bep3.AssetParam"></a>

### AssetParam
AssetParam parameters that must be specified for each bep3 asset


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | name of the asset |
| `coin_id` | [int64](#int64) |  | SLIP-0044 registered coin type - see https://github.com/satoshilabs/slips/blob/master/slip-0044.md |
| `supply_limit` | [SupplyLimit](#bep3.SupplyLimit) |  | asset supply limit |
| `active` | [bool](#bool) |  | denotes if asset is available or paused |
| `deputy_address` | [string](#string) |  | the address of the relayer process |
| `fixed_fee` | [string](#string) |  | It should match the deputy config chain values. The fixed fee charged by the relayer process for outgoing swaps |
| `min_swap_amount` | [string](#string) |  | Minimum swap amount |
| `max_swap_amount` | [string](#string) |  | Maximum swap amount |
| `swap_time` | [int64](#int64) |  | Unix seconds of swap creation block timestamp Original	SwapTimestamp int64 `json:"swap_time" yaml:"swap_time"` |
| `time_span` | [int64](#int64) |  | seconds span before time expiration Original SwapTimeSpan int64 `json:"time_span" yaml:"time_span"` |






<a name="bep3.AssetSupplies"></a>

### AssetSupplies
slice of AssetSupply


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_supplies` | [AssetSupply](#bep3.AssetSupply) | repeated |  |






<a name="bep3.AssetSupply"></a>

### AssetSupply
AssetSupply contains information about an asset's supply


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `incoming_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `outgoing_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `current_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `time_limited_current_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `time_elapsed` | [int64](#int64) |  | the time.duration int64 units of times elapsed |






<a name="bep3.GenesisState"></a>

### GenesisState
type GenesisState struct {


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#bep3.Params) |  |  |
| `atomic_swaps` | [AtomicSwap](#bep3.AtomicSwap) | repeated |  |
| `supplies` | [AssetSupplies](#bep3.AssetSupplies) |  |  |
| `previous_block_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="bep3.Params"></a>

### Params
Params governance parameters for bep3 module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_params` | [AssetParam](#bep3.AssetParam) | repeated |  |






<a name="bep3.SupplyLimit"></a>

### SupplyLimit
SupplyLimit parameters that control the absolute and time-based limits for an assets's supply


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `limit` | [string](#string) |  | the absolute supply limit for an asset |
| `time_limited` | [bool](#bool) |  | boolean for whether the supply is limited by time |
| `time_period` | [int64](#int64) |  | the time.duration int64 units for which the supply time limit applies |
| `time_based_limit` | [string](#string) |  | the supply limit for an asset for each time period |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="bep3/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bep3/query.proto



<a name="bep3.QueryAssetSupplies"></a>

### QueryAssetSupplies
QueryAssetSupplies contains the params for an AssetSupplies query
type QueryAssetSupplies struct {
Page  int `json:"page" yaml:"page"`
Limit int `json:"limit" yaml:"limit"`
}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `page` | [int64](#int64) |  |  |
| `limit` | [int64](#int64) |  |  |






<a name="bep3.QueryAssetSuppliesRequest"></a>

### QueryAssetSuppliesRequest
gRPC assets req






<a name="bep3.QueryAssetSuppliesResponse"></a>

### QueryAssetSuppliesResponse
gRPC asset supplies response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `supplies` | [AssetSupplies](#bep3.AssetSupplies) |  |  |






<a name="bep3.QueryAssetSupply"></a>

### QueryAssetSupply
QueryAssetSupply contains the params for query 'custom/bep3/supply'


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="bep3.QueryAssetSupplyRequest"></a>

### QueryAssetSupplyRequest
gRPC asset req


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="bep3.QueryAssetSupplyResponse"></a>

### QueryAssetSupplyResponse
gRPC asset supply response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `supply` | [AssetSupply](#bep3.AssetSupply) |  |  |






<a name="bep3.QueryAtomicSwapByID"></a>

### QueryAtomicSwapByID
QueryAtomicSwapByID contains the params for query 'custom/bep3/swap'


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `swap_id` | [bytes](#bytes) |  |  |






<a name="bep3.QueryAtomicSwaps"></a>

### QueryAtomicSwaps
QueryAtomicSwaps contains the params for an AtomicSwaps query


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `page` | [int64](#int64) |  |  |
| `limit` | [int64](#int64) |  |  |
| `involve` | [string](#string) |  |  |
| `expiration` | [int64](#int64) |  |  |
| `status` | [uint32](#uint32) |  |  |
| `direction` | [uint32](#uint32) |  |  |






<a name="bep3.QuerySwapRequest"></a>

### QuerySwapRequest
gRPC swap req


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `swap_id` | [bytes](#bytes) |  |  |






<a name="bep3.QuerySwapResponse"></a>

### QuerySwapResponse
gRPC swap response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `swap` | [AtomicSwap](#bep3.AtomicSwap) |  |  |






<a name="bep3.QuerySwapsRequest"></a>

### QuerySwapsRequest
gRPC swaps req


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [QueryAtomicSwaps](#bep3.QueryAtomicSwaps) |  |  |






<a name="bep3.QuerySwapsResponse"></a>

### QuerySwapsResponse
gRPC swap response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `swaps` | [AugmentedAtomicSwaps](#bep3.AugmentedAtomicSwaps) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="bep3.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `AssetSupply` | [QueryAssetSupplyRequest](#bep3.QueryAssetSupplyRequest) | [QueryAssetSupplyResponse](#bep3.QueryAssetSupplyResponse) |  | GET|/e-money/bep3/supply|
| `AssetSupplies` | [QueryAssetSuppliesRequest](#bep3.QueryAssetSuppliesRequest) | [QueryAssetSuppliesResponse](#bep3.QueryAssetSuppliesResponse) |  | GET|/e-money/bep3/supplies|
| `Swap` | [QuerySwapRequest](#bep3.QuerySwapRequest) | [QuerySwapResponse](#bep3.QuerySwapResponse) |  | GET|/e-money/bep3/swap|
| `Swaps` | [QuerySwapsRequest](#bep3.QuerySwapsRequest) | [QuerySwapsResponse](#bep3.QuerySwapsResponse) |  | GET|/e-money/bep3/swap|

 <!-- end services -->



<a name="bep3/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bep3/tx.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="bep3.Msg"></a>

### Msg


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateAtomicSwap` | [MsgCreateAtomicSwap](#bep3.MsgCreateAtomicSwap) | [MsgCreateAtomicSwapResponse](#bep3.MsgCreateAtomicSwapResponse) |  | |
| `ClaimAtomicSwap` | [MsgClaimAtomicSwap](#bep3.MsgClaimAtomicSwap) | [MsgClaimAtomicSwapResponse](#bep3.MsgClaimAtomicSwapResponse) |  | |
| `RefundAtomicSwap` | [MsgRefundAtomicSwap](#bep3.MsgRefundAtomicSwap) | [MsgRefundAtomicSwapResponse](#bep3.MsgRefundAtomicSwapResponse) |  | |

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

