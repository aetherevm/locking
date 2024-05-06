package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aetherevm/locking/locking/keeper"
	"github.com/aetherevm/locking/locking/tests"
	"github.com/aetherevm/locking/locking/types"
)

// TestIncrementLockedDelegationEntryID tests the store of locked delegation entry ids
func (suite *KeeperTestSuite) TestIncrementLockedDelegationEntryID() {
	testCases := []struct {
		name           string
		maleate        func()
		expectedNextID uint64
	}{
		{
			"no call",
			func() {},
			1,
		},
		{
			"single call before",
			func() {
				_ = suite.k.IncrementLockedDelegationEntryID(suite.ctx)
			},
			2,
		},
		{
			"multiple calls before",
			func() {
				for x := 0; x < 1000; x++ {
					_ = suite.k.IncrementLockedDelegationEntryID(suite.ctx)
				}
			},
			1001,
		},
		{
			"multiple calls before - with set initial",
			func() {
				suite.k.SetInitialLockedDelegationEntryID(suite.ctx, 1000)

				for x := 0; x < 1000; x++ {
					_ = suite.k.IncrementLockedDelegationEntryID(suite.ctx)
				}
			},
			2001,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			// Maleate the system
			tc.maleate()

			// Get the next ID and check if it's the expected
			nextID := suite.k.IncrementLockedDelegationEntryID(suite.ctx)
			suite.Require().Equal(tc.expectedNextID, nextID, tc.name)
		})
	}
}

// TestGetSetLockedDelegation tests the store of locked delegation
func (suite *KeeperTestSuite) TestGetSetLockedDelegation() {
	addr1 := sdk.AccAddress([]byte("address1"))
	valAddr1 := sdk.ValAddress([]byte("val1"))

	addr2 := sdk.AccAddress([]byte("address2"))
	valAddr2 := sdk.ValAddress([]byte("val2"))

	rate := types.DefaultRates[0]

	testCases := []struct {
		name   string
		setter func() types.LockedDelegation
		found  bool
	}{
		{
			"no locked delegations",
			func() types.LockedDelegation {
				return types.NewLockedDelegation(
					addr1,
					valAddr1,
					nil,
				)
			},
			false,
		},
		{
			"set locked delegation",
			func() types.LockedDelegation {
				lockedDelegation := types.NewLockedDelegation(
					addr1,
					valAddr1,
					nil,
				)

				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				return lockedDelegation
			},
			true,
		},
		{
			"set locked delegation with entries",
			func() types.LockedDelegation {
				lockedDelegation := types.NewLockedDelegation(
					addr1,
					valAddr1,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyOneDec(),
							rate,
							time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
							false,
							1,
						),
						types.NewLockedDelegationEntry(
							math.LegacyOneDec(),
							rate,
							time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Hour),
							false,
							2,
						),
					},
				)

				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				return lockedDelegation
			},
			true,
		},
		{
			"set bad locked delegation",
			func() types.LockedDelegation {
				// Zero dec should trigger the validate
				lockedDelegation := types.NewLockedDelegation(
					addr1,
					valAddr1,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyZeroDec(),
							rate,
							time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Hour),
							false,
							1,
						),
					},
				)

				// It will return a error and don't store
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().Error(err)

				return lockedDelegation
			},
			false,
		},
		{
			"set multiple locked delegation",
			func() types.LockedDelegation {
				// Set the first address
				lockedDelegation1 := types.NewLockedDelegation(
					addr1,
					valAddr1,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyOneDec(),
							rate,
							time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
							false,
							1,
						),
						types.NewLockedDelegationEntry(
							math.LegacyOneDec(),
							rate,
							time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Hour),
							false,
							2,
						),
					},
				)
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation1)
				suite.Require().NoError(err)

				// Set the second address
				lockedDelegation2 := types.NewLockedDelegation(
					addr2,
					valAddr2,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(10),
							types.DefaultRates[1],
							time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
							false,
							1,
						),
					},
				)
				err = suite.k.SetLockedDelegation(suite.ctx, lockedDelegation2)
				suite.Require().NoError(err)

				return lockedDelegation2
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			lockedDelegation := tc.setter()

			valAddress, err := lockedDelegation.GetValidatorAddr()
			suite.Require().NoError(err)

			outcome, found := suite.k.GetLockedDelegation(
				suite.ctx,
				lockedDelegation.GetDelegatorAddr(),
				valAddress,
			)
			suite.Require().NoError(err)
			if tc.found {
				suite.Require().True(found, tc.name)
				suite.Require().Equal(lockedDelegation, outcome, tc.name)
			} else {
				suite.Require().False(found, tc.name)
				suite.Require().Equal(types.LockedDelegation{}, outcome, tc.name)
			}
		})
	}
}

