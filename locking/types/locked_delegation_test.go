package types_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"

	"github.com/aetherevm/locking/locking/types"
)

// LockedDelegationTestSuite is a test suite for testing the validation of the LockedDelegation
type LockedDelegationTestSuite struct {
	suite.Suite
}

// TestLockedDelegationTestSuite runs the LockedDelegationTestSuite as a test case
func TestLockedDelegationTestSuite(t *testing.T) {
	suite.Run(t, new(LockedDelegationTestSuite))
}

// TestLockedDelegationValidate tests the validation of the LockedDelegation type
func (suite *LockedDelegationTestSuite) TestLockedDelegationValidate() {
	addr := sdk.AccAddress([]byte("address"))
	valAddr := sdk.ValAddress([]byte("val"))

	rate := types.DefaultRates[0]

	testCases := []struct {
		name             string
		lockedDelegation types.LockedDelegation
		expError         bool
	}{
		{
			"pass - valid",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					types.NewLockedDelegationEntry(
						math.LegacyOneDec(),
						rate,
						time.Now(),
						false,
						1,
					),
				},
			),
			false,
		},
		{
			"fail - invalid delegatorAddress",
			types.LockedDelegation{
				DelegatorAddress: "",
				ValidatorAddress: valAddr.String(),
				Entries: []types.LockedDelegationEntry{
					types.NewLockedDelegationEntry(
						math.LegacyOneDec(),
						rate,
						time.Now(),
						false,
						2,
					),
				},
			},
			true,
		},
		{
			"fail - invalid validatorAddress",
			types.LockedDelegation{
				DelegatorAddress: addr.String(),
				ValidatorAddress: "",
				Entries: []types.LockedDelegationEntry{
					types.NewLockedDelegationEntry(
						math.LegacyOneDec(),
						rate,
						time.Now(),
						false,
						3,
					),
				},
			},
			true,
		},
		{
			"fail - invalid rate",
			types.LockedDelegation{
				DelegatorAddress: addr.String(),
				ValidatorAddress: valAddr.String(),
				Entries: []types.LockedDelegationEntry{
					types.NewLockedDelegationEntry(
						math.LegacyOneDec(),
						types.NewRate(0, sdk.ZeroDec()),
						time.Now(),
						false,
						1,
					),
				},
			},
			true,
		},
		{
			"fail - invalid shares",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					types.NewLockedDelegationEntry(
						math.LegacyZeroDec(),
						rate,
						time.Now(),
						false,
						1,
					),
				},
			),
			true,
		},
		{
			"fail - invalid unlockOn",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					types.NewLockedDelegationEntry(
						math.LegacyOneDec(),
						rate,
						time.Time{},
						false,
						1,
					),
				},
			),
			true,
		},
		{
			"fail - duplicated entries",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					types.NewLockedDelegationEntry(
						math.LegacyOneDec(),
						rate,
						time.Time{}.Add(time.Hour),
						false,
						1,
					),
					types.NewLockedDelegationEntry(
						math.LegacyOneDec(),
						rate,
						time.Time{}.Add(time.Hour),
						false,
						1,
					),
				},
			),
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.lockedDelegation.Validate()

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
			suite.Require().Equal(addr, tc.lockedDelegation.GetDelegatorAddr())
			val, err := tc.lockedDelegation.GetValidatorAddr()
			suite.Require().NoError(err, tc.name)
			suite.Require().Equal(valAddr, val, tc.name)
		}
	}
}

