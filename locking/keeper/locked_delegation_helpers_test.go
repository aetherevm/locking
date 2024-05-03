// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/aetherevm/locking/locking/types"
)

// NOTE: If migrated to a provider chain, add slashing tests

// CreateLockedDelegationEntryParams is a struct used to test the CreateLockedDelegationEntry
type CreateLockedDelegationEntryParams struct {
	DelAddr   sdk.AccAddress
	ValAddr   sdk.ValAddress
	Amount    math.Int
	Rate      types.Rate
	AutoRenew bool
}

// TestCreateLockedDelegationEntry tests CreateLockedDelegationEntry
func (suite *KeeperTestSuite) TestCreateLockedDelegationEntry() {
	delAddr := sdk.AccAddress([]byte("address1"))
	valAddr := sdk.ValAddress([]byte("val1"))

	rate := types.DefaultRates[0]

	testCases := []struct {
		name     string
		maleate  func() CreateLockedDelegationEntryParams
		expError bool
	}{
		{
			"fail - max entries reached",
			func() CreateLockedDelegationEntryParams {
				// Add a new entry to the existing ld
				// this will explode the max entries
				ld, found := suite.k.GetLockedDelegation(suite.ctx, delAddr, valAddr)
				suite.Require().True(found)
				entry := types.NewLockedDelegationEntry(
					math.LegacyOneDec(),
					rate,
					time.Now().Add(time.Hour),
					false,
					suite.k.IncrementLockedDelegationEntryID(suite.ctx),
				)
				ld.AddEntry(entry)
				err := suite.k.SetLockedDelegation(suite.ctx, ld)
				suite.Require().NoError(err)

				return CreateLockedDelegationEntryParams{
					DelAddr:   delAddr,
					ValAddr:   valAddr,
					Amount:    math.OneInt(),
					Rate:      entry.Rate,
					AutoRenew: entry.AutoRenew,
				}
			},
			true,
		},
		{
			"fail - invalid entry",
			func() CreateLockedDelegationEntryParams {
				// We force a invalid entry by zero shares and zero time
				return CreateLockedDelegationEntryParams{
					DelAddr:   delAddr,
					ValAddr:   valAddr,
					Amount:    math.ZeroInt(),
					Rate:      rate,
					AutoRenew: false,
				}
			},
			true,
		},
		{
			"fail - invalid locked delegation",
			func() CreateLockedDelegationEntryParams {
				// We force a invalid entry by zero shares and zero time
				return CreateLockedDelegationEntryParams{
					DelAddr:   sdk.AccAddress([]byte("")),
					ValAddr:   valAddr,
					Amount:    math.OneInt(),
					Rate:      rate,
					AutoRenew: false,
				}
			},
			true,
		},
		{
			"pass - entry created",
			func() CreateLockedDelegationEntryParams {
				return CreateLockedDelegationEntryParams{
					DelAddr:   delAddr,
					ValAddr:   valAddr,
					Amount:    math.OneInt(),
					Rate:      rate,
					AutoRenew: false,
				}
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// prepare testing locked delegation entries with 1 less the max entries
			maxEntries := suite.k.MaxEntries(suite.ctx)
			createLDWithEntries(delAddr, valAddr, int(maxEntries-1), rate, suite)

			// Maleate the system and get a target entry to be used on the tests
			params := tc.maleate()

			// Store the locked delegation and check for errors
			entry, err := suite.k.CreateLockedDelegationEntry(
				suite.ctx,
				params.DelAddr,
				params.ValAddr,
				params.Amount,
				params.Rate,
				params.AutoRenew,
			)

			if tc.expError {
				suite.Require().Error(err, tc.name)
			} else {
				suite.Require().NoError(err, tc.name)

				// The ld must be updated, let's check the total num of entries
				finalLd, found := suite.k.GetLockedDelegation(suite.ctx, delAddr, valAddr)
				suite.Require().True(found, tc.name)
				suite.Require().EqualValues(len(finalLd.Entries), maxEntries)

				// A new entry also should be on the queue
				pairs := suite.k.GetAllLockedDelegationQueuePairs(
					suite.ctx,
					bigTime,
				)
				suite.Require().NotEmpty(pairs, tc.name)

				// We also can find the new entry on the ID lookup
				_, found = suite.k.GetLockedDelegationByEntryID(suite.ctx, entry.Id)
				suite.Require().True(found)
			}
		})
	}
}

// LockedDelegationRedelegationParams is a struct use for testing the LockedDelegationRedelegation
type LockedDelegationRedelegationParams struct {
	DelAddr    sdk.AccAddress
	ValSrcAddr sdk.ValAddress
	ValDstAddr sdk.ValAddress
	Ids        []uint64
}

