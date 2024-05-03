// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/aetherevm/locking/locking/types"
)

// TestCreateLockedDelegation tests the msg server TestCreateLockedDelegation
func (suite *KeeperTestSuite) TestCreateLockedDelegation() {
	delAddr := sdk.AccAddress([]byte("address1"))

	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	rate := types.DefaultRates[0]

	testCases := []struct {
		name     string
		maleate  func(sdk.ValAddress) types.MsgCreateLockedDelegation
		expError bool
	}{
		{
			"fail - bad validator addr",
			func(valAddr sdk.ValAddress) types.MsgCreateLockedDelegation {
				return *types.NewMsgCreateLockedDelegation(
					delAddr,
					sdk.ValAddress{},
					sdk.NewCoin(bondDenom, sdk.NewInt(20)),
					rate.Duration,
					false,
				)
			},
			true,
		},
		{
			"fail - validator not found",
			func(valAddr sdk.ValAddress) types.MsgCreateLockedDelegation {
				return *types.NewMsgCreateLockedDelegation(
					delAddr,
					sdk.ValAddress([]byte("val1")),
					sdk.NewCoin(bondDenom, sdk.NewInt(20)),
					rate.Duration,
					false,
				)
			},
			true,
		},
		{
			"fail - bad delegator addr",
			func(valAddr sdk.ValAddress) types.MsgCreateLockedDelegation {
				return *types.NewMsgCreateLockedDelegation(
					sdk.AccAddress{},
					valAddr,
					sdk.NewCoin(bondDenom, sdk.NewInt(20)),
					rate.Duration,
					false,
				)
			},
			true,
		},
		{
			"fail - denom not the bond denom",
			func(valAddr sdk.ValAddress) types.MsgCreateLockedDelegation {
				return *types.NewMsgCreateLockedDelegation(
					delAddr,
					valAddr,
					sdk.NewCoin("test", sdk.NewInt(20)),
					rate.Duration,
					false,
				)
			},
			true,
		},
		{
			"fail - rate not found",
			func(valAddr sdk.ValAddress) types.MsgCreateLockedDelegation {
				return *types.NewMsgCreateLockedDelegation(
					delAddr,
					valAddr,
					sdk.NewCoin(bondDenom, sdk.NewInt(20)),
					rate.Duration+1,
					false,
				)
			},
			true,
		},
		{
			"fail - delegate insufficient funds",
			func(valAddr sdk.ValAddress) types.MsgCreateLockedDelegation {
				return *types.NewMsgCreateLockedDelegation(
					delAddr,
					valAddr,
					sdk.NewCoin(bondDenom, sdk.NewInt(20)),
					rate.Duration,
					false,
				)
			},
			true,
		},
		{
			name: "fail - max entries reached",
			maleate: func(valAddr sdk.ValAddress) types.MsgCreateLockedDelegation {
				err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
					sdk.NewCoin(bondDenom, math.NewInt(20)),
				))
				suite.Require().NoError(err)

				maxEntries := suite.k.MaxEntries(suite.ctx)

				createLDWithEntries(delAddr, valAddr, int(maxEntries), rate, suite)

				return *types.NewMsgCreateLockedDelegation(
					delAddr,
					valAddr,
					sdk.NewCoin(bondDenom, sdk.NewInt(20)),
					rate.Duration,
					false,
				)
			},
			expError: true,
		},
		{
			"pass",
			func(valAddr sdk.ValAddress) types.MsgCreateLockedDelegation {
				err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
					sdk.NewCoin(bondDenom, math.NewInt(20)),
				))
				suite.Require().NoError(err)

				return *types.NewMsgCreateLockedDelegation(
					delAddr,
					valAddr,
					sdk.NewCoin(bondDenom, sdk.NewInt(20)),
					rate.Duration,
					true,
				)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // Restart the whole app each time

		validator := suite.app.StakingKeeper.GetAllValidators(suite.ctx)[0]
		valAddr := validator.GetOperator()

		req := tc.maleate(valAddr)

		_, err := suite.msgSrvr.CreateLockedDelegation(suite.ctx, &req)

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)

			// Check if the delegation was created
			delegation, found := suite.app.StakingKeeper.GetDelegation(
				suite.ctx,
				delAddr,
				valAddr,
			)
			suite.Require().True(found)
			suite.Require().EqualValues(delegation.Shares, req.Amount.Amount)

			// The locked delegation must exists
			lockedDelegation, found := suite.k.GetLockedDelegation(suite.ctx, delAddr, valAddr)
			suite.Require().True(found)
			suite.Require().EqualValues(lockedDelegation.Entries[0].Shares, req.Amount.Amount)

			// It also must exist on the queue
			queuePairs := suite.k.GetAllLockedDelegationQueuePairs(suite.ctx, bigTime)
			suite.Require().Contains(queuePairs, types.LockedDelegationPair{
				DelegatorAddress: delAddr.String(),
				ValidatorAddress: valAddr.String(),
			})
		}
	}
}