// TestGetSetLockedDelegations tests the store and fetch of multiple locked delegations
func (suite *KeeperTestSuite) TestGetDeleteLockedDelegations() {
	addr := sdk.AccAddress([]byte("address1"))
	valAddr := sdk.ValAddress([]byte("val1"))

	testCases := []struct {
		name   string
		setter func() []types.LockedDelegation
	}{
		{
			"no locked delegations",
			func() []types.LockedDelegation {
				return []types.LockedDelegation{}
			},
		},
		{
			"1 locked delegation registered",
			func() []types.LockedDelegation {
				lockedDelegations := []types.LockedDelegation{
					types.NewLockedDelegation(
						addr,
						valAddr,
						nil,
					),
				}
				for _, lockedDelegation := range lockedDelegations {
					err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
					suite.Require().NoError(err)
				}

				return lockedDelegations
			},
		},
		{
			"multiple locked delegations registered",
			func() []types.LockedDelegation {
				var lockedDelegations []types.LockedDelegation

				// Generate and store 500 lockedDelegations
				for i := 0; i < 500; i++ {
					newAddr := sdk.AccAddress([]byte(string(rune(i)) + "addr"))
					newValAddr := sdk.ValAddress([]byte(string(rune(i)) + "val"))
					lockedDelegation := types.NewLockedDelegation(
						newAddr,
						newValAddr,
						nil,
					)
					err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
					suite.Require().NoError(err)

					lockedDelegations = append(lockedDelegations, lockedDelegation)
				}

				return lockedDelegations
			},
		},
		{
			"multiple locked delegations registered - random delete",
			func() []types.LockedDelegation {
				var lockedDelegations []types.LockedDelegation

				// Generate and store 500 lockedDelegations
				for i := 0; i < 500; i++ {
					newAddr := sdk.AccAddress([]byte(string(rune(i)) + "addr"))
					newValAddr := sdk.ValAddress([]byte(string(rune(i)) + "val"))
					lockedDelegation := types.NewLockedDelegation(
						newAddr,
						newValAddr,
						nil,
					)
					err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
					suite.Require().NoError(err)

					lockedDelegations = append(lockedDelegations, lockedDelegation)
				}

				var finalLockedDelegations []types.LockedDelegation

				for i, lockedDelegation := range lockedDelegations {
					// Randomly skip values, add to the final list if skipped
					if i%2 == 0 {
						finalLockedDelegations = append(finalLockedDelegations, lockedDelegation)
						continue
					}
					err := suite.k.DeleteLockedDelegation(suite.ctx, lockedDelegation)
					suite.Require().NoError(err)
				}

				return finalLockedDelegations
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			lockedDelegations := tc.setter()
			res := suite.k.GetAllLockedDelegations(suite.ctx)
			suite.Require().ElementsMatch(lockedDelegations, res, tc.name)
		})
	}
}

