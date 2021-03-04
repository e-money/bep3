package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// NewAtomicSwap returns a new AtomicSwap
func NewAtomicSwap(amount sdk.Coins, randomNumberHash tmbytes.HexBytes, expireTimestamp, timestamp int64,
	sender, recipient sdk.AccAddress, senderOtherChain, recipientOtherChain string, closedBlock int64,
	status SwapStatus, crossChain bool, direction SwapDirection) AtomicSwap {
	return AtomicSwap{
		Amount:              amount,
		RandomNumberHash:    randomNumberHash,
		ExpireTimestamp:     expireTimestamp,
		Timestamp:           timestamp,
		Sender:              sender.String(),
		Recipient:           recipient.String(),
		SenderOtherChain:    senderOtherChain,
		RecipientOtherChain: recipientOtherChain,
		ClosedBlock:         closedBlock,
		Status:              status,
		CrossChain:          crossChain,
		Direction:           direction,
	}
}

// GetSwapID calculates the ID of an atomic swap
func (a AtomicSwap) GetSwapID() tmbytes.HexBytes {
	sender, err := sdk.AccAddressFromBech32(a.Sender)
	if err != nil {
		return nil
	}

	return CalculateSwapID(a.RandomNumberHash, sender, a.SenderOtherChain)
}

// GetCoins returns the swap's amount as sdk.Coins
func (a AtomicSwap) GetCoins() sdk.Coins {
	return sdk.NewCoins(a.Amount...)
}

// Validate performs a basic validation of an atomic swap fields.
func (a AtomicSwap) Validate() error {
	if !a.Amount.IsValid() {
		return fmt.Errorf("invalid amount: %s", a.Amount)
	}
	if !a.Amount.IsAllPositive() {
		return fmt.Errorf("the swapped out coin must be positive: %s", a.Amount)
	}
	if len(a.RandomNumberHash) != RandomNumberHashLength {
		return fmt.Errorf("the length of random number hash should be %d", RandomNumberHashLength)
	}
	if a.ExpireTimestamp == 0 {
		return errors.New("expire timestamp cannot be 0")
	}
	if a.Timestamp == 0 {
		return errors.New("timestamp cannot be 0")
	}
	if len(a.Sender) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender cannot be empty")
	}
	if len(a.Recipient) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "recipient cannot be empty")
	}
	senderAcc, err := sdk.AccAddressFromBech32(a.Sender)
	if err != nil {
		return fmt.Errorf("expected Bech32 address for sender %s", a.Sender)
	}
	if len(senderAcc.Bytes()) != AddrByteCount {
		return fmt.Errorf("the expected address length is %d, actual length is %d", AddrByteCount, len(a.Sender))
	}
	recAcc, err := sdk.AccAddressFromBech32(a.Recipient)
	if err != nil {
		return fmt.Errorf("expected Bech32 address for recipient %s", a.Recipient)
	}
	if len(recAcc.Bytes()) != AddrByteCount {
		return fmt.Errorf("the expected address length is %d, actual length is %d", AddrByteCount, len(a.Recipient))
	}
	// NOTE: These adresses may not have a bech32 prefix.
	if strings.TrimSpace(a.SenderOtherChain) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender other chain cannot be blank")
	}
	if strings.TrimSpace(a.RecipientOtherChain) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "recipient other chain cannot be blank")
	}
	if a.Status == Completed && a.ClosedBlock == 0 {
		return errors.New("closed block cannot be 0")
	}
	if a.Status == NULL || a.Status > 3 {
		return errors.New("invalid swap status")
	}
	if a.Direction == INVALID || a.Direction > 2 {
		return errors.New("invalid swap direction")
	}
	return nil
}

// String implements stringer
func (a AtomicSwap) String() string {
	return fmt.Sprintf("Atomic Swap"+
		"\n    ID:                       %s"+
		"\n    Status:                   %s"+
		"\n    Amount:                   %s"+
		"\n    Random number hash:       %s"+
		"\n    Expire timestamp:         %d"+
		"\n    Timestamp:                %d"+
		"\n    Sender:                   %s"+
		"\n    Recipient:                %s"+
		"\n    Sender other chain:       %s"+
		"\n    Recipient other chain:    %s"+
		"\n    Closed block:             %d"+
		"\n    Cross chain:              %t"+
		"\n    Direction:                %s",
		a.GetSwapID(), a.Status.String(), a.Amount.String(),
		hex.EncodeToString(a.RandomNumberHash), a.ExpireTimestamp,
		a.Timestamp, a.Sender, a.Recipient,
		a.SenderOtherChain, a.RecipientOtherChain, a.ClosedBlock,
		a.CrossChain, a.Direction)
}