// TestLockedDelegationRedelegation tests LockedDelegationRedelegation
// this test only consider a empty id list
func (suite *KeeperTestSuite) TestLockedDelegationRedelegation() {
	delAddr := sdk.AccAddress([]byte("address1"))
	srcValAddr := sdk.ValAddress([]byte("val1"))
	dstValAddr := sdk.ValAddress([]byte("val2"))

	rate := types.DefaultRates[0]
	maxEntries := suite.k.MaxEntries(suite.ctx)

	testCases := []struct {
		name        string
		maleate     func(validator stakingtypes.Validator) LockedDelegationRedelegationParams
		errContains string
	}{
		{
			name: "fail - will reach max entries",
			maleate: func(validator stakingtypes.Validator) LockedDelegationRedelegationParams {
				// Create some entries on the target validator
				createLDWithEntries(delAddr, dstValAddr, int(maxEntries), rate, suite)
				return LockedDelegationRedelegationParams{
					DelAddr:    delAddr,
					ValSrcAddr: srcValAddr,
					ValDstAddr: dstValAddr,
				}
			},
			errContains: "will reach the max entries",
		},
		{
			name: "fail - invalid entry",
			maleate: func(validator stakingtypes.Validator) LockedDelegationRedelegationParams {
				// We return the redelegation from val2 to val1
				// since val2 has no locked delegations it will error out
				return LockedDelegationRedelegationParams{
					DelAddr:    delAddr,
					ValSrcAddr: dstValAddr,
					ValDstAddr: srcValAddr,
				}
			},
			errContains: "pair not found",
		},
		{
			name: "fail - bad target address",
			maleate: func(validator stakingtypes.Validator) LockedDelegationRedelegationParams {
				return LockedDelegationRedelegationParams{
					DelAddr:    delAddr,
					ValSrcAddr: srcValAddr,
					ValDstAddr: sdk.ValAddress([]byte("")),
				}
			},
			errContains: "no delegation",
		},
		{
			name: "fail - can't redelegate zero tokens",
			maleate: func(validator stakingtypes.Validator) LockedDelegationRedelegationParams {
				return LockedDelegationRedelegationParams{
					DelAddr:    delAddr,
					ValSrcAddr: sdk.ValAddress([]byte("val3")),
					ValDstAddr: dstValAddr,
				}
			},
			errContains: "pair not found",
		},
		{
			name: "fail - redelegation with bad id",
			maleate: func(validator stakingtypes.Validator) LockedDelegationRedelegationParams {
				mintAndDelegate(suite, delAddr, validator)
				return LockedDelegationRedelegationParams{
					DelAddr:    delAddr,
					ValSrcAddr: validator.GetOperator(),
					ValDstAddr: dstValAddr,
					Ids:        []uint64{123, 154, 13},
				}
			},
			errContains: "locked delegation entry for specified id not found",
		},
		{
			name: "pass - nil id list",
			maleate: func(validator stakingtypes.Validator) LockedDelegationRedelegationParams {
				mintAndDelegate(suite, delAddr, validator)
				return LockedDelegationRedelegationParams{
					DelAddr:    delAddr,
					ValSrcAddr: validator.GetOperator(),
					ValDstAddr: dstValAddr,
				}
			},
			errContains: "",
		},
		{
			name: "pass - redelegation with valid id list",
			maleate: func(validator stakingtypes.Validator) LockedDelegationRedelegationParams {
				mintAndDelegate(suite, delAddr, validator)
				return LockedDelegationRedelegationParams{
					DelAddr:    delAddr,
					ValSrcAddr: validator.GetOperator(),
					ValDstAddr: dstValAddr,
					Ids:        []uint64{2, 3, 6, 50},
				}
			},
			errContains: "",
		},
		{
			name: "pass - redelegation with valid id list and duplicated items",
			maleate: func(validator stakingtypes.Validator) LockedDelegationRedelegationParams {
				mintAndDelegate(suite, delAddr, validator)
				return LockedDelegationRedelegationParams{
					DelAddr:    delAddr,
					ValSrcAddr: validator.GetOperator(),
					ValDstAddr: dstValAddr,
					Ids:        []uint64{4, 10, 40, 51, 1, 45, 6, 7, 7, 7, 7, 7},
				}
			},
			errContains: "",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Restart the system
			suite.SetupTest()

			// prepare testing locked delegation entries with 1 less the max entries
			validator := suite.app.StakingKeeper.GetAllValidators(suite.ctx)[0]
			srcValAddr = validator.GetOperator()
			createLDWithEntries(delAddr, srcValAddr, int(maxEntries-1), rate, suite)

			// Store the SRC LD
			srcLD, found := suite.k.GetLockedDelegation(suite.ctx, delAddr, srcValAddr)
			suite.Require().True(found, tc.name)

			// Maleate the system and get a target entry to be used on the tests
			params := tc.maleate(validator)

			// Store the locked delegation and check for errors
			movedShares, _, err := suite.k.LockedDelegationRedelegation(
				suite.ctx,
				params.DelAddr,
				params.ValSrcAddr,
				params.ValDstAddr,
				params.Ids,
			)

			if tc.errContains != "" {
				suite.Require().ErrorContains(err, tc.errContains, tc.name)
			} else {
				suite.Require().NoError(err, tc.name)

				// Find all the entries for the specified ids
				exists, foundEntries := srcLD.EntriesForIds(params.Ids)
				suite.Require().True(exists)

				// Calculate the total shares from the found entries, it will be used in further checks
				// The moved shares should be the equal to the total from found entries
				totalFromFoundEntries := math.LegacyZeroDec()
				for _, fEntry := range foundEntries {
					totalFromFoundEntries = totalFromFoundEntries.Add(fEntry.Shares)
				}
				suite.Require().EqualValues(totalFromFoundEntries, movedShares, tc.name)

				// Check if the srcValAddr has been removed if all was redelegated
				// or if not, only that specific entries should not exists
				finalSrcLD, found := suite.k.GetLockedDelegation(suite.ctx, delAddr, srcValAddr)
				if len(params.Ids) == 0 {
					// The source ld should not be found
					suite.Require().False(found, tc.name)
					// The moved shares must be the total shares from the srcLD
					suite.Require().EqualValues(srcLD.TotalShares(), movedShares, tc.name)
				} else {
					suite.Require().NotContains(finalSrcLD.Entries, foundEntries, tc.name)
					suite.Require().EqualValues(srcLD.TotalShares().Sub(totalFromFoundEntries), finalSrcLD.TotalShares(), tc.name)
				}

				// The destination ld should exist and hold all new entries
				dstLD, found := suite.k.GetLockedDelegation(suite.ctx, delAddr, dstValAddr)
				suite.Require().True(found, tc.name)
				suite.Require().Equal(len(foundEntries), len(dstLD.Entries), tc.name)

				// The queue also should hold the new delegations
				pairs := suite.k.GetAllLockedDelegationQueuePairs(
					suite.ctx,
					bigTime,
				)
				var dstPairs []types.LockedDelegationPair
				for _, pair := range pairs {
					if pair.ValidatorAddress == dstValAddr.String() {
						dstPairs = append(dstPairs, pair)
					}
				}

				// Same amount of entries and queue items
				suite.Require().EqualValues(len(dstPairs), len(foundEntries), tc.name)

				// All the old entries must still exists on the lookup, but pointing to the new validator
				for _, entry := range foundEntries {
					ldPerEntryID, found := suite.k.GetLockedDelegationByEntryID(suite.ctx, entry.Id)
					suite.Require().True(found, tc.name)
					suite.Require().Equal(dstValAddr.String(), ldPerEntryID.ValidatorAddress, tc.name)
				}

				// Now all the new entries must exists on the lookup
				for _, entry := range dstLD.Entries {
					foundLD, found := suite.k.GetLockedDelegationByEntryID(suite.ctx, entry.Id)
					suite.Require().True(found, tc.name)
					suite.Require().Equal(dstLD, foundLD, tc.name)
				}
			}
		})
	}
}

