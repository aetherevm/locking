// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/aetherevm/locking/locking/types"
)

// ensureDistributionHooksSet ensures that the distribution hooks are set
func (suite *KeeperTestSuite) ensureDistributionHooksSet() {
	defer func() {
		err := recover()
		suite.Require().NotNil(err)
	}()
	suite.app.DistrKeeper.SetHooks(suite.k.DistributionHooks())
}

// TestAfterWithdrawDelegationRewardsDirectCall tests the AfterWithdrawDelegationRewards function from a direct call
func (suite *KeeperTestSuite) TestAfterWithdrawDelegationRewardsDirectCall() {
	// Ensure that the hooks are set
	suite.ensureDistributionHooksSet()

	// This is would start for delegations
	initial := sdk.TokensFromConsensusPower(1_000_000, PowerReduction)

	denom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	delAddr := sdk.AccAddress([]byte("address"))
	validator := suite.app.StakingKeeper.GetAllValidators(suite.ctx)[0]
	valAddr := validator.GetOperator()

	// Setup a delegation entry
	setupDistributionHooksTest(suite, initial.Mul(math.NewInt(250)), delAddr, validator)

	// Create a locking delegation
	// This replicates the case '50 shares at 5%, 20 shares at 4%, 250 del' from locked_delegation_test.go
	_, err := suite.k.CreateLockedDelegationEntry(
		suite.ctx,
		delAddr,
		valAddr,
		initial.Mul(math.NewInt(50)),
		types.NewRate(200, sdk.NewDec(5)),
		false,
	)
	suite.Require().NoError(err)
	_, err = suite.k.CreateLockedDelegationEntry(
		suite.ctx,
		delAddr,
		valAddr,
		initial.Mul(math.NewInt(20)),
		types.NewRate(200, sdk.NewDec(4)),
		false,
	)
	suite.Require().NoError(err)

	// Define big values for testing
	bigInt1, ok := sdk.NewIntFromString("123456789012345678123456789012345678")
	suite.Require().True(ok)
	bigInt2, ok := sdk.NewIntFromString("9999999999999999999999999999999999999999")
	suite.Require().True(ok)

	testCases := []struct {
		name    string
		rewards sdk.Coins
	}{
		{
			"Multiple tokens",
			sdk.NewCoins(
				sdk.NewCoin(denom, sdk.NewInt(100)),
				sdk.NewCoin("test", sdk.NewInt(200)),
			),
		},
		{
			"Small value",
			sdk.NewCoins(
				sdk.NewCoin(denom, sdk.NewInt(1)),
			),
		},
		{
			"Giant value 1",
			sdk.NewCoins(
				sdk.NewCoin(denom, bigInt1),
			),
		},
		{
			"Giant value 2",
			sdk.NewCoins(
				sdk.NewCoin(denom, bigInt2),
			),
		},
		{
			"Giant value 1 and 2 as new denom",
			sdk.NewCoins(
				sdk.NewCoin("test1", bigInt1),
				sdk.NewCoin("test2", bigInt2),
			),
		},
	}
	for _, tc := range testCases {
		// Save the original balance and call the hook
		initialBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, delAddr)
		err = suite.k.DistributionHooks().AfterWithdrawDelegationRewards(
			suite.ctx,
			delAddr,
			valAddr,
			tc.rewards,
		)
		suite.Require().NoError(err)

		// We ensure that the new tokens were transferred
		newBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, delAddr)
		// Ratio of 0.0132
		ratio := sdk.NewDecWithPrec(132, 4)
		expectedRewards := sdk.NewDecCoinsFromCoins(tc.rewards...).MulDecTruncate(ratio)

		// Add the expected rewards on top of the initial balance
		for _, coin := range initialBalance {
			expectedRewards = expectedRewards.Add(sdk.NewDecCoinFromCoin(coin))
		}

		expectedRewardsTruncated, _ := expectedRewards.TruncateDecimal()

		suite.Require().Equal(expectedRewardsTruncated, newBalance)
	}
}

