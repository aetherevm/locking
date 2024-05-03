// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package types_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aetherevm/locking/locking/types"
)

// ParamsTestSuite is a test suite for testing the validation of the Params
type ParamsTestSuite struct {
	suite.Suite
}

// TestParamsTestSuite runs the ParamsTestSuite as a test case
func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

// TestParamsValidate tests the validation of the Params type
func (suite *ParamsTestSuite) TestParamsValidate() {
	testCases := []struct {
		name     string
		params   func() types.Params
		expError bool
	}{
		{
			"pass - default",
			types.DefaultParams,
			false,
		},
		{
			"fail - invalid rates duration",
			func() types.Params {
				return types.NewParams(types.DefaultMaxEntries, []types.Rate{
					types.NewRate(0, sdk.ZeroDec()),
				})
			},
			true,
		},
		{
			"fail - invalid rates rate",
			func() types.Params {
				return types.NewParams(types.DefaultMaxEntries, []types.Rate{
					types.NewRate(10, sdk.ZeroDec()),
				})
			},
			true,
		},
		{
			"fail - rates no unique",
			func() types.Params {
				return types.NewParams(types.DefaultMaxEntries, []types.Rate{
					types.NewRate(10, sdk.OneDec()),
					types.NewRate(10, sdk.OneDec()),
				})
			},
			true,
		},
		{
			"fail - invalid rewards denom",
			func() types.Params {
				return types.NewParams(0, nil)
			},
			true,
		},
	}

	for _, tc := range testCases {
		params := tc.params()
		err := params.Validate()

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

// TestParamsString tests the return string from the param
func TestParamsString(t *testing.T) {
	p := types.NewParams(types.DefaultMaxEntries+1, nil)
	expected := fmt.Sprintf("maxentries: %d\nrates: []\n", types.DefaultMaxEntries+1)
	got := p.String()
	require.Equal(t, expected, got)
}

// TestParamsGetRateFromDuration tests the params GetRateFromDuration
func TestParamsGetRateFromDuration(t *testing.T) {
	params := types.DefaultParams()
	rates := types.DefaultRates

	tests := []struct {
		name      string
		duration  time.Duration
		found     bool
		rateFound types.Rate
	}{
		{
			name:      "found",
			duration:  rates[0].Duration,
			found:     true,
			rateFound: rates[0],
		},
		{
			name:      "found - last item",
			duration:  rates[len(rates)-1].Duration,
			found:     true,
			rateFound: rates[len(rates)-1],
		},
		{
			name:      "not found",
			duration:  rates[0].Duration + 1,
			found:     false,
			rateFound: types.Rate{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rate, found := params.GetRateFromDuration(tc.duration)

			if tc.found {
				require.True(t, found, tc.name)
				require.Equal(t, tc.rateFound, rate)
			} else {
				require.False(t, found, tc.name)
				require.Equal(t, types.Rate{}, rate)
			}
		})
	}
}