// TestRedelegateLockedDelegations tests the msg server RedelegateLockedDelegations
func (suite *KeeperTestSuite) TestRedelegateLockedDelegations() {
	delAddr := sdk.AccAddress([]byte("address1"))

	testCases := []struct {
		name        string
		maleate     func(sdk.ValAddress, sdk.ValAddress) types.MsgRedelegateLockedDelegations
		errContains string
	}{
		{
			"fail - ids bigger than max entries",
			func(srcValAddr sdk.ValAddress, dstValAddr sdk.ValAddress) types.MsgRedelegateLockedDelegations {
				// Reduce the max entries to 5
				params := suite.k.GetParams(suite.ctx)
				params.MaxEntries = 5
				err := suite.k.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				return *types.NewMsgRedelegateLockedDelegations(
					sdk.AccAddress{},
					srcValAddr,
					dstValAddr,
					[]uint64{1, 2, 4, 5, 6, 8},
				)
			},
			"ids list length is bigger than max entries",
		},
		{
			"fail - bad delegator addr",
			func(srcValAddr sdk.ValAddress, dstValAddr sdk.ValAddress) types.MsgRedelegateLockedDelegations {
				return *types.NewMsgRedelegateLockedDelegations(
					sdk.AccAddress{},
					srcValAddr,
					dstValAddr,
					[]uint64{},
				)
			},
			"empty address string is not allowed",
		},
		{
			"fail - bad src validator addr",
			func(srcValAddr sdk.ValAddress, dstValAddr sdk.ValAddress) types.MsgRedelegateLockedDelegations {
				return *types.NewMsgRedelegateLockedDelegations(
					delAddr,
					sdk.ValAddress{},
					dstValAddr,
					[]uint64{},
				)
			},
			"empty address string is not allowed",
		},
		{
			"fail - bad dst validator addr",
			func(srcValAddr sdk.ValAddress, dstValAddr sdk.ValAddress) types.MsgRedelegateLockedDelegations {
				return *types.NewMsgRedelegateLockedDelegations(
					delAddr,
					srcValAddr,
					sdk.ValAddress{},
					[]uint64{},
				)
			},
			"empty address string is not allowed",
		},
		{
			"fail - zero locked shares",
			func(srcValAddr sdk.ValAddress, dstValAddr sdk.ValAddress) types.MsgRedelegateLockedDelegations {
				return *types.NewMsgRedelegateLockedDelegations(
					delAddr,
					srcValAddr,
					dstValAddr,
					[]uint64{},
				)
			},
			"no locked shares were found to redelegate",
		},
		{
			"fail - id not found",
			func(srcValAddr sdk.ValAddress, dstValAddr sdk.ValAddress) types.MsgRedelegateLockedDelegations {
				mintAndCreateLockeDelegations(suite, 1, delAddr, srcValAddr)

				return *types.NewMsgRedelegateLockedDelegations(
					delAddr,
					srcValAddr,
					dstValAddr,
					[]uint64{1234, 123, 3},
				)
			},
			"locked delegation entry for specified id not found",
		},
		{
			"pass - single LD",
			func(srcValAddr sdk.ValAddress, dstValAddr sdk.ValAddress) types.MsgRedelegateLockedDelegations {
				mintAndCreateLockeDelegations(suite, 1, delAddr, srcValAddr)

				return *types.NewMsgRedelegateLockedDelegations(
					delAddr,
					srcValAddr,
					dstValAddr,
					[]uint64{},
				)
			},
			"",
		},
		{
			"pass - multiple delegations",
			func(srcValAddr sdk.ValAddress, dstValAddr sdk.ValAddress) types.MsgRedelegateLockedDelegations {
				mintAndCreateLockeDelegations(suite, 10, delAddr, srcValAddr)

				return *types.NewMsgRedelegateLockedDelegations(
					delAddr,
					srcValAddr,
					dstValAddr,
					[]uint64{},
				)
			},
			"",
		},
		{
			"pass - multiple delegations, specify ids",
			func(srcValAddr sdk.ValAddress, dstValAddr sdk.ValAddress) types.MsgRedelegateLockedDelegations {
				mintAndCreateLockeDelegations(suite, 10, delAddr, srcValAddr)

				return *types.NewMsgRedelegateLockedDelegations(
					delAddr,
					srcValAddr,
					dstValAddr,
					[]uint64{2, 4, 6},
				)
			},
			"",
		},
		{
			"pass - redelegation with valid id list and duplicated items",
			func(srcValAddr sdk.ValAddress, dstValAddr sdk.ValAddress) types.MsgRedelegateLockedDelegations {
				mintAndCreateLockeDelegations(suite, 10, delAddr, srcValAddr)

				return *types.NewMsgRedelegateLockedDelegations(
					delAddr,
					srcValAddr,
					dstValAddr,
					[]uint64{4, 2, 1, 6, 7, 7, 7, 7, 7, 1},
				)
			},
			"",
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // Restart the whole app each time

		validator := suite.app.StakingKeeper.GetAllValidators(suite.ctx)[0]
		srcValAddr := validator.GetOperator()

		// Create a target validator
		pks := simtestutil.CreateTestPubKeys(1)
		dstValAddr := sdk.ValAddress([]byte("val2"))
		dstValidator, err := stakingtypes.NewValidator(dstValAddr, pks[0], stakingtypes.Description{Moniker: "val2"})
		dstValidator.Status = stakingtypes.Bonded
		suite.Require().NoError(err)
		suite.app.StakingKeeper.SetValidator(suite.ctx, dstValidator)
		err = suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, dstValAddr)
		suite.Require().NoError(err)

		req := tc.maleate(srcValAddr, dstValAddr)

		// Update validators info after maleate
		initialSrcValidator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, srcValAddr)
		suite.Require().True(found, tc.name)
		initialDstValidator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, dstValAddr)
		suite.Require().True(found, tc.name)

		// Save the initial LD before redelegating
		initialSrcLD, _ := suite.k.GetLockedDelegation(suite.ctx, delAddr, srcValAddr)

		_, err = suite.msgSrvr.RedelegateLockedDelegations(suite.ctx, &req)

		if tc.errContains != "" {
			suite.Require().ErrorContains(err, tc.errContains, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)

			// Get the entries that should be moved
			exists, foundEntries := initialSrcLD.EntriesForIds(req.Ids)
			suite.Require().True(exists)

			totalFromFoundEntries := math.LegacyZeroDec()
			for _, fEntry := range foundEntries {
				totalFromFoundEntries = totalFromFoundEntries.Add(fEntry.Shares)
			}

			if req.Ids == nil {
				// The source locked delegation must not exists if all items were redelegated
				_, found := suite.k.GetLockedDelegation(suite.ctx, delAddr, srcValAddr)
				suite.Require().False(found)
			}

			// The dest validator must be populated with new items
			lockedDelegation, found := suite.k.GetLockedDelegation(suite.ctx, delAddr, dstValAddr)
			suite.Require().True(found, tc.name)
			suite.Require().Equal(len(foundEntries), len(lockedDelegation.Entries), tc.name)

			// We also must have the dest locked delegation on the queue
			queuePairs := suite.k.GetAllLockedDelegationQueuePairs(suite.ctx, bigTime)
			suite.Require().Contains(queuePairs, types.LockedDelegationPair{
				DelegatorAddress: delAddr.String(),
				ValidatorAddress: dstValAddr.String(),
			})

			// The target delegator must have the correct amount of shares
			delegation := suite.app.StakingKeeper.Delegation(suite.ctx, delAddr, dstValAddr)
			suite.Require().EqualValues(lockedDelegation.TotalShares(), delegation.GetShares(), tc.name)

			// Check if the shares calculation were done correctly
			srcValidator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, srcValAddr)
			suite.Require().True(found, tc.name)
			dstValidator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, dstValAddr)
			suite.Require().True(found, tc.name)

			movedTokens := types.SimulateValidatorSharesRemoval(totalFromFoundEntries, initialSrcValidator)
			dstShares := types.CalculateSharesFromValidator(movedTokens, initialDstValidator)

			// src validator should have less tokens by exactly what was moved
			suite.Require().EqualValues(initialSrcValidator.Tokens.Sub(movedTokens), srcValidator.Tokens, tc.name)
			suite.Require().EqualValues(initialDstValidator.DelegatorShares.Add(dstShares), dstValidator.DelegatorShares, tc.name)
		}
	}
}

