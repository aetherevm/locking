// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package types_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/aetherevm/locking/locking/types"
)

// KeysTestSuite is a test suite for testing the Keys
type KeysTestSuite struct {
	suite.Suite
}

// TestKeysTestSuite runs the KeysTestSuite as a test case
func TestKeysTestSuite(t *testing.T) {
	suite.Run(t, new(KeysTestSuite))
}

// TestGetLockedDelegationPerDelegatorKey tests the locked delegations prefix generation
func (suite *KeysTestSuite) TestGetLockedDelegationPerDelegatorKey() {
	testCases := []struct {
		name             string
		lockedDelegation types.LockedDelegation
		expHexKey        string
	}{
		{
			"Set string key",
			types.LockedDelegation{
				DelegatorAddress: "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
			},
			"1114003e7ea2b8ff05a1d5a251c3173ded9e0ae3f33c",
		},
		{
			"Set string key 2",
			types.LockedDelegation{
				DelegatorAddress: "cosmos10t8ca2w09ykd6ph0agdz5stvgau47whhaggl9a",
			},
			"11147acf8ea9cf292cdd06efea1a2a416c47795f3af7",
		},
	}

	for _, tc := range testCases {
		keyBytes := types.GetLockedDelegationPerDelegatorKey(tc.lockedDelegation.GetDelegatorAddr())
		key := hex.EncodeToString(keyBytes)

		suite.Require().Equal(tc.expHexKey, key, tc.name)
	}
}

// TestGetLockedDelegationKey tests the locked delegations prefix generation
func (suite *KeysTestSuite) TestGetLockedDelegationKey() {
	testCases := []struct {
		name             string
		lockedDelegation types.LockedDelegation
		expHexKey        string
	}{
		{
			"Set string key",
			types.LockedDelegation{
				DelegatorAddress: "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
				ValidatorAddress: "cosmosvaloper1qql8ag4cluz6r4dz28p3w00dnc9w8ueu6u7aht",
			},
			"1114003e7ea2b8ff05a1d5a251c3173ded9e0ae3f33c14003e7ea2b8ff05a1d5a251c3173ded9e0ae3f33c",
		},
		{
			"Set string key - Update the validator addr",
			types.LockedDelegation{
				DelegatorAddress: "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
				ValidatorAddress: "cosmosvaloper10t8ca2w09ykd6ph0agdz5stvgau47whhcuu2fw",
			},
			"1114003e7ea2b8ff05a1d5a251c3173ded9e0ae3f33c147acf8ea9cf292cdd06efea1a2a416c47795f3af7",
		},
		{
			"Set string key - Update the delegator addr",
			types.LockedDelegation{
				DelegatorAddress: "cosmos10t8ca2w09ykd6ph0agdz5stvgau47whhaggl9a",
				ValidatorAddress: "cosmosvaloper1qql8ag4cluz6r4dz28p3w00dnc9w8ueu6u7aht",
			},
			"11147acf8ea9cf292cdd06efea1a2a416c47795f3af714003e7ea2b8ff05a1d5a251c3173ded9e0ae3f33c",
		},
	}

	for _, tc := range testCases {
		valAddr, err := tc.lockedDelegation.GetValidatorAddr()
		suite.Require().NoError(err, tc.name)
		keyBytes := types.GetLockedDelegationKey(tc.lockedDelegation.GetDelegatorAddr(), valAddr)
		key := hex.EncodeToString(keyBytes)

		suite.Require().Equal(tc.expHexKey, key, tc.name)
	}
}

// TestGetLockedDelegationTimeKey tests the locked delegations queue key generation
func (suite *KeysTestSuite) TestGetLockedDelegationTimeKey() {
	testCases := []struct {
		name      string
		timestamp time.Time
		expHexKey string
	}{
		{
			"Set string key",
			time.Time{},
			"21303030312d30312d30315430303a30303a30302e303030303030303030",
		},
		{
			"Set string key - Add one Nanosecond",
			time.Time{}.Add(time.Nanosecond),
			"21303030312d30312d30315430303a30303a30302e303030303030303031",
		},
		{
			"Set string key - Add years",
			time.Time{}.Add(3 * 365 * 24 * time.Hour),
			"21303030342d30312d30315430303a30303a30302e303030303030303030",
		},
	}

	for _, tc := range testCases {
		keyBytes := types.GetLockedDelegationTimeKey(tc.timestamp)
		key := hex.EncodeToString(keyBytes)

		suite.Require().Equal(tc.expHexKey, key, tc.name)
	}
}

// TestGetLockedDelegationIndexKey tests the locked delegations index key generation
func (suite *KeysTestSuite) TestGetLockedDelegationIndexKey() {
	testCases := []struct {
		name      string
		id        uint64
		expHexKey string
	}{
		{
			"Zero key",
			0,
			"380000000000000000",
		},
		{
			"One key",
			1,
			"380000000000000001",
		},
		{
			"Max int key",
			1<<64 - 1,
			"38ffffffffffffffff",
		},
	}

	for _, tc := range testCases {
		keyBytes := types.GetLockedDelegationIndexKey(tc.id)
		key := hex.EncodeToString(keyBytes)

		suite.Require().Equal(tc.expHexKey, key, tc.name)
	}
}
