package types_test

import (
	"testing"

	"ipex/x/trust/types"

	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{

				SubnetReputationList: []types.SubnetReputation{
					{
						Index: 0,
					},
					{
						Index: 1,
					},
				},
				BlacklistLogList: []types.BlacklistLog{
					{
						Id: 0,
					},
					{
						Id: 1,
					},
				},
				BlacklistLogCount: 2,
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated subnetReputation",
			genState: &types.GenesisState{
				SubnetReputationList: []types.SubnetReputation{
					{
						Index: 0,
					},
					{
						Index: 0,
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated blacklistLog",
			genState: &types.GenesisState{
				BlacklistLogList: []types.BlacklistLog{
					{
						Id: 0,
					},
					{
						Id: 0,
					},
				},
			},
			valid: false,
		},
		{
			desc: "invalid blacklistLog count",
			genState: &types.GenesisState{
				BlacklistLogList: []types.BlacklistLog{
					{
						Id: 1,
					},
				},
				BlacklistLogCount: 0,
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