// TestLockedDelegationAddRemoveEntry tests the validation of the LockedDelegation type
func (suite *LockedDelegationTestSuite) TestLockedDelegationAddRemoveEntry() {
	addr := sdk.AccAddress([]byte("address"))
	valAddr := sdk.ValAddress([]byte("val"))

	rate := types.DefaultRates[0]

	// Lets start with a locked delegation and build on top
	lockedDelegation := types.NewLockedDelegation(addr, valAddr, nil)

	// Create a few entries
	entry := types.LockedDelegationEntry{
		Rate:      rate,
		AutoRenew: false,
		UnlockOn:  time.Time{}.Add(time.Hour),
		Shares:    math.LegacyNewDec(5),
	}
	entryDiffShares := types.LockedDelegationEntry{
		Rate:      rate,
		AutoRenew: false,
		UnlockOn:  time.Time{}.Add(time.Hour),
		Shares:    math.LegacyNewDec(7),
	}
	differentValuesEntry := types.LockedDelegationEntry{
		Rate:      rate,
		AutoRenew: true,
		UnlockOn:  time.Time{}.Add(time.Second),
		Shares:    math.LegacyNewDec(3),
	}

	// Add it once
	lockedDelegation.AddEntry(entry)
	suite.Require().ElementsMatch(lockedDelegation.Entries, []types.LockedDelegationEntry{entry})

	// Add the same entry again, only the shares should be updated
	lockedDelegation.AddEntry(entryDiffShares)
	suite.Require().Equal(len(lockedDelegation.Entries), 1)

	// Shares also should match
	newShares := entry.Shares.Add(entryDiffShares.Shares)
	suite.Require().Equal(lockedDelegation.TotalShares(), newShares)

	// Add one with diff values
	lockedDelegation.AddEntry(differentValuesEntry)
	suite.Require().Equal(len(lockedDelegation.Entries), 2)
	suite.Require().Equal(lockedDelegation.Entries[1], differentValuesEntry)

	// Add again and compare
	lockedDelegation.AddEntry(differentValuesEntry)
	suite.Require().Equal(len(lockedDelegation.Entries), 2)

	// Shares also should match
	newShares = newShares.Add(differentValuesEntry.Shares.Add(differentValuesEntry.Shares))
	suite.Require().Equal(lockedDelegation.TotalShares(), newShares)

	// Now let's delete
	copyEntries := make([]types.LockedDelegationEntry, len(lockedDelegation.Entries))
	copy(copyEntries, lockedDelegation.Entries)
	// At index -1 nothing should happen
	lockedDelegation.RemoveEntryForIndex(-1)
	suite.Require().ElementsMatch(lockedDelegation.Entries, copyEntries)

	// At index bigger than the len nothing should happen
	lockedDelegation.RemoveEntryForIndex(int64(len(lockedDelegation.Entries)))
	suite.Require().ElementsMatch(lockedDelegation.Entries, copyEntries)

	// Delete at index 0, index 1 should be left
	lockedDelegation.RemoveEntryForIndex(0)
	suite.Require().ElementsMatch(lockedDelegation.Entries, []types.LockedDelegationEntry{copyEntries[1]})

	// Restart and delete at index 1, index 0 should be left
	lockedDelegation.Entries = copyEntries
	lockedDelegation.RemoveEntryForIndex(1)
	suite.Require().ElementsMatch(lockedDelegation.Entries, []types.LockedDelegationEntry{copyEntries[0]})
}

