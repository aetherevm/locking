package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/aetherevm/locking/locking/types"
)

// TestMsgCreateLockedDelegationValidateBasic tests the ValidateBasic method of the
// NewMsgCreateLockedDelegation type in the types package
func TestMsgCreateLockedDelegationValidateBasic(t *testing.T) {
	addr := sdk.AccAddress([]byte("address"))
	valAddr := sdk.ValAddress([]byte("val"))
	coin := sdk.NewCoin("test", sdk.OneInt())

	tests := []struct {
		name string
		msg  types.MsgCreateLockedDelegation
		pass bool
	}{
		{
			name: "pass",
			msg: *types.NewMsgCreateLockedDelegation(
				addr,
				valAddr,
				coin,
				time.Hour,
				false,
			),
			pass: true,
		},
		{
			name: "fail - bad DelegatorAddress",
			msg: types.MsgCreateLockedDelegation{
				DelegatorAddress: "",
				ValidatorAddress: valAddr.String(),
				Amount:           coin,
				LockDuration:     time.Hour,
			},
			pass: false,
		},
		{
			name: "fail - bad ValidatorAddress",
			msg: types.MsgCreateLockedDelegation{
				DelegatorAddress: addr.String(),
				ValidatorAddress: "",
				Amount:           coin,
				LockDuration:     time.Hour,
			},
			pass: false,
		},
		{
			name: "fail - bad Amount",
			msg: types.MsgCreateLockedDelegation{
				DelegatorAddress: addr.String(),
				ValidatorAddress: valAddr.String(),
				Amount:           sdk.NewCoin("test", sdk.ZeroInt()),
				LockDuration:     time.Hour,
			},
			pass: false,
		},
		{
			name: "fail - bad lock duration",
			msg: types.MsgCreateLockedDelegation{
				DelegatorAddress: addr.String(),
				ValidatorAddress: valAddr.String(),
				Amount:           sdk.NewCoin("test", sdk.OneInt()),
				LockDuration:     0,
			},
			pass: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()

			if tc.pass {
				require.NoError(t, err)

				// Validate the other params
				require.Equal(t, types.RouterKey, tc.msg.Route())
				require.Equal(t, types.TypeMsgCreateLockedDelegation, tc.msg.Type())

				// Test the Get signers
				delegator, err := sdk.AccAddressFromBech32(tc.msg.DelegatorAddress)
				require.NoError(t, err)
				require.Equal(t, []sdk.AccAddress{delegator}, tc.msg.GetSigners())

				// Test the GetSignBytes
				// Since the object never changes, we can remove the lint for gosec
				//nolint:gosec
				bz := types.ModuleCdc.MustMarshalJSON(&tc.msg)
				require.Equal(t, sdk.MustSortJSON(bz), tc.msg.GetSignBytes())
			} else {
				require.Error(t, err)
			}
		})
	}
}

// TestMsgRedelegateLockedDelegationValidateBasic tests the ValidateBasic method of the
// NewMsgRedelegateLockedDelegation type in the types package
func TestMsgRedelegateLockedDelegationValidateBasic(t *testing.T) {
	addr := sdk.AccAddress([]byte("address"))
	valSrcAddr := sdk.ValAddress([]byte("srcval"))
	valDstAddr := sdk.ValAddress([]byte("dstval"))

	tests := []struct {
		name string
		msg  types.MsgRedelegateLockedDelegations
		pass bool
	}{
		{
			name: "pass",
			msg: *types.NewMsgRedelegateLockedDelegations(
				addr,
				valSrcAddr,
				valDstAddr,
				[]uint64{},
			),
			pass: true,
		},
		{
			name: "fail - bad DelegatorAddress",
			msg: types.MsgRedelegateLockedDelegations{
				DelegatorAddress:    "bad",
				ValidatorSrcAddress: valSrcAddr.String(),
				ValidatorDstAddress: valDstAddr.String(),
			},
			pass: false,
		},
		{
			name: "fail - bad ValidatorSrcAddress",
			msg: types.MsgRedelegateLockedDelegations{
				DelegatorAddress:    addr.String(),
				ValidatorSrcAddress: "bad",
				ValidatorDstAddress: valDstAddr.String(),
			},
			pass: false,
		},
		{
			name: "fail - bad ValidatorDstAddress",
			msg: types.MsgRedelegateLockedDelegations{
				DelegatorAddress:    addr.String(),
				ValidatorSrcAddress: valSrcAddr.String(),
				ValidatorDstAddress: "",
			},
			pass: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()

			if tc.pass {
				require.NoError(t, err)

				// Validate the other params
				require.Equal(t, types.RouterKey, tc.msg.Route())
				require.Equal(t, types.TypeMsgRedelegateLockedDelegation, tc.msg.Type())

				// Test the Get signers
				delegator, err := sdk.AccAddressFromBech32(tc.msg.DelegatorAddress)
				require.NoError(t, err)
				require.Equal(t, []sdk.AccAddress{delegator}, tc.msg.GetSigners())

				// Test the GetSignBytes
				// Since the object never changes, we can remove the lint for gosec
				//nolint:gosec
				bz := types.ModuleCdc.MustMarshalJSON(&tc.msg)
				require.Equal(t, sdk.MustSortJSON(bz), tc.msg.GetSignBytes())
			} else {
				require.Error(t, err)
			}
		})
	}
}