// TestAfterWithdrawDelegationRewardsDistributionClaim tests the AfterWithdrawDelegationRewards function from a claim call from the distribution module
func (suite *KeeperTestSuite) TestAfterWithdrawDelegationRewardsDistributionClaim() {
	// Ensure that the hooks are set
	suite.ensureDistributionHooksSet()

	denom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	delAddr := sdk.AccAddress([]byte("address"))
	validator := suite.app.StakingKeeper.GetAllValidators(suite.ctx)[0]
	valAddr := validator.GetOperator()
	delegationShares := sdk.NewInt(75)

	// Setup a delegation entry
	setupDistributionHooksTest(suite, delegationShares, delAddr, validator)

	// Create a locking delegation
	// This replicates the case 'half shares, same rate, same delegation - output ratio remains the same' from locked_delegation_test.go
	_, err := suite.k.CreateLockedDelegationEntry(
		suite.ctx,
		delAddr,
		valAddr,
		math.NewInt(25),
		types.NewRate(200, sdk.NewDec(5)),
		false,
	)
	suite.Require().NoError(err)
	_, err = suite.k.CreateLockedDelegationEntry(
		suite.ctx,
		delAddr,
		valAddr,
		math.NewInt(50),
		types.NewRate(200, sdk.NewDec(5)),
		false,
	)
	suite.Require().NoError(err)

	// Save the original balance and call the hook
	initialBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, delAddr)

	// Allocate rewards to the validator
	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(1, PowerReduction)
	tokens := sdk.DecCoins{sdk.NewDecCoin(denom, initial)}
	suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, tokens)

	// next block
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	// Now withdraw the delegator rewards
	delegationRewards, err := suite.app.DistrKeeper.WithdrawDelegationRewards(suite.ctx, delAddr, valAddr)
	suite.Require().NoError(err)

	// We ensure that the new tokens were transferred
	newBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, delAddr)
	// Use the same calculation as the distribution module and add the new rewards
	// Ratio of 0.05
	ratio := sdk.NewDecWithPrec(5, 2)
	expectedRewards := sdk.NewDecCoinsFromCoins(delegationRewards...).MulDecTruncate(ratio)

	// Add the delegation rewards to the expected rewards
	for _, coin := range delegationRewards {
		expectedRewards = expectedRewards.Add(sdk.NewDecCoinFromCoin(coin))
	}

	// Add the expected rewards on top of the initial balance
	for _, coin := range initialBalance {
		expectedRewards = expectedRewards.Add(sdk.NewDecCoinFromCoin(coin))
	}

	expectedRewardsTruncated, _ := expectedRewards.TruncateDecimal()
	suite.Require().Equal(expectedRewardsTruncated, newBalance)
}

// TestAfterWithdrawDelegationRewardsBeforeDelegationSharesModified tests the AfterWithdrawDelegationRewards function from a internal hook BeforeDelegationSharesModified
func (suite *KeeperTestSuite) TestAfterWithdrawDelegationRewardsBeforeDelegationSharesModified() {
	// Ensure that the hooks are set
	suite.ensureDistributionHooksSet()

	denom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	delAddr := sdk.AccAddress([]byte("address"))
	validator := suite.app.StakingKeeper.GetAllValidators(suite.ctx)[0]
	valAddr := validator.GetOperator()
	delegationShares := sdk.NewInt(200)

	// Setup a delegation entry
	setupDistributionHooksTest(suite, delegationShares, delAddr, validator)

	// Create a locking delegation
	// This replicates the case 'same shares, same rate, double delegation - output ratio should be halved' from locked_delegation_test.go
	_, err := suite.k.CreateLockedDelegationEntry(
		suite.ctx,
		delAddr,
		valAddr,
		math.NewInt(50),
		types.NewRate(200, sdk.NewDec(5)),
		false,
	)
	suite.Require().NoError(err)
	_, err = suite.k.CreateLockedDelegationEntry(
		suite.ctx,
		delAddr,
		valAddr,
		math.NewInt(50),
		types.NewRate(200, sdk.NewDec(5)),
		false,
	)
	suite.Require().NoError(err)

	// Save the original balance and call the hook
	initialBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, delAddr)

	// Allocate rewards to the validator
	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(1, PowerReduction)
	tokens := sdk.DecCoins{sdk.NewDecCoin(denom, initial)}
	suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, tokens)

	// next block
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	// To activate the hook, we delegate some more to the validator
	// the new delegation should not be considered in the calculation
	// since the hook is called before the delegation is created
	// First we save the delegation rewards from a calculation
	delegation := suite.app.StakingKeeper.Delegation(suite.ctx, delAddr, valAddr)
	endingPeriod := suite.app.DistrKeeper.IncrementValidatorPeriod(suite.ctx, validator)
	delegationRewards := suite.app.DistrKeeper.CalculateDelegationRewards(suite.ctx, validator, delegation, endingPeriod)

	// Now delegate to activate the hook
	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, sdk.NewInt(0), stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	// We ensure that the new tokens were transferred
	newBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, delAddr)

	// Use the same calculation as the distribution module and add the new rewards
	// Ratio of 0.025
	ratio := sdk.NewDecWithPrec(25, 3)
	expectedRewards := delegationRewards.MulDecTruncate(ratio)

	// Add the delegation rewards to the expected rewards
	for _, coin := range delegationRewards {
		expectedRewards = expectedRewards.Add(coin)
	}

	// Add the expected rewards on top of the initial balance
	for _, coin := range initialBalance {
		expectedRewards = expectedRewards.Add(sdk.NewDecCoinFromCoin(coin))
	}

	expectedRewardsTruncated, _ := expectedRewards.TruncateDecimal()
	suite.Require().Equal(expectedRewardsTruncated.String(), newBalance.String())
}

// setupDistributionHooksTest setup a testing delegation
func setupDistributionHooksTest(suite *KeeperTestSuite, delegationValue math.Int, delAddr sdk.AccAddress, validator stakingtypes.Validator) {
	// Set a delegator address and get the current chain validator
	denom := suite.app.StakingKeeper.BondDenom(suite.ctx)

	// Send a few tokens to the delegator
	err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
		sdk.NewCoin(denom, delegationValue.Mul(math.NewInt(2))),
	))
	suite.Require().NoError(err)

	// Create a delegation
	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, delegationValue, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)
}