// TestLockedDelegationWeightedRatio tests the locked delegation WeightedRatio
// the expectedRatio never is bigger than the max rate
func (suite *LockedDelegationTestSuite) TestLockedDelegationWeightedRatio() {
	addr := sdk.AccAddress([]byte("address"))
	valAddr := sdk.ValAddress([]byte("val"))

	// The tolerance accepted by the test
	epsilon := 1e-3

	testCases := []struct {
		name             string
		lockedDelegation types.LockedDelegation
		expectedRatio    float64
	}{
		{
			"no entries",
			types.NewLockedDelegation(
				addr,
				valAddr,
				nil,
			),
			0,
		},
		{
			"same shares, same rate - output ratio remains the same",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
				},
			),
			0.05,
		},
		{
			"half shares, same rate - output ratio remains the same",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(25),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
				},
			),
			0.05,
		},
		{
			"50 shares at 5%, 20 shares at 4%",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDec(20),
						Rate:   types.Rate{Rate: math.LegacyNewDec(4)},
					},
				},
			),
			0.04714,
		},
		{
			"100 shares at 3%, 200 shares at 2%",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(100),
						Rate:   types.Rate{Rate: math.LegacyNewDec(3)},
					},
					{
						Shares: math.LegacyNewDec(200),
						Rate:   types.Rate{Rate: math.LegacyNewDec(2)},
					},
				},
			),
			0.02333,
		},
		{
			"100 shares at 3%, 200 shares at 2%",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(80),
						Rate:   types.Rate{Rate: math.LegacyNewDec(6)},
					},
					{
						Shares: math.LegacyNewDec(120),
						Rate:   types.Rate{Rate: math.LegacyNewDec(4)},
					},
				},
			),
			0.0480,
		},
		{
			"40 shares at 3%, 200 shares at 2%",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(80),
						Rate:   types.Rate{Rate: math.LegacyNewDec(6)},
					},
					{
						Shares: math.LegacyNewDec(120),
						Rate:   types.Rate{Rate: math.LegacyNewDec(4)},
					},
				},
			),
			0.0480,
		},
		{
			"three entries",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(60),
						Rate:   types.Rate{Rate: math.LegacyNewDec(3)},
					},
					{
						Shares: math.LegacyNewDec(40),
						Rate:   types.Rate{Rate: math.LegacyNewDec(4)},
					},
					{
						Shares: math.LegacyNewDec(100),
						Rate:   types.Rate{Rate: math.LegacyNewDec(2)},
					},
				},
			),
			0.027,
		},
		{
			"four entries",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(25),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDec(75),
						Rate:   types.Rate{Rate: math.LegacyNewDec(3)},
					},
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(4)},
					},
					{
						Shares: math.LegacyNewDec(100),
						Rate:   types.Rate{Rate: math.LegacyNewDec(2)},
					},
				},
			),
			0.03,
		},
		{
			"single zero",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(0),
						Rate:   types.Rate{Rate: math.LegacyNewDec(0)},
					},
					{
						Shares: math.LegacyNewDec(75),
						Rate:   types.Rate{Rate: math.LegacyNewDec(3)},
					},
				},
			),
			0.03,
		},
		{
			"all zeros",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(0),
						Rate:   types.Rate{Rate: math.LegacyNewDec(0)},
					},
					{
						Shares: math.LegacyNewDec(0),
						Rate:   types.Rate{Rate: math.LegacyNewDec(0)},
					},
				},
			),
			0,
		},
		{
			"40 atto shares at 3%, 200 atto shares at 2%",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDecWithPrec(80, 18),
						Rate:   types.Rate{Rate: math.LegacyNewDec(6)},
					},
					{
						Shares: math.LegacyNewDecWithPrec(120, 18),
						Rate:   types.Rate{Rate: math.LegacyNewDec(4)},
					},
				},
			),
			0.0480,
		},
		{
			"40 exa shares at 3%, 200 exa shares at 2%",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(80).Mul(math.LegacyNewDec(10).Power(18)),
						Rate:   types.Rate{Rate: math.LegacyNewDec(6)},
					},
					{
						Shares: math.LegacyNewDec(120).Mul(math.LegacyNewDec(10).Power(18)),
						Rate:   types.Rate{Rate: math.LegacyNewDec(4)},
					},
				},
			),
			0.0480,
		},
	}
	for _, tc := range testCases {
		weightedRatio := tc.lockedDelegation.WeightedRatio()

		// Tolerance doesn't work as zero
		if tc.expectedRatio == 0 {
			suite.Require().True(weightedRatio.IsZero(), tc.name)
			continue
		}

		// We check values by having a tolerance
		ratioAsFloat, err := weightedRatio.Float64()
		suite.Require().NoError(err)
		suite.Require().InEpsilon(tc.expectedRatio, ratioAsFloat, epsilon, tc.name)
	}
}

