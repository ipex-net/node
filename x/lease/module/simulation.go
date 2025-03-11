package lease

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"ipex/testutil/sample"
	leasesimulation "ipex/x/lease/simulation"
	"ipex/x/lease/types"
)

// avoid unused import issue
var (
	_ = leasesimulation.FindAccount
	_ = rand.Rand{}
	_ = sample.AccAddress
	_ = sdk.AccAddress{}
	_ = simulation.MsgEntryKind
)

const (
	opWeightMsgCreateContract = "op_weight_msg_create_contract"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateContract int = 100

	opWeightMsgUpdateContract = "op_weight_msg_update_contract"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateContract int = 100

	opWeightMsgDeleteContract = "op_weight_msg_delete_contract"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteContract int = 100

	opWeightMsgCancelContract = "op_weight_msg_cancel_contract"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCancelContract int = 100

	opWeightMsgRevokeContract = "op_weight_msg_revoke_contract"
	// TODO: Determine the simulation weight value
	defaultWeightMsgRevokeContract int = 100

	opWeightMsgEnrollContract = "op_weight_msg_enroll_contract"
	// TODO: Determine the simulation weight value
	defaultWeightMsgEnrollContract int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	leaseGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&leaseGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgCreateContract int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateContract, &weightMsgCreateContract, nil,
		func(_ *rand.Rand) {
			weightMsgCreateContract = defaultWeightMsgCreateContract
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateContract,
		leasesimulation.SimulateMsgCreateContract(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateContract int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateContract, &weightMsgUpdateContract, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateContract = defaultWeightMsgUpdateContract
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateContract,
		leasesimulation.SimulateMsgUpdateContract(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeleteContract int
	simState.AppParams.GetOrGenerate(opWeightMsgDeleteContract, &weightMsgDeleteContract, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteContract = defaultWeightMsgDeleteContract
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteContract,
		leasesimulation.SimulateMsgDeleteContract(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgCancelContract int
	simState.AppParams.GetOrGenerate(opWeightMsgCancelContract, &weightMsgCancelContract, nil,
		func(_ *rand.Rand) {
			weightMsgCancelContract = defaultWeightMsgCancelContract
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCancelContract,
		leasesimulation.SimulateMsgCancelContract(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgRevokeContract int
	simState.AppParams.GetOrGenerate(opWeightMsgRevokeContract, &weightMsgRevokeContract, nil,
		func(_ *rand.Rand) {
			weightMsgRevokeContract = defaultWeightMsgRevokeContract
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRevokeContract,
		leasesimulation.SimulateMsgRevokeContract(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgEnrollContract int
	simState.AppParams.GetOrGenerate(opWeightMsgEnrollContract, &weightMsgEnrollContract, nil,
		func(_ *rand.Rand) {
			weightMsgEnrollContract = defaultWeightMsgEnrollContract
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgEnrollContract,
		leasesimulation.SimulateMsgEnrollContract(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			opWeightMsgCreateContract,
			defaultWeightMsgCreateContract,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				leasesimulation.SimulateMsgCreateContract(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgUpdateContract,
			defaultWeightMsgUpdateContract,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				leasesimulation.SimulateMsgUpdateContract(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgDeleteContract,
			defaultWeightMsgDeleteContract,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				leasesimulation.SimulateMsgDeleteContract(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgCancelContract,
			defaultWeightMsgCancelContract,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				leasesimulation.SimulateMsgCancelContract(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgRevokeContract,
			defaultWeightMsgRevokeContract,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				leasesimulation.SimulateMsgRevokeContract(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgEnrollContract,
			defaultWeightMsgEnrollContract,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				leasesimulation.SimulateMsgEnrollContract(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		// this line is used by starport scaffolding # simapp/module/OpMsg
	}
}
