package types_test

import (
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aetherevm/locking/locking/types"
)

// TestGenesisStateValidate tests the Validate method of GenesisState struct
func TestGenesisStateValidate(t *testing.T) {
	addr := sdk.AccAddress([]byte("address"))
	valAddr := sdk.ValAddress([]byte("val"))
	valAddr2 := sdk.ValAddress([]byte("val2"))

	for _, tc := range []struct {
		desc     string
		genState types.GenesisState
		valid    bool
	}{
		{
			desc:     "valid - default genesis",
			genState: *types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid - changed genesis state",
			genState: *types.NewGenesisState(
				types.Params{
					MaxEntries: types.DefaultMaxEntries,
				},
				nil,
			),
			valid: true,
		},
		{
			desc: "invalid - bad params",
			genState: *types.NewGenesisState(
				types.Params{
					MaxEntries: 0,
				},
				nil,
			),
			valid: false,
		},
		{
			desc: "invalid - bad locked delegation",
			genState: *types.NewGenesisState(
				types.Params{
					MaxEntries: types.DefaultMaxEntries,
				},
				[]types.LockedDelegation{
					{
						"",
						"",
						nil,
					},
				},
			),
			valid: false,
		},
		{
			desc: "invalid - duplicated locked delegation",
			genState: *types.NewGenesisState(
				types.Params{
					MaxEntries: types.DefaultMaxEntries,
				},
				[]types.LockedDelegation{
					types.NewLockedDelegation(addr, valAddr, nil),
					types.NewLockedDelegation(addr, valAddr, nil),
				},
			),
			valid: false,
		},
		{
			desc: "invalid - duplicated locked delegation ID",
			genState: *types.NewGenesisState(
				types.Params{
					MaxEntries: types.DefaultMaxEntries,
				},
				[]types.LockedDelegation{
					types.NewLockedDelegation(
						addr,
						valAddr,
						[]types.LockedDelegationEntry{
							{Id: 0, Shares: math.LegacyOneDec(), Rate: types.DefaultRates[0], UnlockOn: time.Now()},
							{Id: 1, Shares: math.LegacyOneDec(), Rate: types.DefaultRates[0], UnlockOn: time.Now()},
							{Id: 2, Shares: math.LegacyOneDec(), Rate: types.DefaultRates[0], UnlockOn: time.Now()},
							{Id: 3, Shares: math.LegacyOneDec(), Rate: types.DefaultRates[0], UnlockOn: time.Now()},
							{Id: 4, Shares: math.LegacyOneDec(), Rate: types.DefaultRates[0], UnlockOn: time.Now()},
						},
					),
					types.NewLockedDelegation(
						addr,
						valAddr2,
						[]types.LockedDelegationEntry{
							{Id: 5, Shares: math.LegacyOneDec(), Rate: types.DefaultRates[0], UnlockOn: time.Now()},
							{Id: 6, Shares: math.LegacyOneDec(), Rate: types.DefaultRates[0], UnlockOn: time.Now()},
							{Id: 4, Shares: math.LegacyOneDec(), Rate: types.DefaultRates[0], UnlockOn: time.Now()},
							{Id: 7, Shares: math.LegacyOneDec(), Rate: types.DefaultRates[0], UnlockOn: time.Now()},
							{Id: 8, Shares: math.LegacyOneDec(), Rate: types.DefaultRates[0], UnlockOn: time.Now()},
						},
					),
				},
			),
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err, tc.desc)
			} else {
				require.Error(t, err, tc.desc)
			}
		})
	}
}

// GetGenesisStateFromAppState tests the method GetGenesisStateFromAppState
func TestGetGenesisStateFromAppState(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	for _, tc := range []struct {
		desc              string
		maleate           func() map[string]json.RawMessage
		expecGenesisState types.GenesisState
	}{
		{
			desc: "empty genesis",
			maleate: func() map[string]json.RawMessage {
				appState := make(map[string]json.RawMessage)
				rawMessage := json.RawMessage(`
					{
						"params": null
					}
				`)
				appState[types.ModuleName] = rawMessage

				return appState
			},
			expecGenesisState: types.GenesisState{Params: types.Params{}},
		},
		{
			desc: "empty genesis with max entries 1",
			maleate: func() map[string]json.RawMessage {
				appState := make(map[string]json.RawMessage)
				rawMessage := json.RawMessage(`
					{
						"params": {
							"max_entries":1
						}
					}
				`)
				appState[types.ModuleName] = rawMessage

				return appState
			},
			expecGenesisState: types.GenesisState{Params: types.Params{MaxEntries: 1}},
		},
		{
			desc: "genesis with max entries 1 but bad module name",
			maleate: func() map[string]json.RawMessage {
				appState := make(map[string]json.RawMessage)
				rawMessage := json.RawMessage(`
					{
						"params": {
							"max_entries":1
						}
					}
				`)

				// Bad moduleName
				appState["test"] = rawMessage

				return appState
			},
			expecGenesisState: types.GenesisState{},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			appState := tc.maleate()

			genesisState := types.GetGenesisStateFromAppState(cdc, appState)

			// Since we are not changing the values in the loop, we can nolint gosec
			//nolint:gosec
			require.Equal(t, &tc.expecGenesisState, genesisState, tc.desc)
		})
	}
}