// ToggleLockedDelegationEntryAutoRenewParams is a struct use for testing the ToggleLockedDelegationEntryAutoRenew
type ToggleLockedDelegationEntryAutoRenewParams struct {
	DelAddr sdk.AccAddress
	ValAddr sdk.ValAddress
	EntryID uint64
}

// TestToggleLockedDelegationEntryAutoRenew tests ToggleLockedDelegationEntryAutoRenew
func (suite *KeeperTestSuite) TestToggleLockedDelegationEntryAutoRenew() {
	delAddr := sdk.AccAddress([]byte("address1"))
	valAddr := sdk.ValAddress([]byte("val1"))

	rate := types.DefaultRates[0]

	testCases := []struct {
		name     string
		params   ToggleLockedDelegationEntryAutoRenewParams
		expError bool
	}{
		{
			"fail - no locked delegation found",
			ToggleLockedDelegationEntryAutoRenewParams{
				DelAddr: delAddr,
				ValAddr: sdk.ValAddress([]byte("val00001")),
				EntryID: 0,
			},
			true,
		},
		{
			"fail - entry not found",
			ToggleLockedDelegationEntryAutoRenewParams{
				DelAddr: delAddr,
				ValAddr: valAddr,
				EntryID: 1<<64 - 1,
			},
			true,
		},
		{
			"success - entry updated",
			ToggleLockedDelegationEntryAutoRenewParams{
				DelAddr: delAddr,
				ValAddr: valAddr,
				EntryID: 1,
			},
			false,
		},
		{
			"success - entry updated case 2",
			ToggleLockedDelegationEntryAutoRenewParams{
				DelAddr: delAddr,
				ValAddr: valAddr,
				EntryID: 2,
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Restart the system each time to keep the ids
			suite.SetupTest()

			// Prepare testing locked delegation entries with 1 less the max entries
			maxEntries := suite.k.MaxEntries(suite.ctx)
			createLDWithEntries(delAddr, valAddr, int(maxEntries-1), rate, suite)

			// Fetch the original locked delegation
			originalLockedDelegation, _ := suite.k.GetLockedDelegation(
				suite.ctx,
				tc.params.DelAddr,
				tc.params.ValAddr,
			)

			// Do the toggle
			_, err := suite.k.ToggleLockedDelegationEntryAutoRenew(
				suite.ctx,
				tc.params.DelAddr,
				tc.params.ValAddr,
				tc.params.EntryID,
			)

			if tc.expError {
				suite.Require().Error(err, tc.name)
			} else {
				suite.Require().NoError(err, tc.name)

				// Fetch the locked delegation from the database
				lockedDelegation, found := suite.k.GetLockedDelegation(
					suite.ctx,
					tc.params.DelAddr,
					tc.params.ValAddr,
				)
				suite.Require().True(found, tc.name)

				// Now check if the entry was really update
				// Other checks are assured by tests on x/locking/types/locked_delegation_test.go
				for i, entry := range lockedDelegation.Entries {
					originalEntry := originalLockedDelegation.Entries[i]
					if entry.Id == tc.params.EntryID {
						suite.Require().NotEqual(entry.AutoRenew, originalEntry.AutoRenew, tc.name)
						continue
					}
					suite.Require().Equal(entry, originalEntry, tc.name)
				}
			}
		})
	}
}