// SetLockedDelegationEntry tests the set of a locked delegation entry
func (suite *KeeperTestSuite) TestSetLockedDelegationEntry() {
	addr := sdk.AccAddress([]byte("address1"))
	valAddr := sdk.ValAddress([]byte("val1"))

	rates := types.DefaultRates

	lockedDelegation := types.NewLockedDelegation(
		addr,
		valAddr,
		nil,
	)

	baseEntry := types.NewLockedDelegationEntry(
		math.LegacyNewDec(10),
		rates[0],
		time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		false,
		1,
	)

	testCases := []struct {
		name string
		// Do a setter and return the expected locked delegation
		setter func() []types.LockedDelegation
	}{
		{
			"no locked delegation - add entry",
			func() []types.LockedDelegation {
				entry := baseEntry

				_, err := suite.k.SetLockedDelegationEntry(
					suite.ctx,
					addr,
					valAddr,
					entry,
				)
				suite.Require().NoError(err)

				// Expected creation
				// A locked delegation will be created and a entry will be added
				return []types.LockedDelegation{
					types.NewLockedDelegation(
						addr,
						valAddr,
						[]types.LockedDelegationEntry{entry},
					),
				}
			},
		},
		{
			"locked delegation exists - add entry",
			func() []types.LockedDelegation {
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				// Create a entry
				_, err = suite.k.SetLockedDelegationEntry(
					suite.ctx,
					addr,
					valAddr,
					baseEntry,
				)
				suite.Require().NoError(err)

				// Expected creation
				// it will reuse the existing locked delegation
				return []types.LockedDelegation{
					types.NewLockedDelegation(
						addr,
						valAddr,
						[]types.LockedDelegationEntry{baseEntry},
					),
				}
			},
		},
		{
			"locked delegation and entry exists - add entry",
			func() []types.LockedDelegation {
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				// Create a entry
				_, err = suite.k.SetLockedDelegationEntry(
					suite.ctx,
					addr,
					valAddr,
					baseEntry,
				)
				suite.Require().NoError(err)

				// Add again
				_, err = suite.k.SetLockedDelegationEntry(
					suite.ctx,
					addr,
					valAddr,
					baseEntry,
				)
				suite.Require().NoError(err)

				// only the value of the current entry will be updated
				entry := baseEntry
				entry.Shares = baseEntry.Shares.Add(baseEntry.Shares)
				return []types.LockedDelegation{
					types.NewLockedDelegation(
						addr,
						valAddr,
						[]types.LockedDelegationEntry{entry},
					),
				}
			},
		},
		{
			"locked delegation and entry exists - add twice different entries",
			func() []types.LockedDelegation {
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				// Create the entry
				_, err = suite.k.SetLockedDelegationEntry(
					suite.ctx,
					addr,
					valAddr,
					baseEntry,
				)
				suite.Require().NoError(err)

				entry2 := types.NewLockedDelegationEntry(
					math.LegacyNewDec(10),
					rates[0],
					time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Microsecond),
					false,
					2,
				)

				// Add the new entry
				_, err = suite.k.SetLockedDelegationEntry(
					suite.ctx,
					addr,
					valAddr,
					entry2,
				)
				suite.Require().NoError(err)

				// only the value of the current entry will be updated
				return []types.LockedDelegation{
					types.NewLockedDelegation(
						addr,
						valAddr,
						[]types.LockedDelegationEntry{baseEntry, entry2},
					),
				}
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // Reset each run

			expLockedDelegations := tc.setter()
			res := suite.k.GetAllLockedDelegations(suite.ctx)
			suite.Require().ElementsMatch(expLockedDelegations, res, tc.name)
		})
	}
}

// TestIterateAllLockedDelegation verifies the correct functionality of the IterateAllLockedDelegation method
// This forces a stop on the iteration
func (suite *KeeperTestSuite) TestIterateAllLockedDelegation() {
	suite.SetupTest()

	// Generate and store 10 lockedDelegations
	for i := 0; i < 10; i++ {
		newAddr := sdk.AccAddress([]byte(string(rune(i)) + "addr"))
		newValAddr := sdk.ValAddress([]byte(string(rune(i)) + "val"))
		lockedDelegation := types.NewLockedDelegation(
			newAddr,
			newValAddr,
			nil,
		)
		err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
		suite.Require().NoError(err)
	}

	var foundLockedDelegation *types.LockedDelegation
	suite.k.IterateLockedDelegations(suite.ctx, func(lockedDelegation types.LockedDelegation) bool {
		if lockedDelegation.DelegatorAddress == sdk.AccAddress([]byte(string(rune(5))+"addr")).String() {
			foundLockedDelegation = &lockedDelegation
			return true
		}
		return false
	})

	suite.Require().NotNil(foundLockedDelegation)
	suite.Require().Equal(foundLockedDelegation.DelegatorAddress, sdk.AccAddress([]byte(string(rune(5))+"addr")).String())
}

