package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aetherevm/locking/locking/keeper"
	"github.com/aetherevm/locking/locking/tests"
	"github.com/aetherevm/locking/locking/types"
)

// TestParams checks the functionality of the Params method in the locking keeper
func (suite *KeeperTestSuite) TestGRPCParams() {
	// Get a default param
	expcParams := types.DefaultParams()

	c := sdk.WrapSDKContext(suite.ctx)
	response, err := suite.k.Params(c, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expcParams, response.Params)

	// Now modify and save
	expcParams.MaxEntries++
	err = suite.k.SetParams(suite.ctx, expcParams)
	suite.Require().NoError(err)

	// Now fetch again and validate
	response, err = suite.k.Params(c, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expcParams, response.Params)
}

// TestLockedDelegation tests the LockedDelegation from the query server
func (suite *KeeperTestSuite) TestLockedDelegation() {
	// Start the test with a few delegations in store
	addresses, valAddresses := createLockedDelegations(suite)

	testCases := []struct {
		name    string
		request *types.QueryLockedDelegationRequest
		pass    bool
	}{
		{
			"fail - Empty request",
			nil,
			false,
		},
		{
			"fail - Empty delegator",
			&types.QueryLockedDelegationRequest{
				DelegatorAddr: "",
				ValidatorAddr: valAddresses[0].String(),
			},
			false,
		},
		{
			"fail - Empty validator",
			&types.QueryLockedDelegationRequest{
				DelegatorAddr: addresses[0].String(),
				ValidatorAddr: "",
			},
			false,
		},
		{
			"fail - invalid delegator",
			&types.QueryLockedDelegationRequest{
				DelegatorAddr: "test",
				ValidatorAddr: valAddresses[0].String(),
			},
			false,
		},
		{
			"fail - invalid validator",
			&types.QueryLockedDelegationRequest{
				DelegatorAddr: addresses[0].String(),
				ValidatorAddr: "test",
			},
			false,
		},
		{
			"pass - Returns the correct values for the addresses",
			&types.QueryLockedDelegationRequest{
				DelegatorAddr: addresses[0].String(),
				ValidatorAddr: valAddresses[0].String(),
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			res, err := suite.k.LockedDelegations(suite.ctx, tc.request)

			if tc.pass {
				suite.Require().NoError(err, tc.name)

				// Should only return the requested information (based on the addresses sent)
				var expcLockedDelegations []types.LockedDelegationWithTotalShares
				for _, lockedDelegation := range suite.k.GetAllLockedDelegations(suite.ctx) {
					if lockedDelegation.DelegatorAddress == tc.request.DelegatorAddr && lockedDelegation.ValidatorAddress == tc.request.ValidatorAddr {
						expcLockedDelegations = append(expcLockedDelegations, types.LockedDelegationWithTotalShares{
							LockedDelegation: lockedDelegation,
							TotalLocked:      lockedDelegation.TotalShares(),
						})
					}
				}

				suite.Require().NotEmpty(res.LockedDelegations, tc.name)
				suite.Require().ElementsMatch(expcLockedDelegations, res.LockedDelegations, tc.name)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

// TestLockedDelegation tests the LockedDelegation from the query server
func (suite *KeeperTestSuite) TestDelegatorLockedDelegations() {
	// Start the test with a few delegations in store
	addresses, _ := createLockedDelegations(suite)

	testCases := []struct {
		name    string
		request *types.QueryDelegatorLockedDelegationsRequest
		pass    bool
	}{
		{
			"fail - Empty request",
			nil,
			false,
		},
		{
			"fail - Empty delegator",
			&types.QueryDelegatorLockedDelegationsRequest{
				DelegatorAddr: "",
			},
			false,
		},
		{
			"fail - invalid delegator",
			&types.QueryDelegatorLockedDelegationsRequest{
				DelegatorAddr: "test",
			},
			false,
		},
		{
			"pass - Returns the correct values for the addresses",
			&types.QueryDelegatorLockedDelegationsRequest{
				DelegatorAddr: addresses[0].String(),
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			res, err := suite.k.DelegatorLockedDelegations(suite.ctx, tc.request)

			if tc.pass {
				suite.Require().NoError(err, tc.name)

				// Should only return the requested information (based on the addresses sent)
				var expcLockedDelegations []types.LockedDelegationWithTotalShares
				for _, lockedDelegation := range suite.k.GetAllLockedDelegations(suite.ctx) {
					if lockedDelegation.DelegatorAddress == tc.request.DelegatorAddr {
						expcLockedDelegations = append(expcLockedDelegations, types.LockedDelegationWithTotalShares{
							LockedDelegation: lockedDelegation,
							TotalLocked:      lockedDelegation.TotalShares(),
						})
					}
				}

				suite.Require().NotEmpty(res.LockedDelegations, tc.name)
				suite.Require().ElementsMatch(expcLockedDelegations, res.LockedDelegations, tc.name)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

// TestLockedDelegationRewards tests the LockedDelegationRewards from the query server
func (suite *KeeperTestSuite) TestLockedDelegationRewards() {
	// Start the test with a few delegations in store
	addresses, valAddresses := createLockedDelegations(suite)
	validator := suite.app.StakingKeeper.GetAllValidators(suite.ctx)[0]
	valAddr := validator.GetOperator()

	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)

	testCases := []struct {
		name        string
		maleate     func()
		request     *types.QueryLockedDelegationRewardsRequest
		errContains string
	}{
		{
			"fail - Empty request",
			func() {},
			nil,
			"empty request",
		},
		{
			"fail - Empty delegator",
			func() {},
			&types.QueryLockedDelegationRewardsRequest{
				DelegatorAddress: "",
				ValidatorAddress: valAddresses[0].String(),
			},
			"delegator address cannot be empty",
		},
		{
			"fail - invalid validator",
			func() {},
			&types.QueryLockedDelegationRewardsRequest{
				DelegatorAddress: addresses[0].String(),
				ValidatorAddress: "test",
			},
			"invalid bech32 string",
		},
		{
			"fail - no validator exists",
			func() {},
			&types.QueryLockedDelegationRewardsRequest{
				DelegatorAddress: addresses[0].String(),
				ValidatorAddress: valAddresses[0].String(),
			},
			"validator does not exist",
		},
		{
			"fail - invalid delegator",
			func() {},
			&types.QueryLockedDelegationRewardsRequest{
				DelegatorAddress: "test",
				ValidatorAddress: valAddr.String(),
			},
			"invalid bech32 string",
		},
		{
			"fail - delegation doesn't exists",
			func() {},
			&types.QueryLockedDelegationRewardsRequest{
				DelegatorAddress: addresses[0].String(),
				ValidatorAddress: valAddr.String(),
			},
			"delegation does not exist",
		},
		{
			"pass",
			func() {
				// Send tokens and create a locked delegation
				err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(
					suite.ctx,
					types.ModuleName,
					addresses[0],
					sdk.NewCoins(
						sdk.NewCoin(bondDenom, sdk.NewInt(1_000_000)),
					))
				suite.Require().NoError(err)
				_, err = suite.app.LockingKeeper.CreateLockedDelegationEntryAndDelegate(
					suite.ctx,
					addresses[0],
					validator.GetOperator(),
					sdk.NewInt(1_000_000),
					types.DefaultRates[0].Duration,
					true,
				)
				suite.Require().NoError(err)

				// Allocate rewards to the validator
				// allocate some rewards
				initial := sdk.TokensFromConsensusPower(10, PowerReduction)
				tokens := sdk.DecCoins{sdk.NewDecCoin(bondDenom, initial)}
				suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, tokens)

				// next block
				suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
			},
			&types.QueryLockedDelegationRewardsRequest{
				DelegatorAddress: addresses[0].String(),
				ValidatorAddress: valAddr.String(),
			},
			"",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.maleate()

			res, err := suite.k.LockedDelegationRewards(suite.ctx, tc.request)
			if tc.errContains == "" {
				suite.Require().NoError(err, tc.name)

				// Check if the values are correct
				expectedLocked := res.DistributionReward.MulDec(types.DefaultRates[0].Rate).QuoDec(math.LegacyNewDec(100))
				expected, _ := expectedLocked.TruncateDecimal()
				actual, _ := res.LockingReward.TruncateDecimal()
				suite.Require().EqualValues(expected, actual, tc.name)

				// Check if the total is correct
				suite.Require().EqualValues(res.DistributionReward.Add(res.LockingReward...), res.Total, tc.name)
			} else {
				suite.Require().ErrorContains(err, tc.errContains, tc.name)
			}
		})
	}
}

// TestLockedDelegationTotalRewards tests the LockedDelegationTotalRewards from the query server
func (suite *KeeperTestSuite) TestLockedDelegationTotalRewards() {
	// Start the test with a few delegations in store
	addresses, _ := createLockedDelegations(suite)
	validator := suite.app.StakingKeeper.GetAllValidators(suite.ctx)[0]

	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)

	testCases := []struct {
		name        string
		maleate     func()
		request     *types.QueryLockedDelegationTotalRewardsRequest
		errContains string
	}{
		{
			"fail - Empty request",
			func() {},
			nil,
			"invalid request",
		},
		{
			"fail - Empty delegator",
			func() {},
			&types.QueryLockedDelegationTotalRewardsRequest{
				DelegatorAddress: "",
			},
			"empty delegator address",
		},
		{
			"fail - bad delegator",
			func() {},
			&types.QueryLockedDelegationTotalRewardsRequest{
				DelegatorAddress: "test",
			},
			"invalid bech32 string",
		},
		{
			"pass",
			func() {
				// Send tokens and create a locked delegation
				err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(
					suite.ctx,
					types.ModuleName,
					addresses[0],
					sdk.NewCoins(
						sdk.NewCoin(bondDenom, sdk.NewInt(1_000_000)),
					))
				suite.Require().NoError(err)
				_, err = suite.app.LockingKeeper.CreateLockedDelegationEntryAndDelegate(
					suite.ctx,
					addresses[0],
					validator.GetOperator(),
					sdk.NewInt(1_000_000),
					types.DefaultRates[0].Duration,
					true,
				)
				suite.Require().NoError(err)

				// Allocate rewards to the validator
				// allocate some rewards
				initial := sdk.TokensFromConsensusPower(10, PowerReduction)
				tokens := sdk.DecCoins{sdk.NewDecCoin(bondDenom, initial)}
				suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, tokens)

				// next block
				suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
			},
			&types.QueryLockedDelegationTotalRewardsRequest{
				DelegatorAddress: addresses[0].String(),
			},
			"",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.maleate()

			res, err := suite.k.LockedDelegationTotalRewards(suite.ctx, tc.request)

			if tc.errContains == "" {
				suite.Require().NoError(err, tc.name)

				// Validate the response
				grandTotal := sdk.DecCoins{}
				for _, reward := range res.Rewards {
					// Check if the values are correct
					expectedLocked := reward.DistributionReward.MulDec(types.DefaultRates[0].Rate).QuoDec(math.LegacyNewDec(100))
					expected, _ := expectedLocked.TruncateDecimal()
					actual, _ := reward.LockingReward.TruncateDecimal()
					suite.Require().EqualValues(expected, actual, tc.name)

					// Check if the total is correct
					suite.Require().EqualValues(reward.DistributionReward.Add(reward.LockingReward...), res.Total, tc.name)

					// Add to the grand total
					grandTotal = grandTotal.Add(res.Total...)
				}

				// Grand total should be equal to the final total
				suite.Require().EqualValues(grandTotal, res.Total, tc.name)
			} else {
				suite.Require().ErrorContains(err, tc.errContains, tc.name)
			}
		})
	}
}

// TestStoreLockedDelegationQueryBadMarshal tests a bad marshal function by using a mocked marshal
func (suite *KeeperTestSuite) TestStoreLockedDelegationQueryBadMarshal() {
	// Start with a few addresses
	addresses, valAddresses := createLockedDelegations(suite)

	// Create the new codec on top of the old one
	mockAppCoded := tests.MockAppCoded{Codec: suite.app.AppCodec()}
	// Create a new locking keeper with the bad app Coded
	lockingKeeper := keeper.NewKeeper(
		suite.app.GetKey(types.StoreKey),
		mockAppCoded,
		suite.app.StakingKeeper,
		suite.app.DistrKeeper,
		suite.app.BankKeeper,
		authAddr,
	)

	// Try to fetch the DelegatorLockedDelegations
	_, err := lockingKeeper.DelegatorLockedDelegations(suite.ctx, &types.QueryDelegatorLockedDelegationsRequest{
		DelegatorAddr: addresses[0].String(),
	})
	suite.Require().Error(err)

	// Try to fetch the LockedDelegations
	_, err = lockingKeeper.LockedDelegations(suite.ctx, &types.QueryLockedDelegationRequest{
		DelegatorAddr: addresses[0].String(),
		ValidatorAddr: valAddresses[0].String(),
	})
	suite.Require().Error(err)
}

// createLockedDelegations sets up locked delegations for a predefined set of addresses and validators
func createLockedDelegations(suite *KeeperTestSuite) (addresses []sdk.AccAddress, valAddresses []sdk.ValAddress) {
	addresses = []sdk.AccAddress{
		sdk.AccAddress([]byte("address1")),
		sdk.AccAddress([]byte("address2")),
	}
	valAddresses = []sdk.ValAddress{
		sdk.ValAddress([]byte("val1")),
		sdk.ValAddress([]byte("val2")),
	}

	for _, addr := range addresses {
		for _, valAddr := range valAddresses {
			lockedDelegation := types.NewLockedDelegation(
				addr,
				valAddr,
				nil,
			)
			suite.Require().NoError(suite.k.SetLockedDelegation(suite.ctx, lockedDelegation))
		}
	}
	return
}
