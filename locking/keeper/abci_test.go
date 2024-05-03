// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/aetherevm/locking/locking/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Get the last rate on the default rate set
var rate = types.DefaultRates[len(types.DefaultRates)-1]

// This a delegation pair, having a validator and a delegator
type DelegationShares map[string]map[string]math.LegacyDec

// TestEndBlock test the endblock function without any locked delegation
func (suite *KeeperTestSuite) TestEndBlockEmpty() {
	delAddresses, valAddresses, delegationShares := setupEndblockTest(suite)

	suite.Require().NotPanics(func() {
		suite.k.EndBlock(suite.ctx)
	})
	// All the delegations should remain untouched
	for _, del := range delAddresses {
		for _, val := range valAddresses {
			oldDelegation := delegationShares[del.String()][val.String()]
			actualDelegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, del, val)
			suite.Require().True(found)
			suite.Require().EqualValues(oldDelegation, actualDelegation.Shares)
		}
	}
}

// TestEndBlockNoExpiration test the endblock function without any expiration
func (suite *KeeperTestSuite) TestEndBlockNoExpiration() {
	delAddresses, valAddresses, delegationShares := setupEndblockTest(suite)

	// Create a set of locked delegations
	createTestingLDs(suite, delAddresses, valAddresses, delegationShares)

	suite.Require().NotPanics(func() {
		suite.k.EndBlock(suite.ctx)
	})
	// All the delegations should remain untouched
	for _, del := range delAddresses {
		for _, val := range valAddresses {
			oldDelegation := delegationShares[del.String()][val.String()]
			actualDelegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, del, val)
			suite.Require().True(found)
			suite.Require().EqualValues(oldDelegation, actualDelegation.Shares)
		}
	}
}

// TestEndBlockWithLDUndelegate test the endblock with undelegate
func (suite *KeeperTestSuite) TestEndBlockWithLDUndelegate() {
	delAddresses, valAddresses, delegationShares := setupEndblockTest(suite)

	// Create a a locked delegation
	delAddress := delAddresses[len(delAddresses)-1]
	valAddress := valAddresses[len(valAddresses)-1]
	totalDelegated, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, delAddress, valAddress)
	suite.Require().True(found)
	_, err := suite.k.CreateLockedDelegationEntry(
		suite.ctx,
		delAddress,
		valAddress,
		math.Int(totalDelegated.Shares).Quo(math.NewInt(2)),
		rate,
		false,
	)
	suite.Require().NoError(err)

	suite.Require().NotPanics(func() {
		// Move the block head in the future by the rate duration
		newBlockTime := suite.ctx.BlockTime().Add(rate.Duration)
		suite.ctx = suite.ctx.WithBlockTime(newBlockTime)
		suite.k.EndBlock(suite.ctx)
	})
	// All the delegations should remain untouched besides the locked and expired one
	for _, del := range delAddresses {
		for _, val := range valAddresses {
			oldDelegation := delegationShares[del.String()][val.String()]
			actualDelegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, del, val)
			suite.Require().True(found)

			if del.Equals(delAddress) && val.Equals(valAddress) {
				// We should have only the half of what we started
				suite.Require().EqualValues(oldDelegation.Quo(math.LegacyNewDec(2)), actualDelegation.Shares)
				continue
			}

			suite.Require().EqualValues(oldDelegation, actualDelegation.Shares)
		}
	}
}

// TestEndBlockWithLDRenew test the endblock with renew
func (suite *KeeperTestSuite) TestEndBlockWithLDRenew() {
	delAddresses, valAddresses, delegationShares := setupEndblockTest(suite)

	// Create a a locked delegation
	delAddress := delAddresses[len(delAddresses)-1]
	valAddress := valAddresses[len(valAddresses)-1]
	totalDelegated, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, delAddress, valAddress)
	suite.Require().True(found)
	_, err := suite.k.CreateLockedDelegationEntry(
		suite.ctx,
		delAddress,
		valAddress,
		math.Int(totalDelegated.Shares).Quo(math.NewInt(2)),
		rate,
		true,
	)
	suite.Require().NoError(err)

	suite.Require().NotPanics(func() {
		// Move the block head in the future by the rate duration
		newBlockTime := suite.ctx.BlockTime().Add(rate.Duration)
		suite.ctx = suite.ctx.WithBlockTime(newBlockTime)
		suite.k.EndBlock(suite.ctx)
	})
	// Since we have renewed, all the delegations should remain the same
	for _, del := range delAddresses {
		for _, val := range valAddresses {
			oldDelegation := delegationShares[del.String()][val.String()]
			actualDelegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, del, val)
			suite.Require().True(found)
			suite.Require().EqualValues(oldDelegation, actualDelegation.Shares)
		}
	}
}