// TestHasMaxLockedDelegationEntries tests HasMaxLockedDelegationEntries
func (suite *KeeperTestSuite) TestHasMaxLockedDelegationEntries() {
	delAddr := sdk.AccAddress([]byte("address1"))
	valAddr := sdk.ValAddress([]byte("val1"))

	rate := types.DefaultRates[0]

	testCases := []struct {
		name     string
		maleate  func()
		delAddr  sdk.AccAddress
		valAddr  sdk.ValAddress
		hasMaxLD bool
	}{
		{
			"not max - default ld",
			func() {},
			delAddr,
			valAddr,
			false,
		},
		{
			"not max - ld not found (by delegator)",
			func() {},
			sdk.AccAddress([]byte("test")),
			valAddr,
			false,
		},
		{
			"not max - ld not found (by validator)",
			func() {},
			delAddr,
			sdk.ValAddress([]byte("test")),
			false,
		},
		{
			"max - max entries reached",
			func() {
				// We add a new entry to the LD
				// The current LD is about to reach max
				_, err := suite.k.CreateLockedDelegationEntry(
					suite.ctx,
					delAddr,
					valAddr,
					math.OneInt(),
					rate,
					false,
				)
				suite.Require().NoError(err)
			},
			delAddr,
			valAddr,
			true,
		},
		{
			"max - max entries surpassed",
			func() {
				lockedDelegation, found := suite.k.GetLockedDelegation(suite.ctx, delAddr, valAddr)
				suite.Require().True(found)
				// Create extra entries
				for i := 0; i < 10; i++ {
					lockedDelegation.AddEntry(
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(int64(i)+1),
							rate,
							time.Now().Add(time.Duration(i)*time.Duration(i)),
							false,
							uint64(i),
						),
					)
				}
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)
			},
			delAddr,
			valAddr,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// prepare testing locked delegation entries with 1 less the max entries
			maxEntries := suite.k.MaxEntries(suite.ctx)
			createLDWithEntries(delAddr, valAddr, int(maxEntries-1), rate, suite)

			// Maleate the system and get a target entry to be used on the tests
			tc.maleate()

			// Store the locked delegation and check for errors
			res := suite.k.HasMaxLockedDelegationEntries(
				suite.ctx,
				tc.delAddr,
				tc.valAddr,
			)

			if tc.hasMaxLD {
				suite.Require().True(res, tc.name)

				// Also check the max entries param
				maxEntries := suite.k.MaxEntries(suite.ctx)
				lockedDelegation, found := suite.k.GetLockedDelegation(suite.ctx, tc.delAddr, tc.valAddr)
				suite.Require().True(found, tc.name)
				suite.Require().GreaterOrEqual(uint32(len(lockedDelegation.Entries)), maxEntries)
			} else {
				suite.Require().False(res, tc.name)
			}
		})
	}
}

// TestLDRedelegationWillReachMaxEntries tests LDRedelegationWillReachMaxEntries
func (suite *KeeperTestSuite) TestLDRedelegationWillReachMaxEntries() {
	// Set a few addresses for testing
	delAddr := sdk.AccAddress([]byte("address1"))
	srcValAddr := sdk.ValAddress([]byte("val1"))
	dstValAddr := sdk.ValAddress([]byte("val2"))

	// Default rate and max entries
	rate := types.DefaultRates[0]
	maxEntries := suite.k.MaxEntries(suite.ctx)

	testCases := []struct {
		name       string
		setter     func()
		delAddr    sdk.AccAddress
		srcValAddr sdk.ValAddress
		dstValAddr sdk.ValAddress
		hasMaxLD   bool
	}{
		{
			"not max - no source locked delegation found",
			func() {},
			delAddr,
			srcValAddr,
			dstValAddr,
			false,
		},
		{
			"not max - no target locked delegation found",
			func() {
				// Create a few entries on top of the source validator
				createLDWithEntries(delAddr, srcValAddr, int(maxEntries-1), rate, suite)
			},
			delAddr,
			srcValAddr,
			dstValAddr,
			false,
		},
		{
			"not max - source entries and target entries sum less than max entries",
			func() {
				// Create a 2 entries on top of the source validator
				createLDWithEntries(delAddr, srcValAddr, 2, rate, suite)
				// Create a 2 entries on top of the target validator
				createLDWithEntries(delAddr, dstValAddr, 2, rate, suite)
			},
			delAddr,
			srcValAddr,
			dstValAddr,
			false,
		},
		{
			"not max - source entries and target entries sum 1 less than max entries",
			func() {
				// Create a max entries on top of the source validator
				createLDWithEntries(delAddr, srcValAddr, int((maxEntries/2)-1), rate, suite)
				// Create a 2 entries on top of the target validator
				createLDWithEntries(delAddr, dstValAddr, int(maxEntries/2), rate, suite)
			},
			delAddr,
			srcValAddr,
			dstValAddr,
			false,
		},
		{
			"not max - source entries and target with diff entries sum 1 less than max entries",
			func() {
				// Create max entries less 2 on top of the source validator
				createLDWithEntries(delAddr, srcValAddr, int(maxEntries-2), rate, suite)
				// Create 1 entry on top of the target validator
				createLDWithEntries(delAddr, dstValAddr, 1, rate, suite)
			},
			delAddr,
			srcValAddr,
			dstValAddr,
			false,
		},
		{
			"not max - source entries and target with diff entries sum 1 less than max entries - bigger dst",
			func() {
				// Create 1 entry on top of the source validator
				createLDWithEntries(delAddr, srcValAddr, 1, rate, suite)
				// Create max entries less 2 on top of the target validator
				createLDWithEntries(delAddr, dstValAddr, int(maxEntries-2), rate, suite)
			},
			delAddr,
			srcValAddr,
			dstValAddr,
			false,
		},
		{
			"max - source entries and target entries sum equal to max entries",
			func() {
				// Create half max entries on top of the source validator
				createLDWithEntries(delAddr, srcValAddr, int(maxEntries/2), rate, suite)
				// Create half max entries on top of the target validator
				createLDWithEntries(delAddr, dstValAddr, int(maxEntries/2), rate, suite)
			},
			delAddr,
			srcValAddr,
			dstValAddr,
			true,
		},
		{
			"max - source entries and target entries sum bigger than max entries",
			func() {
				// Create 1000 entries on top of the source validator
				createLDWithEntries(delAddr, srcValAddr, int(1000), rate, suite)
				// Create 1000 entries on top of the target validator
				createLDWithEntries(delAddr, dstValAddr, int(1000), rate, suite)
			},
			delAddr,
			srcValAddr,
			dstValAddr,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Restart the chain
			suite.SetupTest()

			// Set the system
			tc.setter()

			// Store the locked delegation and check for errors
			res := suite.k.LDRedelegationWillReachMaxEntries(
				suite.ctx,
				tc.delAddr,
				tc.srcValAddr,
				tc.dstValAddr,
			)

			if tc.hasMaxLD {
				suite.Require().True(res, tc.name)
			} else {
				suite.Require().False(res, tc.name)
			}
		})
	}
}