// TestLockedDelegationCalculateDelegationRatio tests the locked delegation CalculateDelegationRatio
func (suite *LockedDelegationTestSuite) TestLockedDelegationCalculateDelegationRatio() {
	addr := sdk.AccAddress([]byte("address"))
	valAddr := sdk.ValAddress([]byte("val"))

	// The tolerance accepted by the test
	epsilon := 1e-3

	testCases := []struct {
		name             string
		lockedDelegation types.LockedDelegation
		delegationShare  math.LegacyDec
		expectedRatio    float64
	}{
		{
			"no entries",
			types.NewLockedDelegation(
				addr,
				valAddr,
				nil,
			),
			math.LegacyZeroDec(),
			0,
		},
		{
			"same shares, same rate, delegation smaller than locked - output ratio remains the same",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
				},
			),
			math.LegacyNewDec(50),
			0.05,
		},
		{
			"same shares, same rate, delegation equals to locked - output ratio remains the same",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
				},
			),
			math.LegacyNewDec(100),
			0.05,
		},
		{
			"same shares, same rate, double delegation - output ratio should be halved",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
				},
			),
			math.LegacyNewDec(200),
			0.025,
		},
		{
			"half shares, same rate, same delegation - output ratio remains the same",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(25),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
				},
			),
			math.LegacyNewDec(75),
			0.05,
		},
		{
			"50 shares at 5%, 20 shares at 4%, 250 del",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDec(20),
						Rate:   types.Rate{Rate: math.LegacyNewDec(4)},
					},
				},
			),
			math.LegacyNewDec(250),
			0.0132,
		},
		{
			"del zero",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(50),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDec(20),
						Rate:   types.Rate{Rate: math.LegacyNewDec(4)},
					},
				},
			),
			math.LegacyZeroDec(),
			0,
		},
		{
			"single zero",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(0),
						Rate:   types.Rate{Rate: math.LegacyNewDec(0)},
					},
					{
						Shares: math.LegacyNewDec(75),
						Rate:   types.Rate{Rate: math.LegacyNewDec(3)},
					},
				},
			),
			math.LegacyNewDec(150),
			0.015,
		},
		{
			"50 atto shares at 5%, 20 atto shares at 4%, 250 atto del",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDecWithPrec(50, 18),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDecWithPrec(20, 18),
						Rate:   types.Rate{Rate: math.LegacyNewDec(4)},
					},
				},
			),
			math.LegacyNewDecWithPrec(250, 18),
			0.0132,
		},
		{
			"50 exa shares at 5%, 20 exa shares at 4%, 250 exa del",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{
					{
						Shares: math.LegacyNewDec(50).Mul(math.LegacyNewDec(10).Power(18)),
						Rate:   types.Rate{Rate: math.LegacyNewDec(5)},
					},
					{
						Shares: math.LegacyNewDec(20).Mul(math.LegacyNewDec(10).Power(18)),
						Rate:   types.Rate{Rate: math.LegacyNewDec(4)},
					},
				},
			),
			math.LegacyNewDec(250).Mul(math.LegacyNewDec(10).Power(18)),
			0.0132,
		},
	}
	for _, tc := range testCases {
		ratio := tc.lockedDelegation.CalculateDelegationRatio(tc.delegationShare)

		// Tolerance doesn't work as zero
		if tc.expectedRatio == 0 {
			suite.Require().True(ratio.IsZero(), tc.name)
			continue
		}

		// We check values by having a tolerance
		ratioAsFloat, err := ratio.Float64()
		suite.Require().NoError(err)
		suite.Require().InEpsilon(tc.expectedRatio, ratioAsFloat, epsilon, tc.name)
	}
}

