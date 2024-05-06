package keeper_test

import (
	"time"

	"github.com/aetherevm/locking/locking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestSetLockedDelegationQueueTimeSlice tests the store of queues on locked
func (suite *KeeperTestSuite) TestSetLockedDelegationQueueTimeSlice() {
	// Set a few testing pairs
	addr1 := sdk.AccAddress([]byte("address1"))
	valAddr1 := sdk.ValAddress([]byte("val1"))
	pair1 := types.LockedDelegationPair{
		DelegatorAddress: addr1.String(), ValidatorAddress: valAddr1.String(),
	}

	addr2 := sdk.AccAddress([]byte("address2"))
	valAddr2 := sdk.ValAddress([]byte("val2"))
	pair2 := types.LockedDelegationPair{
		DelegatorAddress: addr2.String(), ValidatorAddress: valAddr2.String(),
	}

	testCases := []struct {
		name          string
		setter        func()
		expectedPairs []types.LockedDelegationPair
	}{
		{
			"no locked delegations pairs queue",
			func() {},
			nil,
		},
		{
			"1 locked delegation pair to queue",
			func() {
				suite.k.SetLockedDelegationQueueTimeSlice(
					suite.ctx,
					time.Now(),
					[]types.LockedDelegationPair{pair1},
				)
			},
			[]types.LockedDelegationPair{pair1},
		},
		{
			"2 locked delegation pair to queue",
			func() {
				suite.k.SetLockedDelegationQueueTimeSlice(
					suite.ctx,
					time.Now(),
					[]types.LockedDelegationPair{pair1, pair2},
				)
			},
			[]types.LockedDelegationPair{pair1, pair2},
		},
		{
			"set the same pair twice",
			func() {
				suite.k.SetLockedDelegationQueueTimeSlice(
					suite.ctx,
					time.Time{},
					[]types.LockedDelegationPair{pair1},
				)
				suite.k.SetLockedDelegationQueueTimeSlice(
					suite.ctx,
					time.Time{},
					[]types.LockedDelegationPair{pair1},
				)
			},
			[]types.LockedDelegationPair{pair1},
		},
		{
			"set the same pair different timestamp",
			func() {
				suite.k.SetLockedDelegationQueueTimeSlice(
					suite.ctx,
					time.Time{},
					[]types.LockedDelegationPair{pair1},
				)
				suite.k.SetLockedDelegationQueueTimeSlice(
					suite.ctx,
					time.Time{}.Add(time.Nanosecond),
					[]types.LockedDelegationPair{pair1},
				)
			},
			[]types.LockedDelegationPair{pair1, pair1},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.setter()

			outcome := suite.k.GetAllLockedDelegationQueuePairs(
				suite.ctx,
				bigTime,
			)
			suite.Require().ElementsMatch(tc.expectedPairs, outcome, tc.name)
		})
	}
}