// TestLockedDelegationTotalShares tests LockedDelegationTotalShares
func (suite *KeeperTestSuite) TestLockedDelegationTotalShares() {
	delAddr := sdk.AccAddress([]byte("address1"))
	valAddr := sdk.ValAddress([]byte("val1"))

	rate := types.DefaultRates[0]

	//  For this test we create our own set of entries
	ld := types.NewLockedDelegation(delAddr, valAddr, nil)
	for i := 0; i < 5; i++ {
		ld.AddEntry(
			types.NewLockedDelegationEntry(
				math.LegacyNewDec(int64(10)),
				rate,
				time.Time{}.Add(time.Hour*time.Duration(i+1)),
				false,
				uint64(i),
			),
		)
	}
	err := suite.k.SetLockedDelegation(suite.ctx, ld)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		maleate     func()
		delAddr     sdk.AccAddress
		valAddr     sdk.ValAddress
		totalShares math.LegacyDec
	}{
		{
			"locked delegation not found (by delegator)",
			func() {},
			sdk.AccAddress([]byte("test")),
			valAddr,
			sdk.ZeroDec(),
		},
		{
			"locked delegation not found (by validator val)",
			func() {},
			sdk.AccAddress([]byte("test")),
			valAddr,
			sdk.ZeroDec(),
		},
		{
			"get the sum",
			func() {},
			delAddr,
			valAddr,
			math.LegacyNewDec(10 * 5),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Maleate the system and get a target entry to be used on the tests
			tc.maleate()

			res := suite.k.LockedDelegationTotalShares(suite.ctx, tc.delAddr, tc.valAddr)

			suite.Require().Equal(tc.totalShares, res, tc.name)
		})
	}
}

// TestGetDelegatorLockedShares tests GetDelegatorLockedShares
func (suite *KeeperTestSuite) TestGetDelegatorLockedShares() {
	rate := types.DefaultRates[0]

	delAddresses := createMultipleLDsWithEntries(3, rate, suite)

	testCases := []struct {
		name        string
		maleate     func()
		delAddr     sdk.AccAddress
		totalShares math.LegacyDec
	}{
		{
			"locked delegation not found",
			func() {},
			sdk.AccAddress([]byte("test")),
			math.LegacyZeroDec(),
		},
		{
			"locked delegation found",
			func() {},
			delAddresses[0],
			math.LegacyNewDec(1 + 2 + 3),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Maleate the system and get a target entry to be used on the tests
			tc.maleate()

			res := suite.k.GetDelegatorLockedShares(suite.ctx, tc.delAddr)

			suite.Require().Equal(tc.totalShares, res, tc.name)
		})
	}
}