// TestLockedDelegationEntryExpired tests the expired function
func (suite *LockedDelegationTestSuite) TestLockedDelegationEntryExpired() {
	rate := types.DefaultRates[0]
	entry := types.LockedDelegationEntry{
		Rate:      rate,
		AutoRenew: false,
		UnlockOn:  time.Time{}.Add(time.Hour),
		Shares:    math.LegacyNewDec(5),
	}

	testCases := []struct {
		name        string
		currentTime time.Time
		expired     bool
	}{
		{
			"not expired - ms before",
			time.Time{}.Add(time.Hour - time.Millisecond),
			false,
		},
		{
			"expired - same time",
			time.Time{}.Add(time.Hour),
			true,
		},
		{
			"expired - second after",
			time.Time{}.Add(time.Hour),
			true,
		},
	}
	for _, tc := range testCases {
		expired := entry.Expired(tc.currentTime)
		suite.Require().Equal(expired, tc.expired, tc.name)
	}
}

// TestLockedDelegationToggleAutoRenewForId tests the ToggleAutoRenewForId function
func (suite *LockedDelegationTestSuite) TestLockedDelegationToggleAutoRenewForId() {
	addr := sdk.AccAddress([]byte("address"))
	valAddr := sdk.ValAddress([]byte("val"))
	rate := types.DefaultRates[0]

	// Base locked delegation to be used
	// For this test, we only care for ID and auto renew
	baseLockedDelegation := types.NewLockedDelegation(
		addr,
		valAddr,
		[]types.LockedDelegationEntry{
			types.NewLockedDelegationEntry(math.LegacyOneDec(), rate, time.Now(), false, 1),
			types.NewLockedDelegationEntry(math.LegacyOneDec(), rate, time.Now(), true, 60),
			types.NewLockedDelegationEntry(math.LegacyOneDec(), rate, time.Now(), false, 256),
			types.NewLockedDelegationEntry(math.LegacyOneDec(), rate, time.Now(), true, 1000),
			types.NewLockedDelegationEntry(math.LegacyOneDec(), rate, time.Now(), false, 3),
			types.NewLockedDelegationEntry(math.LegacyOneDec(), rate, time.Now(), true, 2),
		},
	)

	testCases := []struct {
		name             string
		lockedDelegation types.LockedDelegation
		id               uint64
		found            bool
	}{
		{
			"not found - empty locked delegation",
			types.NewLockedDelegation(
				addr,
				valAddr,
				[]types.LockedDelegationEntry{},
			),
			0,
			false,
		},
		{
			"not found - id not in entries list",
			baseLockedDelegation,
			10,
			false,
		},
		{
			"not found - id as big number",
			baseLockedDelegation,
			1<<64 - 1,
			false,
		},
		{
			"not found - id as zero",
			baseLockedDelegation,
			0,
			false,
		},
		{
			"found - autoRenew true",
			baseLockedDelegation,
			60,
			true,
		},
		{
			"found - autoRenew false",
			baseLockedDelegation,
			1,
			true,
		},
	}
	for _, tc := range testCases {
		// Copy the locked delegation before updating it
		lockedDelegation := copyLockedDelegation(tc.lockedDelegation)

		// Update the values
		entry, found := lockedDelegation.ToggleAutoRenewForID(tc.id)
		suite.Require().Equal(found, tc.found, tc.name)

		if found {
			// Entry must no be empty
			suite.Require().NotEqual(entry, types.LockedDelegationEntry{}, tc.name)

			// All the locked delegation entries with different id must be equal as before
			// The id must have the auto renew flag flipped and all other values remains the same
			for i, entry := range lockedDelegation.Entries {
				originalEntry := tc.lockedDelegation.Entries[i]

				if entry.Id == tc.id {
					suite.Require().NotEqual(entry.AutoRenew, originalEntry.AutoRenew, tc.name)
					// Check if the remaining values are equal
					// Before comparing, let flip the AutoRenew
					copyEntry := entry
					copyEntry.AutoRenew = !copyEntry.AutoRenew
					suite.Require().Equal(copyEntry, originalEntry, tc.name)
					continue
				}

				// If it's not the target id, the entry should be the same
				suite.Require().Equal(entry, originalEntry, tc.name)
			}

		} else {
			// The entry should be a empty locked delegation
			suite.Require().Equal(entry, types.LockedDelegationEntry{}, tc.name)

			// The locked delegation should remain unchanged
			suite.Require().Equal(lockedDelegation, tc.lockedDelegation, tc.name)
		}
	}
}