// TestEndBlockWithBigSet tests the endblock with a bit set of locked delegations with undelegate and renew
func (suite *KeeperTestSuite) TestEndBlockWithBigSet() {
	delAddresses, valAddresses, delegationShares := setupEndblockTest(suite)

	createTestingLDs(suite, delAddresses, valAddresses, delegationShares)

	allLockedDelegations := suite.k.GetAllLockedDelegations(suite.ctx)
	// Build a map on top for quick access
	lockedDelegations := make(map[string]map[string]types.LockedDelegation)
	for _, ld := range allLockedDelegations {
		if lockedDelegations[ld.DelegatorAddress] == nil {
			lockedDelegations[ld.DelegatorAddress] = make(map[string]types.LockedDelegation)
		}
		lockedDelegations[ld.DelegatorAddress][ld.ValidatorAddress] = ld
	}

	suite.Require().NotPanics(func() {
		// Move the block head in the future by the rate duration
		newBlockTime := suite.ctx.BlockTime().Add(types.DefaultRates[1].Duration)
		suite.ctx = suite.ctx.WithBlockTime(newBlockTime)
		suite.k.EndBlock(suite.ctx)
	})

	// Check all the new delegations
	for _, del := range delAddresses {
		for _, val := range valAddresses {
			ld := lockedDelegations[del.String()][val.String()]

			undelegated := math.LegacyZeroDec()
			for _, entry := range ld.Entries {
				if entry.Expired(suite.ctx.BlockTime()) {
					if !entry.AutoRenew {
						undelegated = undelegated.Add(entry.Shares)
					}
				}
			}

			oldDelegation := delegationShares[del.String()][val.String()]
			actualDelegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, del, val)
			suite.Require().True(found)

			// The delegation should have changed exactly by the undelegated amount
			suite.Require().EqualValues(oldDelegation.Sub(undelegated), actualDelegation.Shares)
		}
	}
}

// setupEndblockTest prepares the endblock testing
// We create random delegations
func setupEndblockTest(suite *KeeperTestSuite) (delAddresses []sdk.AccAddress, valAddresses []sdk.ValAddress, delegationShares DelegationShares) {
	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	delegationShares = make(map[string]map[string]math.LegacyDec)
	pks := simtestutil.CreateTestPubKeys(5)

	// Create a few more validators
	for i := 1; i < 5; i++ {
		valMoniker := "val" + fmt.Sprintln(i+1)
		valAddr := sdk.ValAddress([]byte(valMoniker))
		validator, err := stakingtypes.NewValidator(valAddr, pks[i], stakingtypes.Description{Moniker: valMoniker})
		validator.Status = stakingtypes.Bonded
		suite.Require().NoError(err)
		suite.app.StakingKeeper.SetValidator(suite.ctx, validator)
		err = suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, valAddr)
		suite.Require().NoError(err)
	}

	// Create a few delegations on top of the validators
	validators := suite.app.StakingKeeper.GetAllValidators(suite.ctx)
	for i, validator := range validators {
		for j := 1; j < 5; j++ {
			delAddr := sdk.AccAddress([]byte("address" + fmt.Sprint((j))))
			delegatedTokens := sdk.TokensFromConsensusPower(int64(100*(i*j+1)), PowerReduction)
			err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
				sdk.NewCoin(bondDenom, delegatedTokens),
			))
			suite.Require().NoError(err)
			newShares, err := suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, delegatedTokens, stakingtypes.Unbonded, validator, true)
			suite.Require().NoError(err)

			// Save the shares
			if delegationShares[delAddr.String()] == nil {
				delegationShares[delAddr.String()] = make(map[string]math.LegacyDec)
			}
			delegationShares[delAddr.String()][validator.GetOperator().String()] = newShares
			// Append to our final list of delegator addresses
			delAddresses = append(delAddresses, delAddr)
		}
		// Append to our final list of delegator addresses
		valAddresses = append(valAddresses, validator.GetOperator())
	}

	return delAddresses, valAddresses, delegationShares
}

// createTestingLDs creates a set of locked delegations for endblock testing
func createTestingLDs(suite *KeeperTestSuite, delAddresses []sdk.AccAddress, valAddresses []sdk.ValAddress, delegationShares DelegationShares) {
	blockTime := suite.ctx.BlockTime()

	for _, del := range delAddresses {
		for i, val := range valAddresses {
			delegation := delegationShares[del.String()][val.String()]
			if !delegation.IsZero() {
				for j := 1; j < 5; j++ {
					// Go back hours to not create duplicated entries
					suite.ctx = suite.ctx.WithBlockTime(blockTime.Add(time.Hour * time.Duration(i) * -1))
					_, err := suite.k.CreateLockedDelegationEntry(
						suite.ctx,
						del,
						val,
						math.NewInt(int64(10*(i+1))),
						types.DefaultRates[i%len(types.DefaultRates)],
						i%2 == 0,
					)
					suite.Require().NoError(err)
				}
			}
		}
	}

	// Restore the block time
	suite.ctx = suite.ctx.WithBlockTime(blockTime)
}