// TestInsertLockedDelegationQueue tests the insert of queues on locked
func (suite *KeeperTestSuite) TestInsertLockedDelegationQueue() {
	// Set a few testing pairs
	addr1 := sdk.AccAddress([]byte("address1"))
	valAddr1 := sdk.ValAddress([]byte("val1"))
	pair1 := types.LockedDelegationPair{
		DelegatorAddress: addr1.String(), ValidatorAddress: valAddr1.String(),
	}
	lockedDelegation1 := types.NewLockedDelegation(addr1, valAddr1, nil)

	addr2 := sdk.AccAddress([]byte("address2"))
	valAddr2 := sdk.ValAddress([]byte("val2"))
	pair2 := types.LockedDelegationPair{
		DelegatorAddress: addr2.String(), ValidatorAddress: valAddr2.String(),
	}
	lockedDelegation2 := types.NewLockedDelegation(addr2, valAddr2, nil)

	testCases := []struct {
		name          string
		setter        func()
		expectedPairs []types.LockedDelegationPair
	}{
		{
			"1 locked delegation pair to queue",
			func() {
				suite.k.InsertLockedDelegationQueue(
					suite.ctx,
					lockedDelegation1,
					time.Now(),
				)
			},
			[]types.LockedDelegationPair{pair1},
		},
		{
			"2 locked delegation pair to queue",
			func() {
				suite.k.InsertLockedDelegationQueue(
					suite.ctx,
					lockedDelegation1,
					time.Now(),
				)
				suite.k.InsertLockedDelegationQueue(
					suite.ctx,
					lockedDelegation2,
					time.Now().Add(time.Hour),
				)
			},
			[]types.LockedDelegationPair{pair1, pair2},
		},
		{
			"2 locked delegation pair to queue with same timestamp",
			func() {
				suite.k.InsertLockedDelegationQueue(
					suite.ctx,
					lockedDelegation1,
					time.Time{}.Add(time.Hour),
				)
				suite.k.InsertLockedDelegationQueue(
					suite.ctx,
					lockedDelegation2,
					time.Time{}.Add(time.Hour),
				)
			},
			[]types.LockedDelegationPair{pair1, pair2},
		},
		{
			"same pair twice at the same timestamp",
			func() {
				suite.k.InsertLockedDelegationQueue(
					suite.ctx,
					lockedDelegation1,
					time.Time{},
				)
				suite.k.InsertLockedDelegationQueue(
					suite.ctx,
					lockedDelegation1,
					time.Time{},
				)
			},
			// We don't care that the pair was return twice, we can do some sanitization
			[]types.LockedDelegationPair{pair1, pair1},
		},
		{
			"same pair twice at the different timestamp",
			func() {
				suite.k.InsertLockedDelegationQueue(
					suite.ctx,
					lockedDelegation1,
					time.Time{},
				)
				suite.k.InsertLockedDelegationQueue(
					suite.ctx,
					lockedDelegation1,
					time.Time{}.Add(time.Hour),
				)
			},
			[]types.LockedDelegationPair{pair1, pair1},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.setter()

			// We send a giant time
			outcome := suite.k.GetAllLockedDelegationQueuePairs(
				suite.ctx,
				bigTime,
			)
			suite.Require().ElementsMatch(tc.expectedPairs, outcome, tc.name)
		})
	}
}

// TestLockedDelegationQueueIterator tests the LockedDelegationQueueIterator function
func (suite *KeeperTestSuite) TestLockedDelegationQueueIterator() {
	now := time.Now()

	// The tracking of the final set will be done in this variable:
	var expectedPairs []types.LockedDelegationPair

	// First we generate a random set of pairs
	var lockedDelegations []types.LockedDelegation
	for i := 0; i < 10; i++ {
		newAddr := sdk.AccAddress([]byte(string(rune(i)) + "addr"))
		newValAddr := sdk.ValAddress([]byte(string(rune(i)) + "val"))
		lockedDelegations = append(lockedDelegations, types.NewLockedDelegation(
			newAddr,
			newValAddr,
			nil,
		))
	}

	// Now we add each one 10 times by random time
	saves := 0
	for i, ld := range lockedDelegations {
		for j := 0; j < 10; j++ {
			timestamp := shuffledTimestamp(i * j)

			suite.k.InsertLockedDelegationQueue(
				suite.ctx,
				ld,
				timestamp,
			)
			saves++

			// Let's add to our final expected pairs if we are bellow now
			if timestamp.Before(now) {
				expectedPairs = append(expectedPairs, types.LockedDelegationPair{
					DelegatorAddress: ld.DelegatorAddress,
					ValidatorAddress: ld.ValidatorAddress,
				})
			}
		}
	}

	// Now fetch all locked delegations from the store by the timestamp
	pairs := suite.k.GetAllLockedDelegationQueuePairs(suite.ctx, now)

	// This all should match but the final list should be smaller than the one we started
	suite.Require().ElementsMatch(expectedPairs, pairs)
	suite.Require().Less(len(pairs), saves)
}

// shuffledTimestamp generates a pseudo random time using module values
func shuffledTimestamp(i int) time.Time {
	// Create a base time for consistency.
	baseTime := time.Now()

	// Use modulo to determine the variation
	variation := time.Duration(i%100) * time.Hour

	// Determine if the time should be positive or negative based on the parity of i
	if i%2 == 0 {
		return baseTime.Add(variation)
	}
	return baseTime.Add(-variation)
}