// TestGetValidatorAddrBadPath tests a specific bad path were LockedDelegation can't get validatorAddress
func (suite *LockedDelegationTestSuite) TestGetValidatorAddrBadPath() {
	// Build a lockedDelegation where the validator is empty
	addr := sdk.AccAddress([]byte("address"))
	lockedDelegation := types.LockedDelegation{
		DelegatorAddress: addr.String(),
		ValidatorAddress: "",
		Entries:          nil,
	}
	_, err := lockedDelegation.GetValidatorAddr()
	suite.Require().Error(err)
}

// TestLockedDelegationString tests the return string from the lockedDelegation
func (suite *LockedDelegationTestSuite) TestLockedDelegationString() {
	entry := types.NewLockedDelegationEntry(
		math.LegacyOneDec(),
		types.DefaultRates[0],
		time.Time{},
		false,
		1,
	)
	lockedDelegation := types.NewLockedDelegation(
		nil,
		nil,
		[]types.LockedDelegationEntry{entry},
	)
	// Locked delegation
	expected, err := yaml.Marshal(lockedDelegation)
	suite.Require().NoError(err)
	got := lockedDelegation.String()
	suite.Require().Equal(string(expected), got)
}

// TestPairString tests the return string from the pair
func (suite *LockedDelegationTestSuite) TestPairString() {
	pair := types.LockedDelegationPair{
		"nil",
		"nil",
	}
	expected, err := yaml.Marshal(pair)
	suite.Require().NoError(err)
	got := pair.String()
	suite.Require().Equal(string(expected), got)
}

// copyLockedDelegation does a deep copy of a locked delegation
func copyLockedDelegation(ld types.LockedDelegation) types.LockedDelegation {
	newEntries := make([]types.LockedDelegationEntry, len(ld.Entries))
	for i, entry := range ld.Entries {
		newEntries[i] = types.NewLockedDelegationEntry(
			entry.Shares,
			entry.Rate,
			entry.UnlockOn,
			entry.AutoRenew,
			entry.Id,
		)
	}

	return types.LockedDelegation{
		DelegatorAddress: ld.DelegatorAddress,
		ValidatorAddress: ld.ValidatorAddress,
		Entries:          newEntries,
	}
}

// TestCalculateSharesFromValidator tests the return from CalculateSharesFromValidator
// This is not a extensive tests, further testing is done on Cosmos-SDK
//
// since this test is quite similar to TestSimulateValidatorSharesRemoval (as it being the inverse) dupl lint is disabled
//
//nolint:dupl
func (suite *LockedDelegationTestSuite) TestCalculateSharesFromValidator() {
	testCases := []struct {
		name         string
		validator    stakingtypes.Validator
		amount       math.Int
		expectShares math.LegacyDec
	}{
		{
			"Same delegator shares and same amount of tokens",
			stakingtypes.Validator{
				DelegatorShares: math.LegacyNewDec(1000),
				Tokens:          math.NewInt(1000),
			},
			math.NewInt(300_000),
			math.LegacyNewDec(300_000),
		},
		{
			"Slashed validator with half tokens",
			stakingtypes.Validator{
				DelegatorShares: math.LegacyNewDec(1000),
				Tokens:          math.NewInt(500),
			},
			math.NewInt(300),
			math.LegacyNewDec(600),
		},
	}
	for _, tc := range testCases {
		shares := types.CalculateSharesFromValidator(tc.amount, tc.validator)
		suite.Require().EqualValues(tc.expectShares, shares, tc.name)

		// Also check the original function
		_, vShares := tc.validator.AddTokensFromDel(tc.amount)
		suite.Require().EqualValues(shares, vShares, tc.name)
	}
}

