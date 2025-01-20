package subnet

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"ipex/testutil/sample"
	subnetsimulation "ipex/x/subnet/simulation"
	"ipex/x/subnet/types"
)

// avoid unused import issue
var (
	_ = subnetsimulation.FindAccount
	_ = rand.Rand{}
	_ = sample.AccAddress
	_ = sdk.AccAddress{}
	_ = simulation.MsgEntryKind
)

const (
	opWeightMsgCreateSubnet = "op_weight_msg_create_subnet"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateSubnet int = 100

	opWeightMsgUpdateSubnetUser = "op_weight_msg_update_subnet_user"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateSubnetUser int = 100

	opWeightMsgUpdateSubnetGeofeed = "op_weight_msg_update_subnet_geofeed"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateSubnetGeofeed int = 100

	opWeightMsgRemoveSubnetUser = "op_weight_msg_remove_subnet_user"
	// TODO: Determine the simulation weight value
	defaultWeightMsgRemoveSubnetUser int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	subnetGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&subnetGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgCreateSubnet int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateSubnet, &weightMsgCreateSubnet, nil,
		func(_ *rand.Rand) {
			weightMsgCreateSubnet = defaultWeightMsgCreateSubnet
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateSubnet,
		subnetsimulation.SimulateMsgCreateSubnet(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateSubnetUser int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateSubnetUser, &weightMsgUpdateSubnetUser, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateSubnetUser = defaultWeightMsgUpdateSubnetUser
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateSubnetUser,
		subnetsimulation.SimulateMsgUpdateSubnetUser(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateSubnetGeofeed int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateSubnetGeofeed, &weightMsgUpdateSubnetGeofeed, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateSubnetGeofeed = defaultWeightMsgUpdateSubnetGeofeed
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateSubnetGeofeed,
		subnetsimulation.SimulateMsgUpdateSubnetGeofeed(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgRemoveSubnetUser int
	simState.AppParams.GetOrGenerate(opWeightMsgRemoveSubnetUser, &weightMsgRemoveSubnetUser, nil,
		func(_ *rand.Rand) {
			weightMsgRemoveSubnetUser = defaultWeightMsgRemoveSubnetUser
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRemoveSubnetUser,
		subnetsimulation.SimulateMsgRemoveSubnetUser(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			opWeightMsgCreateSubnet,
			defaultWeightMsgCreateSubnet,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				subnetsimulation.SimulateMsgCreateSubnet(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgUpdateSubnetUser,
			defaultWeightMsgUpdateSubnetUser,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				subnetsimulation.SimulateMsgUpdateSubnetUser(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgUpdateSubnetGeofeed,
			defaultWeightMsgUpdateSubnetGeofeed,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				subnetsimulation.SimulateMsgUpdateSubnetGeofeed(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgRemoveSubnetUser,
			defaultWeightMsgRemoveSubnetUser,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				subnetsimulation.SimulateMsgRemoveSubnetUser(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		// this line is used by starport scaffolding # simapp/module/OpMsg
	}
}