// TestDequeueExpiredLockedDelegations tests DequeueExpiredLockedDelegations
func (suite *KeeperTestSuite) TestDequeueExpiredLockedDelegations() {
	now := time.Now()

	// Create a few locked delegations with entries
	rate := types.DefaultRates[0]
	_ = createMultipleLDsWithEntries(5, rate, suite)

	// Pairs should not be empty
	allPairs := suite.k.GetAllLockedDelegationQueuePairs(suite.ctx, bigTime)
	suite.Require().NotEmpty(allPairs)

	// Do the dequeuing and get the expired pairs
	// This is the function we want to test
	expiredPairs := suite.k.DequeueExpiredLockedDelegations(suite.ctx, now)
	suite.Require().NotEmpty(expiredPairs)

	// Check if they are really out of the queue
	// We can do this by fetching all pairs in the queue based on now
	currPairs := suite.k.GetAllLockedDelegationQueuePairs(suite.ctx, now)
	suite.Require().Empty(currPairs)

	// The remaining items on the queue should be the total - len of expiredPairs
	afterDequeueAllPairs := suite.k.GetAllLockedDelegationQueuePairs(suite.ctx, bigTime)
	suite.Require().Equal(len(afterDequeueAllPairs), len(allPairs)-len(expiredPairs))

	// Check if the expiredPairs are really what we mean them to be
	lds := suite.k.GetAllLockedDelegations(suite.ctx)
	var expectedPairs []types.LockedDelegationPair
	for _, ld := range lds {
		for _, entry := range ld.Entries {
			if entry.Expired(now) {
				expectedPairs = append(expectedPairs, types.LockedDelegationPair{
					DelegatorAddress: ld.DelegatorAddress,
					ValidatorAddress: ld.ValidatorAddress,
				})
			}
		}
	}
	suite.Require().ElementsMatch(expectedPairs, expectedPairs)
}

// TestDequeueExpiredLockedDelegationsEmpty tests a empty path on the DequeueExpiredLockedDelegations function
func (suite *KeeperTestSuite) TestDequeueExpiredLockedDelegationsEmpty() {
	now := time.Now()

	// Call in a system with empty expired
	expiredPairs := suite.k.DequeueExpiredLockedDelegations(suite.ctx, now)
	suite.Require().Empty(expiredPairs)
}

// TestDequeueExpiredLockedDelegationsEmpty tests a empty path on the DequeueExpiredLockedDelegations function
// Test the capability of expiring the queue and only passing forward one unique pair
func (suite *KeeperTestSuite) TestDequeueExpiredLockedDelegationsMultiplePairs() {
	now := time.Now()

	// Add the same pair multiple times on the same queue
	delAddr := sdk.AccAddress([]byte("address1"))
	valAddr := sdk.ValAddress([]byte("val1"))
	ld := types.NewLockedDelegation(
		delAddr,
		valAddr,
		nil,
	)
	suite.k.InsertLockedDelegationQueue(suite.ctx, ld, now)
	suite.k.InsertLockedDelegationQueue(suite.ctx, ld, now)
	suite.k.InsertLockedDelegationQueue(suite.ctx, ld, now)

	// We will have a single pair
	expiredPairs := suite.k.DequeueExpiredLockedDelegations(suite.ctx, now)
	suite.Require().NotEmpty(expiredPairs)
	suite.Require().ElementsMatch(expiredPairs, []types.LockedDelegationPair{
		{
			DelegatorAddress: delAddr.String(),
			ValidatorAddress: valAddr.String(),
		},
	})
}

