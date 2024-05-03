// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	"github.com/aetherevm/locking/locking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestCalculateLockedDelegationRewards tests the CalculateLockedDelegationRewards
func (suite *KeeperTestSuite) TestCalculateLockedDelegationRewards() {
	denom := suite.app.StakingKeeper.BondDenom(suite.ctx)

	delAddr := sdk.AccAddress([]byte("address"))

	rate := types.NewRate(32*30*24*time.Hour, sdk.NewDecWithPrec(55, 1))

	testCases := []struct {
		name                  string
		maleate               func(stakingtypes.Validator)
		rewards               sdk.Coins
		expectedLockingReward sdk.DecCoins
	}{
		{
			"empty reward",
			func(_ stakingtypes.Validator) {},
			sdk.NewCoins(sdk.NewCoin(denom, sdk.ZeroInt())),
			sdk.NewDecCoins(),
		},
		{
			"no delegation",
			func(_ stakingtypes.Validator) {},
			sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt())),
			sdk.NewDecCoins(),
		},
		{
			"no locking delegation",
			func(validator stakingtypes.Validator) {
				// Create a delegation
				delegatedShares := math.NewInt(30)
				_, err := suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, delegatedShares, stakingtypes.Unbonded, validator, true)
				suite.Require().NoError(err)
			},
			sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt())),
			sdk.NewDecCoins(),
		},
		{
			"no locking delegation entries",
			func(validator stakingtypes.Validator) {
				// Create a delegation
				delegatedShares := math.NewInt(30)
				_, err := suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, delegatedShares, stakingtypes.Unbonded, validator, true)
				suite.Require().NoError(err)

				// Create a locking delegation
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					validator.GetOperator(),
					[]types.LockedDelegationEntry{},
				)
				err = suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)
			},
			sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt())),
			sdk.NewDecCoins(),
		},
		{
			"single entry - should receive the 5.5% as rewards",
			func(validator stakingtypes.Validator) {
				// Create a delegation
				delegatedTokens := math.NewInt(30)
				newShares := types.CalculateSharesFromValidator(delegatedTokens, validator)
				_, err := suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, delegatedTokens, stakingtypes.Unbonded, validator, true)
				suite.Require().NoError(err)

				// Create a locking delegation
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					validator.GetOperator(),
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							newShares,
							rate,
							time.Now(),
							false,
							1,
						),
					},
				)
				err = suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)
			},
			sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))),
			sdk.NewDecCoins(
				sdk.NewDecCoin(
					denom,
					sdk.NewInt(100),
				),
			).MulDec(sdk.NewDecWithPrec(55, 3)),
		},
		{
			// Reference to three entries test on locked_delegation_test.go (TestLockedDelegationWeightedRatio)
			"multiple entries - should receive the 5% as rewards",
			func(validator stakingtypes.Validator) {
				// Create a delegation
				delegatedTokens := math.NewInt(400)
				newShares := types.CalculateSharesFromValidator(delegatedTokens, validator)
				_, err := suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, delegatedTokens, stakingtypes.Unbonded, validator, true)
				suite.Require().NoError(err)

				// Create a locking delegation
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					validator.GetOperator(),
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							newShares.QuoInt64(2).Mul(math.LegacyNewDec(60).QuoInt64(200)),
							types.NewRate(100, math.LegacyNewDec(3)),
							time.Now(),
							false,
							1,
						),
						types.NewLockedDelegationEntry(
							newShares.QuoInt64(2).Mul(math.LegacyNewDec(40).QuoInt64(200)),
							types.NewRate(100, math.LegacyNewDec(4)),
							time.Now(),
							false,
							2,
						),
						types.NewLockedDelegationEntry(
							newShares.QuoInt64(2).Mul(math.LegacyNewDec(100).QuoInt64(200)),
							types.NewRate(100, math.LegacyNewDec(2)),
							time.Now(),
							false,
							3,
						),
					},
				)
				err = suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)
			},
			sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))),
			sdk.NewDecCoins(
				sdk.NewDecCoin(
					denom,
					sdk.NewInt(100),
				),
			).MulDec(sdk.NewDecWithPrec(135, 4)),
		},
		{
			"single entry, half locked - should receive the 2.75% as rewards",
			func(validator stakingtypes.Validator) {
				// Create a delegation
				delegatedTokens := math.NewInt(30)
				newShares := types.CalculateSharesFromValidator(delegatedTokens, validator)
				_, err := suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, delegatedTokens, stakingtypes.Unbonded, validator, true)
				suite.Require().NoError(err)

				// Create a locking delegation
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					validator.GetOperator(),
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							newShares.QuoInt64(2),
							rate,
							time.Now(),
							false,
							1,
						),
					},
				)
				err = suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)
			},
			sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))),
			sdk.NewDecCoins(
				sdk.NewDecCoin(
					denom,
					sdk.NewInt(100),
				),
			).MulDec(sdk.NewDecWithPrec(275, 4)),
		},
		{
			"single entry, half locked - should receive the 2.75% as rewards - multi tokens",
			func(validator stakingtypes.Validator) {
				// Create a delegation
				delegatedTokens := math.NewInt(30)
				newShares := types.CalculateSharesFromValidator(delegatedTokens, validator)
				_, err := suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, delegatedTokens, stakingtypes.Unbonded, validator, true)
				suite.Require().NoError(err)

				// Create a locking delegation
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					validator.GetOperator(),
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							newShares.QuoInt64(2),
							rate,
							time.Now(),
							false,
							1,
						),
					},
				)
				err = suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)
			},
			sdk.NewCoins(
				sdk.NewCoin(denom, sdk.NewInt(100)),
				sdk.NewCoin("test", sdk.NewInt(200)),
			),
			sdk.NewDecCoins(
				sdk.NewDecCoin(
					denom,
					sdk.NewInt(100),
				),
				sdk.NewDecCoin(
					"test",
					sdk.NewInt(200),
				),
			).MulDec(sdk.NewDecWithPrec(275, 4)),
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // Restart the whole app each time
		validator := suite.app.StakingKeeper.GetAllValidators(suite.ctx)[0]
		valAddr := validator.GetOperator()

		// Send a few tokens to the delegator
		err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
			sdk.NewCoin(denom, math.NewInt(3000)),
		))
		suite.Require().NoError(err)

		// maleate the system
		tc.maleate(validator)

		// This is the real operation
		reward := suite.k.CalculateLockedDelegationRewards(suite.ctx, delAddr, valAddr, tc.rewards)

		suite.Require().Equal(tc.expectedLockingReward, reward, tc.name)
	}
}
