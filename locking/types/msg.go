package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// locking message types
const (
	TypeMsgCreateLockedDelegation     = "create_locked_delegation"
	TypeMsgRedelegateLockedDelegation = "redelegate_locked_delegations"
	TypeMsgToggleAutoRenew            = "toggle_auto_renew"
	TypeMsgUpdateParams               = "update_params"
)

var (
	_ sdk.Msg = &MsgCreateLockedDelegation{}
	_ sdk.Msg = &MsgRedelegateLockedDelegations{}
	_ sdk.Msg = &MsgToggleAutoRenew{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgCreateLockedDelegation creates a new MsgCreateLockedDelegation
func NewMsgCreateLockedDelegation(
	delAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	amount sdk.Coin,
	lockDuration time.Duration,
	autoRenew bool,
) *MsgCreateLockedDelegation {
	return &MsgCreateLockedDelegation{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
		Amount:           amount,
		LockDuration:     lockDuration,
		AutoRenew:        autoRenew,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgCreateLockedDelegation) Route() string { return RouterKey }

// Type implements the sdk.Msg interface
func (msg MsgCreateLockedDelegation) Type() string { return TypeMsgCreateLockedDelegation }

// GetSigners implements the sdk.Msg interface
func (msg MsgCreateLockedDelegation) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCreateLockedDelegation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgCreateLockedDelegation) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(ErrDelegatorAddressInvalid, ModuleName, err)
	}
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(ErrValidatorAddressInvalid, ModuleName, err)
	}
	if err := ValidatePositiveCoin(msg.Amount); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf(ErrSharesInvalid, ModuleName, err)
	}
	if err := ValidateNonZeroDuration(msg.LockDuration); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf(ErrLockDurationInvalid, ModuleName, err)
	}
	return nil
}

// MsgRedelegateLockedDelegation creates a new MsgRedelegateLockedDelegation
func NewMsgRedelegateLockedDelegations(
	delAddr sdk.AccAddress,
	valSrcAddr,
	valDstAddr sdk.ValAddress,
	ids []uint64,
) *MsgRedelegateLockedDelegations {
	return &MsgRedelegateLockedDelegations{
		DelegatorAddress:    delAddr.String(),
		ValidatorSrcAddress: valSrcAddr.String(),
		ValidatorDstAddress: valDstAddr.String(),
		Ids:                 ids,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgRedelegateLockedDelegations) Route() string { return RouterKey }

// Type implements the sdk.Msg interface
func (msg MsgRedelegateLockedDelegations) Type() string { return TypeMsgRedelegateLockedDelegation }

// GetSigners implements the sdk.Msg interface
func (msg MsgRedelegateLockedDelegations) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgRedelegateLockedDelegations) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgRedelegateLockedDelegations) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(ErrDelegatorAddressInvalid, ModuleName, err)
	}
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorSrcAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(ErrValidatorAddressInvalid, ModuleName, err)
	}
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(ErrValidatorAddressInvalid, ModuleName, err)
	}
	return nil
}

// NewMsgToggleAutoRenew creates a new MsgToggleAutoRenew
func NewMsgToggleAutoRenew(
	delAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	id uint64,
) *MsgToggleAutoRenew {
	return &MsgToggleAutoRenew{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
		Id:               id,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgToggleAutoRenew) Route() string { return RouterKey }

// Type implements the sdk.Msg interface
func (msg MsgToggleAutoRenew) Type() string { return TypeMsgToggleAutoRenew }

// GetSigners implements the sdk.Msg interface
func (msg MsgToggleAutoRenew) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgToggleAutoRenew) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgToggleAutoRenew) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(ErrDelegatorAddressInvalid, ModuleName, err)
	}
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(ErrValidatorAddressInvalid, ModuleName, err)
	}
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (m *MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic executes sanity validation on the provided data
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(ErrAuthorityInvalid, ModuleName, err)
	}
	return m.Params.Validate()
}

// GetSigners returns the expected signers for a MsgUpdateParams message
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}