// TestToggleAutoRenew tests the msg server ToggleAutoRenew
func (suite *KeeperTestSuite) TestToggleAutoRenew() {
	delAddr := sdk.AccAddress([]byte("address1"))
	valAddr := sdk.ValAddress([]byte("val1"))

	testCases := []struct {
		name     string
		maleate  func() types.MsgToggleAutoRenew
		expError bool
	}{
		{
			"fail - bad delegator addr",
			func() types.MsgToggleAutoRenew {
				return *types.NewMsgToggleAutoRenew(
					sdk.AccAddress{},
					valAddr,
					0,
				)
			},
			true,
		},
		{
			"fail - bad validator addr",
			func() types.MsgToggleAutoRenew {
				return *types.NewMsgToggleAutoRenew(
					delAddr,
					sdk.ValAddress{},
					0,
				)
			},
			true,
		},
		{
			"fail - not locked delegation",
			func() types.MsgToggleAutoRenew {
				return *types.NewMsgToggleAutoRenew(
					delAddr,
					valAddr,
					0,
				)
			},
			true,
		},
		{
			"pass",
			func() types.MsgToggleAutoRenew {
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					valAddr,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(math.LegacyOneDec(), rate, time.Now(), false, 1),
						types.NewLockedDelegationEntry(math.LegacyOneDec(), rate, time.Now(), false, 256),
						types.NewLockedDelegationEntry(math.LegacyOneDec(), rate, time.Now(), true, 2),
					},
				)

				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				return *types.NewMsgToggleAutoRenew(
					delAddr,
					valAddr,
					2,
				)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // Restart the whole app each time

		req := tc.maleate()

		// Save the original locked delegation
		originalLockedDelegation, _ := suite.k.GetLockedDelegation(
			suite.ctx,
			delAddr,
			valAddr,
		)

		_, err := suite.msgSrvr.ToggleAutoRenew(suite.ctx, &req)

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)

			// Get the new locked delegation
			lockedDelegation, found := suite.k.GetLockedDelegation(
				suite.ctx,
				delAddr,
				valAddr,
			)
			suite.Require().True(found)

			// Now check if the entry was really update
			// Other checks are assured by tests on x/locking/types/locked_delegation_test.go
			for i, entry := range lockedDelegation.Entries {
				originalEntry := originalLockedDelegation.Entries[i]
				if entry.Id == req.Id {
					suite.Require().NotEqual(entry.AutoRenew, originalEntry.AutoRenew, tc.name)
					continue
				}
				suite.Require().Equal(entry, originalEntry, tc.name)
			}
		}
	}
}

