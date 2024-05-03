// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package types_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aetherevm/locking/locking/types"
)

// ValidateTestSuite is a test suite for testing the validation functions
type ValidateTestSuite struct {
	suite.Suite
}

// TestValidateTestSuite runs the ValidateTestSuite as a test case
func TestValidateTestSuite(t *testing.T) {
	suite.Run(t, new(ValidateTestSuite))
}

// TestValidateTime tests the validate time
func (suite *ValidateTestSuite) TestValidateTime() {
	testCases := []struct {
		name     string
		argument interface{}
		expError bool
	}{
		{
			"pass",
			tmtime.Now(),
			false,
		},
		{
			"fail - not a time",
			"now",
			true,
		},
		{
			"fail - time is zero",
			time.Time{},
			true,
		},
	}

	for _, tc := range testCases {
		err := types.ValidateNonZeroTime(tc.argument)

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

// TestValidateDuration tests the validate duration
func (suite *ValidateTestSuite) TestValidateDuration() {
	testCases := []struct {
		name     string
		argument interface{}
		expError bool
	}{
		{
			"pass",
			time.Hour,
			false,
		},
		{
			"fail - not a duration",
			"now",
			true,
		},
		{
			"fail - duration is zero",
			time.Duration(0),
			true,
		},
	}

	for _, tc := range testCases {
		err := types.ValidateNonZeroDuration(tc.argument)

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

// ValidateDec tests the validate dec
func (suite *ValidateTestSuite) TestValidateDec() {
	testCases := []struct {
		name     string
		argument interface{}
		expError bool
	}{
		{
			"pass",
			sdk.NewDec(1),
			false,
		},
		{
			"fail - not a Dec",
			1,
			true,
		},
		{
			"fail - dec is zero",
			sdk.ZeroDec(),
			true,
		},
	}

	for _, tc := range testCases {
		err := types.ValidateNonZeroDec(tc.argument)

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

// TestValidateInt tests the validate math/int
func (suite *ValidateTestSuite) TestValidateInt() {
	testCases := []struct {
		name     string
		argument interface{}
		expError bool
	}{
		{
			"pass",
			math.NewInt(1),
			false,
		},
		{
			"fail - not a Int",
			1,
			true,
		},
		{
			"fail - int is zero",
			math.ZeroInt(),
			true,
		},
		{
			"fail - int is negative",
			math.NewInt(-1),
			true,
		},
		{
			"fail - int is nil",
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		err := types.ValidatePositiveInt(tc.argument)

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

// ValidateCoin tests the validate coin
// SDK already does validation for bad denom and negative tokens
func (suite *ValidateTestSuite) TestValidateCoin() {
	testCases := []struct {
		name     string
		argument interface{}
		expError bool
	}{
		{
			"pass",
			sdk.NewCoin("test", sdk.OneInt()),
			false,
		},
		{
			"fail - not a coin",
			1,
			true,
		},
		{
			"fail - coin is zero",
			sdk.NewCoin("test", sdk.ZeroInt()),
			true,
		},
		{
			"fail - coin is zero",
			sdk.NewCoin("test", sdk.ZeroInt()),
			true,
		},
		{
			"fail - coin is negative",
			sdk.Coin{
				Denom:  "test",
				Amount: sdk.NewInt(-1),
			},
			true,
		},
		{
			"fail - bad coin denom",
			sdk.Coin{
				Denom:  "",
				Amount: sdk.OneInt(),
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := types.ValidatePositiveCoin(tc.argument)

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

// TestValidateU32 tests the validate uint32
func (suite *ValidateTestSuite) TestValidateU32() {
	testCases := []struct {
		name     string
		argument interface{}
		expError bool
	}{
		{
			"pass",
			uint32(1),
			false,
		},
		{
			"fail - not a uint32",
			"1",
			true,
		},
		{
			"fail - uint32 is zero",
			uint32(0),
			true,
		},
	}

	for _, tc := range testCases {
		err := types.ValidatePositiveU32(tc.argument)

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}