// TestCompleteLockedDelegations tests CompleteLockedDelegations
// and consequently the private functions processEntries and handleAutoRenew
func (suite *KeeperTestSuite) TestCompleteLockedDelegations() {
	delegatedTokens := sdk.TokensFromConsensusPower(1_000_000, PowerReduction)
	delAddr := sdk.AccAddress([]byte("address1"))

	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	powerReduction := sdk.NewDecFromInt(suite.app.StakingKeeper.PowerReduction(suite.ctx))

	rate := types.DefaultRates[0]

	testCases := []struct {
		name         string
		maleate      func(valAddr sdk.ValAddress) types.LockedDelegationPair
		sharesChange math.Int
		expError     bool
	}{
		{
			"fail - bad validator address",
			func(valAddr sdk.ValAddress) types.LockedDelegationPair {
				return types.LockedDelegationPair{
					DelegatorAddress: delAddr.String(),
					ValidatorAddress: "test",
				}
			},
			math.ZeroInt(),
			true,
		},
		{
			"pass - locked delegation not found",
			func(valAddr sdk.ValAddress) types.LockedDelegationPair {
				return types.LockedDelegationPair{
					DelegatorAddress: delAddr.String(),
					ValidatorAddress: valAddr.String(),
				}
			},
			math.ZeroInt(),
			false,
		},
		{
			"pass - locked delegation not found",
			func(valAddr sdk.ValAddress) types.LockedDelegationPair {
				return types.LockedDelegationPair{
					DelegatorAddress: delAddr.String(),
					ValidatorAddress: valAddr.String(),
				}
			},
			math.ZeroInt(),
			false,
		},
		{
			"pass - no expired entry",
			func(valAddr sdk.ValAddress) types.LockedDelegationPair {
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					valAddr,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(30),
							rate,
							time.Now().Add(time.Hour), // in the future
							false,
							1,
						),
					},
				)
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				return types.LockedDelegationPair{
					DelegatorAddress: delAddr.String(),
					ValidatorAddress: valAddr.String(),
				}
			},
			math.NewInt(0),
			false,
		},
		{
			"pass - undelegate",
			func(valAddr sdk.ValAddress) types.LockedDelegationPair {
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					valAddr,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(30),
							rate,
							time.Now().Add(time.Hour*-1), // in the past
							false,
							1,
						),
					},
				)
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				return types.LockedDelegationPair{
					DelegatorAddress: delAddr.String(),
					ValidatorAddress: valAddr.String(),
				}
			},
			math.NewInt(30),
			false,
		},
		{
			"pass - multiple undelegate",
			func(valAddr sdk.ValAddress) types.LockedDelegationPair {
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					valAddr,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(30),
							rate,
							time.Now().Add(time.Hour*-1), // in the past
							false,
							1,
						),
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(31),
							rate,
							time.Now().Add(time.Hour*-2), // in the past
							false,
							2,
						),
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(190),
							rate,
							time.Now().Add(time.Hour*2), // in the future
							false,
							3,
						),
					},
				)
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				return types.LockedDelegationPair{
					DelegatorAddress: delAddr.String(),
					ValidatorAddress: valAddr.String(),
				}
			},
			math.NewInt(30 + 31),
			false,
		},
		{
			"fail - exceeding undelegate",
			func(valAddr sdk.ValAddress) types.LockedDelegationPair {
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					valAddr,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(30),
							rate,
							time.Now().Add(time.Hour*-1),
							false,
							1,
						),
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(2900).Power(10),
							rate,
							time.Now().Add(time.Hour*-2),
							false,
							2,
						),
					},
				)
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				return types.LockedDelegationPair{
					DelegatorAddress: delAddr.String(),
					ValidatorAddress: valAddr.String(),
				}
			},
			math.NewInt(1900 + 30),
			true,
		},
		{
			"pass - renew",
			func(valAddr sdk.ValAddress) types.LockedDelegationPair {
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					valAddr,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(30),
							rate,
							time.Now().Add(time.Hour*-1), // in the past
							true,
							suite.k.IncrementLockedDelegationEntryID(suite.ctx),
						),
					},
				)
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)
				return types.LockedDelegationPair{
					DelegatorAddress: delAddr.String(),
					ValidatorAddress: valAddr.String(),
				}
			},
			// Shares should be unchanged
			math.NewInt(0),
			false,
		},
		{
			"pass - multiple renews",
			func(valAddr sdk.ValAddress) types.LockedDelegationPair {
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					valAddr,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(30),
							rate,
							time.Now().Add(time.Hour*-1), // in the past
							true,
							suite.k.IncrementLockedDelegationEntryID(suite.ctx),
						),
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(31),
							rate,
							time.Now().Add(time.Hour), // in the future
							true,
							suite.k.IncrementLockedDelegationEntryID(suite.ctx),
						),
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(32),
							rate,
							time.Now().Add(time.Hour*-2), // in the past
							true,
							suite.k.IncrementLockedDelegationEntryID(suite.ctx),
						),
					},
				)
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				return types.LockedDelegationPair{
					DelegatorAddress: delAddr.String(),
					ValidatorAddress: valAddr.String(),
				}
			},
			// Shares should be unchanged
			math.NewInt(0),
			false,
		},
		{
			"pass - multiple renews and undelegations",
			func(valAddr sdk.ValAddress) types.LockedDelegationPair {
				lockedDelegation := types.NewLockedDelegation(
					delAddr,
					valAddr,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(30),
							rate,
							time.Now().Add(time.Hour*-1), // in the past
							false,
							suite.k.IncrementLockedDelegationEntryID(suite.ctx),
						),
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(31),
							rate,
							time.Now().Add(time.Hour*-2), // in the past
							false,
							suite.k.IncrementLockedDelegationEntryID(suite.ctx),
						),
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(190),
							rate,
							time.Now().Add(time.Hour*3), // in the future
							false,
							suite.k.IncrementLockedDelegationEntryID(suite.ctx),
						),
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(30),
							rate,
							time.Now().Add(time.Hour*-1), // in the past
							true,
							suite.k.IncrementLockedDelegationEntryID(suite.ctx),
						),
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(31),
							rate,
							time.Now().Add(time.Hour), // in the future
							true,
							suite.k.IncrementLockedDelegationEntryID(suite.ctx),
						),
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(32),
							rate,
							time.Now().Add(time.Hour*-2), // in the past
							true,
							suite.k.IncrementLockedDelegationEntryID(suite.ctx),
						),
					},
				)
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				return types.LockedDelegationPair{
					DelegatorAddress: delAddr.String(),
					ValidatorAddress: valAddr.String(),
				}
			},
			// Only two are expired and with undelegate
			math.NewInt(30 + 31),
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // Restart the whole app each time
			validator := suite.app.StakingKeeper.GetAllValidators(suite.ctx)[0]
			valAddr := validator.GetOperator()

			// To fully test these we must create delegations to a existing validator
			// In this snippet we are delegating 200 tokens to the first validator
			err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
				sdk.NewCoin(bondDenom, delegatedTokens),
			))
			suite.Require().NoError(err)
			_, err = suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, delegatedTokens, stakingtypes.Unbonded, validator, true)
			suite.Require().NoError(err)

			pair := tc.maleate(valAddr)

			// Fetch the old self of locked delegations for validation
			oldLockedDelegation, _ := suite.k.GetLockedDelegation(suite.ctx, delAddr, valAddr)

			err = suite.k.CompleteLockedDelegations(suite.ctx, pair)

			delegation := suite.app.StakingKeeper.Delegation(suite.ctx, delAddr, valAddr)
			delegatedShares := types.CalculateSharesFromValidator(delegatedTokens, validator)
			if tc.expError {
				suite.Require().Error(err, tc.name)

				// Shares should remain untouched
				newShares := sdk.NewDecFromInt(delegatedTokens).Quo(powerReduction)
				suite.Require().Equal(newShares, delegation.GetShares())

				// lockdedDelegation also should be untouched
				newLockedDelegation, _ := suite.k.GetLockedDelegation(suite.ctx, delAddr, valAddr)
				suite.Require().ElementsMatch(newLockedDelegation.Entries, oldLockedDelegation.Entries, tc.name)
			} else {
				suite.Require().NoError(err, tc.name)

				// Check the delegation value
				newShares := delegatedShares.Sub(math.LegacyNewDecFromInt(tc.sharesChange))
				suite.Require().EqualValues(newShares, delegation.GetShares(), tc.name)

				// The locked delegation should only exists if it had at least one entry after the complete
				// This way we can check if the entry was really removed
				newLockedDelegation, found := suite.k.GetLockedDelegation(suite.ctx, delAddr, valAddr)
				if len(oldLockedDelegation.Entries) == 1 &&
					!oldLockedDelegation.Entries[0].AutoRenew &&
					oldLockedDelegation.Entries[0].Expired(suite.ctx.BlockTime()) {
					suite.Require().False(found, tc.name)
					return
				}

				// Check all the expired entries, they should not exists on the original list
				var expiredEntries []types.LockedDelegationEntry
				for _, entry := range oldLockedDelegation.Entries {
					if entry.Expired(suite.ctx.BlockTime()) {
						suite.Require().NotContains(newLockedDelegation.Entries, entry, tc.name)
						expiredEntries = append(expiredEntries, entry)
					}
				}

				// For each expired entries that are renewed, we should have a newer counterpart
				for _, expiredEntry := range expiredEntries {
					if expiredEntry.AutoRenew {
						// All old IDs must exist on lookup
						_, found := suite.k.GetLockedDelegationByEntryID(suite.ctx, expiredEntry.Id)
						suite.Require().True(found, tc.name)

						// A new entry with the correct timestamp must have been added to the list
						var entry types.LockedDelegationEntry
						for _, newEntry := range newLockedDelegation.Entries {
							if expiredEntry.UnlockOn.Add(expiredEntry.Rate.Duration).Equal(newEntry.UnlockOn) {
								entry = newEntry
							}
						}

						// All values must be equal besides unlock on
						counterpartEntry := types.NewLockedDelegationEntry(
							expiredEntry.Shares,
							expiredEntry.Rate,
							expiredEntry.UnlockOn.Add(expiredEntry.Rate.Duration), // new duration,
							expiredEntry.AutoRenew,
							expiredEntry.Id,
						)
						suite.Require().Equal(entry, counterpartEntry, tc.name)
					} else {
						// If not renew, it should not exist on the lookup
						_, found := suite.k.GetLockedDelegationByEntryID(suite.ctx, expiredEntry.Id)
						suite.Require().False(found, tc.name)
					}
				}
			}
		})
	}
}

