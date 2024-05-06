package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aetherevm/locking/locking/keeper"
	"github.com/aetherevm/locking/locking/tests"
	"github.com/aetherevm/locking/locking/types"
)

// TestParamsStore tests the functionality related to the parameters
func (suite *KeeperTestSuite) TestParamsStore() {
	params := types.DefaultParams()
	err := suite.app.LockingKeeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	testCases := []struct {
		name      string
		paramsFun func() types.Params
		equal     bool
	}{
		{
			"equal - Checks if the default params are set correctly",
			types.DefaultParams,
			true,
		},
		{
			"equal - Update values",
			func() types.Params {
				params := types.NewParams(
					types.DefaultMaxEntries,
					[]types.Rate{
						types.NewRate(10, sdk.OneDec()),
					},
				)
				err := suite.app.LockingKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params
			},
			true,
		},
		{
			"not equal - Bad param key, returning default",
			func() types.Params {
				// Set a new param
				params := types.NewParams(
					types.DefaultMaxEntries,
					nil,
				)
				err := suite.app.LockingKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				// Override the ParamsKey
				types.ParamsKey = []byte("Bad")

				return params
			},
			false,
		},
		{
			"not equal - invalid param",
			func() types.Params {
				// Set a new param
				params := types.NewParams(
					types.DefaultMaxEntries,
					[]types.Rate{types.NewRate(0, sdk.ZeroDec())},
				)
				err := suite.app.LockingKeeper.SetParams(suite.ctx, params)
				suite.Require().Error(err)

				return params
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Some tests are overriding the params key, let make sure it doesn't leak
			oldParamsKey := types.ParamsKey
			defer func() { types.ParamsKey = oldParamsKey }()

			outcome := tc.paramsFun()
			expected := suite.app.LockingKeeper.GetParams(suite.ctx)
			if tc.equal {
				suite.Require().Equal(expected, outcome, tc.name)

				// Since we are here, lets also check the helper functions
				suite.Require().Equal(suite.k.MaxEntries(suite.ctx), outcome.MaxEntries)
				suite.Require().Equal(suite.k.Rates(suite.ctx), outcome.Rates)
			} else {
				suite.Require().NotEqual(expected, outcome, tc.name)
			}
		})
	}
}

// TestSetParamsBadMarshal tests a bad marshal function by using a mocked marshal
func (suite *KeeperTestSuite) TestSetParamsBadMarshal() {
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

	// Now try to add a new value
	param := types.DefaultParams()
	err := lockingKeeper.SetParams(suite.ctx, param)
	suite.Require().Error(err)
}
