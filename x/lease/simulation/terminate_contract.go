package simulation

import (
	"math/rand"

	"ipex/x/lease/keeper"
	"ipex/x/lease/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgTerminateContract(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgTerminateContract{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the TerminateContract simulation

		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "TerminateContract simulation not implemented"), nil, nil
	}
}