// createLDWithEntries creates a new locked delegation for testing
// the locked delegation is created with maxEntries - 1 as number of entries
func createLDWithEntries(delAddr sdk.AccAddress, valAddr sdk.ValAddress, nEntries int, rate types.Rate, suite *KeeperTestSuite) {
	ld := types.NewLockedDelegation(delAddr, valAddr, nil)
	for i := 0; i < nEntries; i++ {
		newEntry := types.NewLockedDelegationEntry(
			math.LegacyNewDec(int64(i+1)),
			rate,
			shuffledTimestamp(i),
			i%2 == 0,
			suite.k.IncrementLockedDelegationEntryID(suite.ctx),
		)
		ld.AddEntry(newEntry)
		// Add to queue
		suite.k.InsertLockedDelegationQueue(suite.ctx, ld, newEntry.UnlockOn)
	}
	err := suite.k.SetLockedDelegation(suite.ctx, ld)
	suite.Require().NoError(err)
}

// createMultipleLDsWithEntries is a helper function that can create locked delegations for
// multiple delegators and validators
func createMultipleLDsWithEntries(entries int, rate types.Rate, suite *KeeperTestSuite) []sdk.AccAddress {
	delAddresses := []sdk.AccAddress{
		sdk.AccAddress([]byte("del1")),
		sdk.AccAddress([]byte("del2")),
	}

	valAddresses := []sdk.ValAddress{
		sdk.ValAddress([]byte("val1")),
		sdk.ValAddress([]byte("val2")),
		sdk.ValAddress([]byte("val3")),
	}

	// Create entries
	for _, del := range delAddresses {
		for _, val := range valAddresses {
			createLDWithEntries(del, val, entries, rate, suite)
		}
	}

	return delAddresses
}

// mintAndDelegate is a help function to mint new tokens from power and delegate to a validator
func mintAndDelegate(suite *KeeperTestSuite, delAddr sdk.AccAddress, validator stakingtypes.Validator) {
	denom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	tokens := sdk.TokensFromConsensusPower(1_000_000, PowerReduction)
	err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, delAddr, sdk.NewCoins(
		sdk.NewCoin(denom, tokens),
	))
	suite.Require().NoError(err)

	// Delegate
	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, delAddr, tokens, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)
}