// TestMsgToggleAutoRenew tests the ValidateBasic method of the MsgToggleAutoRenew
func TestMsgToggleAutoRenewValidateBasic(t *testing.T) {
	addr := sdk.AccAddress([]byte("address"))
	valAddr := sdk.ValAddress([]byte("val"))

	tests := []struct {
		name string
		msg  types.MsgToggleAutoRenew
		pass bool
	}{
		{
			name: "pass",
			msg: *types.NewMsgToggleAutoRenew(
				addr,
				valAddr,
				0,
			),
			pass: true,
		},
		{
			name: "fail - bad DelegatorAddress",
			msg: types.MsgToggleAutoRenew{
				DelegatorAddress: "",
				ValidatorAddress: valAddr.String(),
			},
			pass: false,
		},
		{
			name: "fail - bad ValidatorAddress",
			msg: types.MsgToggleAutoRenew{
				DelegatorAddress: addr.String(),
				ValidatorAddress: "",
			},
			pass: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()

			if tc.pass {
				require.NoError(t, err)

				// Validate the other params
				require.Equal(t, types.RouterKey, tc.msg.Route())
				require.Equal(t, types.TypeMsgToggleAutoRenew, tc.msg.Type())

				// Test the Get signers
				delegator, err := sdk.AccAddressFromBech32(tc.msg.DelegatorAddress)
				require.NoError(t, err)
				require.Equal(t, []sdk.AccAddress{delegator}, tc.msg.GetSigners())

				// Test the GetSignBytes
				// Since the object never changes, we can remove the lint for gosec
				//nolint:gosec
				bz := types.ModuleCdc.MustMarshalJSON(&tc.msg)
				require.Equal(t, sdk.MustSortJSON(bz), tc.msg.GetSignBytes())
			} else {
				require.Error(t, err)
			}
		})
	}
}

// TestMsgUpdateParamsValidateBasic tests the ValidateBasic method of the MsgUpdateParams
func TestMsgUpdateParamsValidateBasic(t *testing.T) {
	tests := []struct {
		name        string
		msg         types.MsgUpdateParams
		errContains string
	}{
		{
			name: "pass",
			msg: types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.DefaultParams(),
			},
		},
		{
			name: "fail - bad Authority",
			msg: types.MsgUpdateParams{
				Authority: "",
				Params:    types.DefaultParams(),
			},
			errContains: "locking invalid authority address",
		},
		{
			name: "fail - bad param",
			msg: types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.Params{},
			},
			errContains: "locking max entries is invalid",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()

			if tc.errContains == "" {
				require.NoError(t, err)
				// Validate the other params

				// Test the Get signers
				authority, err := sdk.AccAddressFromBech32(tc.msg.Authority)
				require.NoError(t, err, tc.name)
				require.Equal(t, []sdk.AccAddress{authority}, tc.msg.GetSigners(), tc.name)

				// Test the GetSignBytes
				// Since the object never changes, we can remove the lint for gosec
				//nolint:gosec
				bz := types.ModuleCdc.MustMarshalJSON(&tc.msg)
				require.Equal(t, sdk.MustSortJSON(bz), tc.msg.GetSignBytes())
			} else {
				require.ErrorContains(t, err, tc.errContains, tc.name)
			}
		})
	}
}