// AtomicSwaps is a slice of AtomicSwap
type AtomicSwaps []AtomicSwap

// String implements stringer
func (swaps AtomicSwaps) String() string {
	out := ""
	for _, swap := range swaps {
		out += swap.String() + "\n"
	}
	return out
}

// SwapStatus is the status of an AtomicSwap
type SwapStatus byte

// swap statuses
const (
	NULL      SwapStatus = 0x00
	Open      SwapStatus = 0x01
	Completed SwapStatus = 0x02
	Expired   SwapStatus = 0x03
)

// NewSwapStatusFromString converts string to SwapStatus type
func NewSwapStatusFromString(str string) SwapStatus {
	switch str {
	case "Open", "open":
		return Open
	case "Completed", "completed":
		return Completed
	case "Expired", "expired":
		return Expired
	default:
		return NULL
	}
}

// String returns the string representation of a SwapStatus
func (status SwapStatus) String() string {
	switch status {
	case Open:
		return "Open"
	case Completed:
		return "Completed"
	case Expired:
		return "Expired"
	default:
		return "NULL"
	}
}

// MarshalJSON marshals the SwapStatus
func (status SwapStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

// UnmarshalJSON unmarshals the SwapStatus
func (status *SwapStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*status = NewSwapStatusFromString(s)
	return nil
}

// IsValid returns true if the swap status is valid and false otherwise.
func (status SwapStatus) IsValid() bool {
	if status == Open ||
		status == Completed ||
		status == Expired {
		return true
	}
	return false
}

// SwapDirection is the direction of an AtomicSwap
type SwapDirection byte

const (
	INVALID  SwapDirection = 0x00
	Incoming SwapDirection = 0x01
	Outgoing SwapDirection = 0x02
)

// NewSwapDirectionFromString converts string to SwapDirection type
func NewSwapDirectionFromString(str string) SwapDirection {
	switch str {
	case "Incoming", "incoming", "inc", "I", "i":
		return Incoming
	case "Outgoing", "outgoing", "out", "O", "o":
		return Outgoing
	default:
		return INVALID
	}
}

// String returns the string representation of a SwapDirection
func (direction SwapDirection) String() string {
	switch direction {
	case Incoming:
		return "Incoming"
	case Outgoing:
		return "Outgoing"
	default:
		return "INVALID"
	}
}

// MarshalJSON marshals the SwapDirection
func (direction SwapDirection) MarshalJSON() ([]byte, error) {
	return json.Marshal(direction.String())
}

// UnmarshalJSON unmarshals the SwapDirection
func (direction *SwapDirection) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*direction = NewSwapDirectionFromString(s)
	return nil
}

// IsValid returns true if the swap direction is valid and false otherwise.
func (direction SwapDirection) IsValid() bool {
	if direction == Incoming ||
		direction == Outgoing {
		return true
	}
	return false
}

func NewAugmentedAtomicSwap(swap AtomicSwap) AugmentedAtomicSwap {
	return AugmentedAtomicSwap{
		ID:                  hex.EncodeToString(swap.GetSwapID()),
		Amount:              swap.Amount,
		RandomNumberHash:    swap.RandomNumberHash,
		ExpireTimestamp:     swap.ExpireTimestamp,
		Timestamp:           swap.Timestamp,
		Sender:              swap.Sender,
		Recipient:           swap.Recipient,
		SenderOtherChain:    swap.SenderOtherChain,
		RecipientOtherChain: swap.RecipientOtherChain,
		ClosedBlock:         swap.ClosedBlock,
		Status:              swap.Status,
		CrossChain:          swap.CrossChain,
		Direction:           swap.Direction,
	}
}

type AugmentedAtomicSwaps []AugmentedAtomicSwap
