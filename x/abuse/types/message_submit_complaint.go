package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgSubmitComplaint{}

func NewMsgSubmitComplaint(creator string, leaseId uint64, subnetId uint64, abuseType uint64, incidentTime int32, sourceIp uint64, lessorTicketRef string, evidenceHash string, summary string) *MsgSubmitComplaint {
	return &MsgSubmitComplaint{
		Creator:         creator,
		LeaseId:         leaseId,
		SubnetId:        subnetId,
		AbuseType:       abuseType,
		IncidentTime:    incidentTime,
		SourceIp:        sourceIp,
		LessorTicketRef: lessorTicketRef,
		EvidenceHash:    evidenceHash,
		Summary:         summary,
	}
}

func (msg *MsgSubmitComplaint) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