// TestUpdateParams tests the msg server UpdateParams
func (suite *KeeperTestSuite) TestUpdateParams() {
	testCases := []struct {
		name        string
		msg         types.MsgUpdateParams
		errContains string
	}{
		{
			name: "pass - valid params",
			msg: types.MsgUpdateParams{
				Authority: suite.k.GetAuthority(),
				Params:    types.DefaultParams(),
			},
		},
		{
			name: "invalid - invalid authority",
			msg: types.MsgUpdateParams{
				Authority: "bad",
				Params:    types.DefaultParams(),
			},
			errContains: "invalid authority",
		},
		{
			name: "invalid - invalid param",
			msg: types.MsgUpdateParams{
				Authority: suite.k.GetAuthority(),
				Params:    types.Params{},
			},
			errContains: "locking max entries is invalid",
		},
	}
	for _, tc := range testCases {
		// The message never changes, updating lint
		//nolint:gosec
		_, err := suite.msgSrvr.UpdateParams(suite.ctx, &tc.msg)

		if tc.errContains == "" {
			suite.Require().NoError(err, tc.name)
			// Check if params were updated

			params := suite.k.GetParams(suite.ctx)
			suite.Require().Equal(tc.msg.Params, params, tc.name)
		} else {
			suite.Require().ErrorContains(err, tc.errContains, tc.name)
		}
	}
}

// mintAndCreateLockeDelegations mint new tokens and create new locked delegation
func mintAndCreateLockeDelegations(suite *KeeperTestSuite, amountOfLD int64, delAddr sdk.AccAddress, srcValAddr sdk.ValAddress) {
	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)

	tokens := sdk.TokensFromConsensusPower(1_000_000, PowerReduction)
	// Send some tokens
	err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
		sdk.NewCoin(bondDenom, tokens.Mul(math.NewInt(10))),
	))
	suite.Require().NoError(err)

	for x := int64(0); x < amountOfLD; x++ {
		currTime := suite.ctx.BlockTime()
		suite.ctx = suite.ctx.WithBlockTime(currTime.Add(time.Second))
		// Create a delegation and a locked delegation
		_, err = suite.msgSrvr.CreateLockedDelegation(suite.ctx,
			types.NewMsgCreateLockedDelegation(
				delAddr,
				srcValAddr,
				sdk.NewCoin(bondDenom, tokens.Quo(math.NewInt(x+5))),
				rate.Duration,
				false,
			),
		)
		suite.Require().NoError(err)
	}
}
