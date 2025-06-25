package abuse

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"ipex/testutil/sample"
	abusesimulation "ipex/x/abuse/simulation"
	"ipex/x/abuse/types"
)

// avoid unused import issue
var (
	_ = abusesimulation.FindAccount
	_ = rand.Rand{}
	_ = sample.AccAddress
	_ = sdk.AccAddress{}
	_ = simulation.MsgEntryKind
)

const (
	opWeightMsgSubmitComplaint = "op_weight_msg_submit_complaint"
	// TODO: Determine the simulation weight value
	defaultWeightMsgSubmitComplaint int = 100

	opWeightMsgRespondComplaint = "op_weight_msg_respond_complaint"
	// TODO: Determine the simulation weight value
	defaultWeightMsgRespondComplaint int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	abuseGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&abuseGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgSubmitComplaint int
	simState.AppParams.GetOrGenerate(opWeightMsgSubmitComplaint, &weightMsgSubmitComplaint, nil,
		func(_ *rand.Rand) {
			weightMsgSubmitComplaint = defaultWeightMsgSubmitComplaint
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSubmitComplaint,
		abusesimulation.SimulateMsgSubmitComplaint(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgRespondComplaint int
	simState.AppParams.GetOrGenerate(opWeightMsgRespondComplaint, &weightMsgRespondComplaint, nil,
		func(_ *rand.Rand) {
			weightMsgRespondComplaint = defaultWeightMsgRespondComplaint
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRespondComplaint,
		abusesimulation.SimulateMsgRespondComplaint(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			opWeightMsgSubmitComplaint,
			defaultWeightMsgSubmitComplaint,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				abusesimulation.SimulateMsgSubmitComplaint(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgRespondComplaint,
			defaultWeightMsgRespondComplaint,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				abusesimulation.SimulateMsgRespondComplaint(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		// this line is used by starport scaffolding # simapp/module/OpMsg
	}
}
