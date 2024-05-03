// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/aetherevm/locking/locking/types"
)

// ensureStakingHooksSet ensures that the staking hooks are set
func (suite *KeeperTestSuite) ensureStakingHooksSet() {
	defer func() {
		err := recover()
		suite.Require().NotNil(err)
	}()
	suite.app.StakingKeeper.SetHooks(suite.k.StakingHooks())
}

// TestDelegationModified tests the hooks AfterDelegationModified and BeforeDelegationRemoved
func (suite *KeeperTestSuite) TestDelegationModified() {
	delAddr := sdk.AccAddress([]byte("address1"))
	delegatedShares := sdk.NewInt(200)

	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	rate := types.DefaultRates[0]
	powerReduction := sdk.NewDecFromInt(suite.app.StakingKeeper.PowerReduction(suite.ctx))

	testCases := []struct {
		name             string
		execute          func(valAddr sdk.ValAddress, validator stakingtypes.Validator)
		undelegateAmount math.Int
		expError         bool
	}{
		{
			"fail - try to undelegate with locked delegation",
			func(valAddr sdk.ValAddress, validator stakingtypes.Validator) {
				// Send tokens to the delegator
				err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
					sdk.NewCoin(bondDenom, delegatedShares),
				))
				suite.Require().NoError(err)

				// Delegate some and create a locked delegation
				// Create a delegation and a locked delegation
				_, err = suite.msgSrvr.CreateLockedDelegation(suite.ctx,
					types.NewMsgCreateLockedDelegation(
						delAddr,
						valAddr,
						sdk.NewCoin(bondDenom, sdk.NewInt(20)),
						rate.Duration,
						false,
					),
				)
				suite.Require().NoError(err)
			},
			sdk.NewInt(1),
			true,
		},
		{
			"fail - try to undelegate with locked delegation (exact value)",
			func(valAddr sdk.ValAddress, validator stakingtypes.Validator) {
				// Send tokens to the delegator
				err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
					sdk.NewCoin(bondDenom, delegatedShares),
				))
				suite.Require().NoError(err)

				// Delegate some and create a locked delegation
				// Create a delegation and a locked delegation
				_, err = suite.msgSrvr.CreateLockedDelegation(suite.ctx,
					types.NewMsgCreateLockedDelegation(
						delAddr,
						valAddr,
						sdk.NewCoin(bondDenom, sdk.NewInt(20)),
						rate.Duration,
						false,
					),
				)
				suite.Require().NoError(err)
			},
			sdk.NewInt(20),
			true,
		},
		{
			"good path - try to undelegate with locked delegation (by one)",
			func(valAddr sdk.ValAddress, validator stakingtypes.Validator) {
				// Send tokens to the delegator
				err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
					sdk.NewCoin(bondDenom, delegatedShares),
				))
				suite.Require().NoError(err)

				// Delegate some and create a locked delegation
				// Create a delegation and a locked delegation
				_, err = suite.msgSrvr.CreateLockedDelegation(suite.ctx,
					types.NewMsgCreateLockedDelegation(
						delAddr,
						valAddr,
						sdk.NewCoin(bondDenom, sdk.NewInt(20)),
						rate.Duration,
						false,
					),
				)
				suite.Require().NoError(err)

				// Delegate one more
				_, err = suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, sdk.NewInt(21), stakingtypes.Unbonded, validator, true)
				suite.Require().NoError(err)
			},
			sdk.NewInt(1),
			false,
		},
		{
			"good path",
			func(valAddr sdk.ValAddress, validator stakingtypes.Validator) {
				// Send tokens to the delegator
				err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
					sdk.NewCoin(bondDenom, delegatedShares),
				))
				suite.Require().NoError(err)

				// Delegate some and create a locked delegation
				// Create a delegation and a locked delegation
				_, err = suite.msgSrvr.CreateLockedDelegation(suite.ctx,
					types.NewMsgCreateLockedDelegation(
						delAddr,
						valAddr,
						sdk.NewCoin(bondDenom, sdk.NewInt(20)),
						rate.Duration,
						false,
					),
				)
				suite.Require().NoError(err)

				// And a delegation
				_, err = suite.app.StakingKeeper.Delegate(
					suite.ctx,
					delAddr,
					sdk.NewInt(10),
					stakingtypes.Unbonded,
					validator,
					true,
				)
				suite.Require().NoError(err)
			},
			sdk.NewInt(1),
			false,
		},
		{
			"good path - no locked delegation",
			func(valAddr sdk.ValAddress, validator stakingtypes.Validator) {
				// Send tokens to the delegator
				err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
					sdk.NewCoin(bondDenom, delegatedShares),
				))
				suite.Require().NoError(err)

				// And a delegation
				_, err = suite.app.StakingKeeper.Delegate(
					suite.ctx,
					delAddr,
					sdk.NewInt(10),
					stakingtypes.Unbonded,
					validator,
					true,
				)
				suite.Require().NoError(err)
			},
			sdk.NewInt(1),
			false,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // Restart the whole app each time
		suite.ensureStakingHooksSet()

		// Create a new validator
		validator := suite.app.StakingKeeper.GetAllValidators(suite.ctx)[0]
		valAddr := validator.GetOperator()
		err := suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, valAddr)
		suite.Require().NoError(err)

		tc.execute(valAddr, validator)

		// Undelegate call the hook we want to test
		totalUndelegateDec := sdk.NewDecFromInt(tc.undelegateAmount).Quo(powerReduction)
		_, err = suite.app.StakingKeeper.Undelegate(
			suite.ctx,
			delAddr,
			valAddr,
			totalUndelegateDec,
		)

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}