// TestSimulateValidatorSharesRemoval tests the return from SimulateValidatorSharesRemoval
// This is not a extensive tests, further testing is done on Cosmos-SDK
//
// since this test is quite similar to TestCalculateSharesFromValidator (as it being the inverse) dupl lint is disabled
//
//nolint:dupl
func (suite *LockedDelegationTestSuite) TestSimulateValidatorSharesRemoval() {
	testCases := []struct {
		name         string
		validator    stakingtypes.Validator
		amount       math.LegacyDec
		expectShares math.Int
	}{
		{
			"Same delegator shares and same amount of tokens",
			stakingtypes.Validator{
				DelegatorShares: math.LegacyNewDec(1_000_000),
				Tokens:          math.NewInt(1_000_000),
			},
			math.LegacyNewDec(300_000),
			math.NewInt(300_000),
		},
		{
			"Slashed validator with half tokens",
			stakingtypes.Validator{
				DelegatorShares: math.LegacyNewDec(10_000_000),
				Tokens:          math.NewInt(5_000_000),
			},
			math.LegacyNewDec(300),
			math.NewInt(150),
		},
	}
	for _, tc := range testCases {
		shares := types.SimulateValidatorSharesRemoval(tc.amount, tc.validator)
		suite.Require().EqualValues(tc.expectShares, shares, tc.name)

		// Also check the original function
		_, vShares := tc.validator.RemoveDelShares(tc.amount)
		suite.Require().EqualValues(shares, vShares, tc.name)
	}
}

// TestEntriesForIds tests the function EntriesForIds on the locked delegation
func (suite *LockedDelegationTestSuite) TestEntriesForIds() {
	defaultRate := types.DefaultRates[0]
	defaultLockedDelegation := types.LockedDelegation{
		DelegatorAddress: "",
		ValidatorAddress: "",
		Entries: []types.LockedDelegationEntry{
			types.NewLockedDelegationEntry(math.LegacyOneDec(), defaultRate, time.Now(), false, 3),
			types.NewLockedDelegationEntry(math.LegacyOneDec(), defaultRate, time.Now(), false, 7),
			types.NewLockedDelegationEntry(math.LegacyOneDec(), defaultRate, time.Now(), false, 15),
		},
	}

	testCases := []struct {
		name             string
		lockedDelegation types.LockedDelegation
		ids              []uint64
		expect           bool
	}{
		{
			"found - empty list of Ids",
			defaultLockedDelegation,
			[]uint64{},
			true,
		},
		{
			"found - nil list",
			defaultLockedDelegation,
			nil,
			true,
		},
		{
			"found - single item in the LD",
			defaultLockedDelegation,
			[]uint64{3},
			true,
		},
		{
			"found - all items in the LD",
			defaultLockedDelegation,
			[]uint64{3, 7, 15},
			true,
		},
		{
			"found - all items in the LD, duplicated request id",
			defaultLockedDelegation,
			[]uint64{3, 3, 7, 15},
			true,
		},
		{
			"not found - single non existing item",
			defaultLockedDelegation,
			[]uint64{4},
			false,
		},
		{
			"not found - all items in the LD, single non existing item",
			defaultLockedDelegation,
			[]uint64{3, 7, 15, 26},
			false,
		},
		{
			"not found - empty LD, single item",
			types.LockedDelegation{},
			[]uint64{3},
			false,
		},
		{
			"not found - empty LD, empty IDs list",
			types.LockedDelegation{},
			[]uint64{3},
			false,
		},
	}
	for _, tc := range testCases {
		exists, entries := tc.lockedDelegation.EntriesForIds(tc.ids)
		suite.Require().Equal(tc.expect, exists, tc.name)

		// If exists, check if the returned list is really on the
		if exists {
			// If the requested list is empty, all ids are returned
			if len(tc.ids) == 0 {
				suite.Require().ElementsMatch(entries, tc.lockedDelegation.Entries, tc.name)
				continue
			}

			// Create a set from the ids list
			setIds := make(map[uint64]struct{})
			for _, id := range tc.ids {
				setIds[id] = struct{}{}
			}

			// Fetch the expected list
			expectedEntries := []types.LockedDelegationEntry{}
			for id := range setIds {
				// We can brute force to find ids for the test
				for _, entry := range tc.lockedDelegation.Entries {
					if entry.Id == id {
						expectedEntries = append(expectedEntries, entry)
						break
					}
				}
			}
			// Check if the entries has the correct len
			suite.Require().Equal(len(setIds), len(entries), tc.name)

			// Check if it has found the correct entries
			suite.Require().ElementsMatch(expectedEntries, entries, tc.name)
		}
	}
}

