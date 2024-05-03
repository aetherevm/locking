// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package locking_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	tmtypesproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aetherevm/locking/locking"
	"github.com/aetherevm/locking/locking/types"
	"github.com/aetherevm/locking/testing/simapp"
)

var (
	TestChainId = "chain-0"
)

// GenesisTestSuite defines a suite for testing genesis functionalities of the locking module
type GenesisTestSuite struct {
	suite.Suite
	ctx     sdk.Context
	app     *simapp.SimApp
	genesis types.GenesisState
}

// TestGenesisTestSuite initiates the test run for the GenesisTestSuite
func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

// SetupTest initializes the context, app, and genesis for each test in the suite
func (suite *GenesisTestSuite) SetupTest() {
	suite.app = simapp.Setup(suite.T(), false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmtypesproto.Header{Height: 1, ChainID: TestChainId})

	suite.genesis = *types.DefaultGenesis()
}

// TestInitGenesis tests the initialization of the genesis state for the locking module
func (suite *GenesisTestSuite) TestInitGenesis() {
	addr := sdk.AccAddress([]byte("address1"))
	valAddr := sdk.ValAddress([]byte("val1"))

	rate := types.DefaultRates[0]

	testCases := []struct {
		name         string
		genesisState *types.GenesisState
	}{
		{
			"default genesis",
			types.DefaultGenesis(),
		},
		{
			"custom genesis",
			types.NewGenesisState(
				types.NewParams(
					types.DefaultMaxEntries+1,
					[]types.Rate{
						types.NewRate(10, sdk.OneDec()),
					},
				),
				[]types.LockedDelegation{
					types.NewLockedDelegation(
						addr,
						valAddr,
						[]types.LockedDelegationEntry{
							{
								Shares:    math.LegacyOneDec(),
								Rate:      rate,
								UnlockOn:  time.Now().UTC(),
								AutoRenew: false,
								Id:        421,
							},
							{
								Shares:    math.LegacyOneDec(),
								Rate:      rate,
								UnlockOn:  time.Now().UTC().Add(1),
								AutoRenew: false,
								Id:        65,
							},
							{
								Shares:    math.LegacyOneDec(),
								Rate:      rate,
								UnlockOn:  time.Now().UTC().Add(2),
								AutoRenew: false,
								Id:        542,
							},
						},
					),
				},
			),
		},
	}

	for _, tc := range testCases {
		suite.Require().NotPanics(func() {
			locking.InitGenesis(suite.ctx, suite.app.LockingKeeper, *tc.genesisState)
		})
		params := suite.app.LockingKeeper.GetParams(suite.ctx)
		suite.Require().Equal(tc.genesisState.Params, params, tc.name)

		lockedDelegations := suite.app.LockingKeeper.GetAllLockedDelegations(suite.ctx)
		suite.Require().ElementsMatch(tc.genesisState.LockedDelegations, lockedDelegations, tc.name)

		// The queue also must exist
		bigTime := time.Unix(1<<63-1, 0)
		outcome := suite.app.LockingKeeper.GetAllLockedDelegationQueuePairs(
			suite.ctx,
			bigTime,
		)
		var pairs []types.LockedDelegationPair
		for _, lockedDelegation := range tc.genesisState.LockedDelegations {
			for range lockedDelegation.Entries {
				pairs = append(pairs, types.LockedDelegationPair{
					DelegatorAddress: lockedDelegation.DelegatorAddress,
					ValidatorAddress: lockedDelegation.ValidatorAddress,
				})
			}
		}
		suite.Require().ElementsMatch(outcome, pairs, tc.name)

		// Check if the initial index was set
		initialID := uint64(0)
		for _, ld := range tc.genesisState.LockedDelegations {
			for _, entry := range ld.Entries {
				// The look up for this entry must exist
				foundLd, found := suite.app.LockingKeeper.GetLockedDelegationByEntryID(suite.ctx, entry.Id)
				suite.Require().True(found)
				suite.Require().Equal(ld, foundLd, tc.name)

				if entry.Id > initialID {
					initialID = entry.Id
				}
			}
		}

		nextID := suite.app.LockingKeeper.IncrementLockedDelegationEntryID(suite.ctx)
		suite.Require().Equal(initialID+1, nextID, tc.name)
	}
}

// TestExportGenesis tests the export functionality of the locking module's genesis state
func (suite *GenesisTestSuite) TestExportGenesis() {
	addr := sdk.AccAddress([]byte("address1"))
	valAddr := sdk.ValAddress([]byte("val1"))

	rate := types.DefaultRates[0]

	testGenCases := []struct {
		name         string
		genesisState *types.GenesisState
	}{
		{
			"default genesis",
			types.DefaultGenesis(),
		},
		{
			"custom genesis",
			types.NewGenesisState(
				types.NewParams(
					types.DefaultMaxEntries+1,
					[]types.Rate{
						types.NewRate(10, sdk.OneDec()),
					},
				),
				[]types.LockedDelegation{
					types.NewLockedDelegation(
						addr,
						valAddr,
						[]types.LockedDelegationEntry{
							{
								Shares:    math.LegacyOneDec(),
								Rate:      rate,
								UnlockOn:  time.Now().UTC(),
								AutoRenew: false,
								Id:        421,
							},
							{
								Shares:    math.LegacyOneDec(),
								Rate:      rate,
								UnlockOn:  time.Now().UTC().Add(1),
								AutoRenew: false,
								Id:        65,
							},
							{
								Shares:    math.LegacyOneDec(),
								Rate:      rate,
								UnlockOn:  time.Now().UTC().Add(2),
								AutoRenew: false,
								Id:        542,
							},
						},
					),
				},
			),
		},
	}

	for _, tc := range testGenCases {
		locking.InitGenesis(suite.ctx, suite.app.LockingKeeper, *tc.genesisState)
		suite.Require().NotPanics(func() {
			genesisExported := locking.ExportGenesis(suite.ctx, suite.app.LockingKeeper)
			suite.Require().Equal(tc.genesisState.Params, genesisExported.Params)
			suite.Require().Equal(tc.genesisState.LockedDelegations, genesisExported.LockedDelegations)
		})
	}
}

// TestInitGenesisAddrBadPath tests a specific path were genesis store fails on bad validatorAddress
func (suite *GenesisTestSuite) TestInitGenesisAddrBadPath() {
	suite.SetupTest()

	// First test a bad Param
	genesisState := types.NewGenesisState(
		types.NewParams(types.DefaultMaxEntries, []types.Rate{types.NewRate(0, sdk.ZeroDec())}),
		[]types.LockedDelegation{},
	)
	suite.Require().Panics(
		func() {
			locking.InitGenesis(suite.ctx, suite.app.LockingKeeper, *genesisState)
		},
	)

	// Build a lockedDelegation where the validator is empty
	addr := sdk.AccAddress([]byte("address"))
	genesisState = types.NewGenesisState(
		types.NewParams(types.DefaultMaxEntries, nil),
		[]types.LockedDelegation{
			{
				DelegatorAddress: addr.String(),
				ValidatorAddress: "",
				Entries:          nil,
			},
		},
	)
	suite.Require().Panics(
		func() {
			locking.InitGenesis(suite.ctx, suite.app.LockingKeeper, *genesisState)
		},
	)
}