// TestGetLockedDelegationByEntryID tests the store of locked delegation by entry ID
func (suite *KeeperTestSuite) TestGetLockedDelegationByEntryID() {
	addr1 := sdk.AccAddress([]byte("address1"))
	valAddr1 := sdk.ValAddress([]byte("val1"))

	addr2 := sdk.AccAddress([]byte("address2"))
	valAddr2 := sdk.ValAddress([]byte("val2"))

	// rate := types.DefaultRates[0]

	testCases := []struct {
		name   string
		setter func()
		id     uint64
		found  bool
	}{
		{
			"not found - no entry on look up",
			func() {},
			0,
			false,
		},
		{
			"not found - invalid val address, not found",
			func() {
				err := suite.k.SetLockedDelegationByEntryID(suite.ctx, types.LockedDelegation{
					DelegatorAddress: addr1.String(),
					ValidatorAddress: "",
				}, 0)
				suite.Require().Error(err)
			},
			0,
			false,
		},
		{
			"not found - bad locked delegation key",
			func() {
				key := suite.app.GetKey(types.StoreKey)
				store := suite.ctx.KVStore(key)

				// Store a bad bin in the key
				store.Set(types.GetLockedDelegationIndexKey(0), make([]byte, 8))
			},
			0,
			false,
		},
		{
			"not found - bad unmarshal",
			func() {
				key := suite.app.GetKey(types.StoreKey)
				store := suite.ctx.KVStore(key)

				// Store a bad locked delegation
				store.Set(types.GetLockedDelegationKey(addr1, valAddr1), []byte("val1"))

				// Store a correct path
				store.Set(types.GetLockedDelegationIndexKey(0), types.GetLockedDelegationKey(addr1, valAddr1))
			},
			0,
			false,
		},
		{
			"found - store locked delegation",
			func() {
				lockedDelegation := types.NewLockedDelegation(
					addr1,
					valAddr1,
					nil,
				)

				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
				suite.Require().NoError(err)

				err = suite.k.SetLockedDelegationByEntryID(suite.ctx, lockedDelegation, 0)
				suite.Require().NoError(err)
			},
			0,
			true,
		},
		{
			"found - store multiple locked delegations",
			func() {
				// Set the first address
				lockedDelegation1 := types.NewLockedDelegation(
					addr1,
					valAddr1,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyOneDec(),
							rate,
							time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
							false,
							1,
						),
						types.NewLockedDelegationEntry(
							math.LegacyOneDec(),
							rate,
							time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Hour),
							false,
							2,
						),
					},
				)
				err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation1)
				suite.Require().NoError(err)

				// Set the second address
				lockedDelegation2 := types.NewLockedDelegation(
					addr2,
					valAddr2,
					[]types.LockedDelegationEntry{
						types.NewLockedDelegationEntry(
							math.LegacyNewDec(10),
							types.DefaultRates[1],
							time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
							false,
							3,
						),
					},
				)
				err = suite.k.SetLockedDelegation(suite.ctx, lockedDelegation2)
				suite.Require().NoError(err)

				for i := 0; i < 10; i++ {
					err = suite.k.SetLockedDelegationByEntryID(suite.ctx, lockedDelegation1, uint64(i))
					suite.Require().NoError(err)
					err = suite.k.SetLockedDelegationByEntryID(suite.ctx, lockedDelegation2, uint64(i+10))
					suite.Require().NoError(err)
				}
			},
			2,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.setter()

			outcome, found := suite.k.GetLockedDelegationByEntryID(
				suite.ctx,
				tc.id,
			)
			if tc.found {
				suite.Require().True(found, tc.name)
			} else {
				suite.Require().False(found, tc.name)
				suite.Require().Equal(types.LockedDelegation{}, outcome, tc.name)
			}
		})
	}
}

// TestStoreLockedDelegationAddrBadPath tests a specific bad path were storing or deleting can't get validatorAddress
func (suite *KeeperTestSuite) TestStoreLockedDelegationAddrBadPath() {
	suite.SetupTest()

	// Build a lockedDelegation where the validator is empty
	addr := sdk.AccAddress([]byte("address"))
	lockedDelegation := types.LockedDelegation{
		DelegatorAddress: addr.String(),
		ValidatorAddress: "",
		Entries:          nil,
	}
	err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
	suite.Require().Error(err)
	err = suite.k.DeleteLockedDelegation(suite.ctx, lockedDelegation)
	suite.Require().Error(err)
}

// TestStoreLockedDelegationBadMarshal tests a bad marshal function by using a mocked marshal
func (suite *KeeperTestSuite) TestStoreLockedDelegationBadMarshal() {
	suite.SetupTest()

	// Generate a valid lockedDelegation
	addr := sdk.AccAddress([]byte("address1"))
	valAddr := sdk.ValAddress([]byte("val1"))
	lockedDelegation := types.NewLockedDelegation(
		addr,
		valAddr,
		nil,
	)
	// Store it using a valid locking keeper
	err := suite.k.SetLockedDelegation(suite.ctx, lockedDelegation)
	suite.Require().NoError(err)

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

	// Try to store again, should fail
	err = lockingKeeper.SetLockedDelegation(suite.ctx, lockedDelegation)
	suite.Require().Error(err)
}