// TestRemoveEntries tests the function RemoveEntries on the locked delegation
func (suite *LockedDelegationTestSuite) TestERemoveEntries() {
	defaultRate := types.DefaultRates[0]
	defaultLockedDelegation := types.LockedDelegation{
		DelegatorAddress: "",
		ValidatorAddress: "",
		Entries: []types.LockedDelegationEntry{
			types.NewLockedDelegationEntry(math.LegacyOneDec(), defaultRate, time.Now(), false, 3),
			types.NewLockedDelegationEntry(math.LegacyOneDec(), defaultRate, time.Now(), false, 7),
			types.NewLockedDelegationEntry(math.LegacyOneDec(), defaultRate, time.Now(), false, 15),
		},
	}

	testCases := []struct {
		name               string
		lockedDelegation   types.LockedDelegation
		entriesToRemove    []types.LockedDelegationEntry
		expectTotalEntries int
	}{
		{
			"empty list of Ids",
			defaultLockedDelegation,
			[]types.LockedDelegationEntry{},
			3,
		},
		{
			"nil list",
			defaultLockedDelegation,
			nil,
			3,
		},
		{
			"single entry",
			defaultLockedDelegation,
			[]types.LockedDelegationEntry{
				{Id: 3},
			},
			2,
		},
		{
			"all entries",
			defaultLockedDelegation,
			[]types.LockedDelegationEntry{
				{Id: 3},
				{Id: 7},
				{Id: 15},
			},
			0,
		},
		{
			"all entries, duplicated id",
			defaultLockedDelegation,
			[]types.LockedDelegationEntry{
				{Id: 3},
				{Id: 7},
				{Id: 7},
				{Id: 15},
			},
			0,
		},
		{
			"single non existing item",
			defaultLockedDelegation,
			[]types.LockedDelegationEntry{
				{Id: 4},
			},
			3,
		},
		{
			"all items in the LD, single non existing item",
			defaultLockedDelegation,
			[]types.LockedDelegationEntry{
				{Id: 3},
				{Id: 7},
				{Id: 15},
				{Id: 26},
			},
			0,
		},
		{
			"empty LD, single item",
			types.LockedDelegation{},
			[]types.LockedDelegationEntry{
				{Id: 3},
			},
			0,
		},
		{
			"empty LD, empty IDs list",
			types.LockedDelegation{},
			[]types.LockedDelegationEntry{},
			0,
		},
	}
	for _, tc := range testCases {
		tc.lockedDelegation.RemoveEntries(tc.entriesToRemove)

		// Check if the final items are the same amount as expected
		suite.Require().EqualValues(tc.expectTotalEntries, len(tc.lockedDelegation.Entries), tc.name)

		// Create a set from the entries IDs
		setIds := make(map[uint64]struct{})
		for _, entry := range tc.entriesToRemove {
			setIds[entry.Id] = struct{}{}
		}

		// Fetch the expected list
		removedEntries := []types.LockedDelegationEntry{}
		for id := range setIds {
			// We can brute force to find ids for the test
			for _, entry := range tc.lockedDelegation.Entries {
				if entry.Id == id {
					removedEntries = append(removedEntries, entry)
					break
				}
			}
		}

		// The final list should not contain the removed entries
		suite.Require().NotContains(removedEntries, tc.lockedDelegation.Entries, tc.name)
	}
}
