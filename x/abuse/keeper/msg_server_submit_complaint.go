package keeper

import (
	"context"
	"regexp"
	"strings"
	"time"

	"ipex/x/abuse/types"
	leasetypes "ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SubmitComplaint(goCtx context.Context, msg *types.MsgSubmitComplaint) (*types.MsgSubmitComplaintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate source IP is in valid range
    const maxIPv4 = uint64(4294967295) 
    
    if msg.SourceIp == 0 {
        return nil, errorsmod.Wrap(types.ErrInvalidIP,
			"source IP cannot be 0")
    }
    
    if msg.SourceIp > maxIPv4 {
        return nil, errorsmod.Wrapf(types.ErrInvalidIP, 
            "source IP %d exceeds maximum IPv4 value %d", 
            msg.SourceIp, 
            maxIPv4)
    }

    // Convert abuseType uint to enum
    abuseType := types.AbuseType(msg.AbuseType)
    if abuseType < types.AbuseType_ABUSE_TYPE_COPYRIGHT_DMCA || abuseType > types.AbuseType_ABUSE_TYPE_OTHER {
    	return nil, errorsmod.Wrapf(types.ErrInvalidAbuseType,
			"got %d",
			msg.AbuseType)
    }

    incidentTime := time.Unix(msg.IncidentTime, 0)
    currentTime := ctx.BlockTime()

    // Check if incident time is in the future
    if incidentTime.After(currentTime) {
        return nil, errorsmod.Wrapf(types.ErrInvalidIncidentTime, 
            "incident time %s is in the future (current: %s)", 
            incidentTime.Format(time.RFC3339), 
            currentTime.Format(time.RFC3339))
    }
	
	// Check if incident time is too old
	maxAge := 7 * 24 * time.Hour
	if currentTime.Sub(incidentTime) > maxAge {
        return nil, errorsmod.Wrapf(types.ErrInvalidIncidentTime, 
            "incident time %s is too old (max age: %s)", 
            incidentTime.Format(time.RFC3339), 
            maxAge.String())
    }


	// Summary validation
	minSummaryLength := 20
	maxSummaryLength := 500
	// Clean summary
	summary := strings.TrimSpace(msg.Summary)
	summary = strings.Join(strings.Fields(summary), " ")

	// Validate summary size 
	if len(summary) < minSummaryLength {
		return nil, errorsmod.Wrapf(types.ErrInvalidSummary, 
			"summary too short (minimum %d characters)", 
			minSummaryLength)
	}

	if len(summary) > maxSummaryLength {
		return nil, errorsmod.Wrapf(types.ErrInvalidSummary, 
			"summary too long (maximum %d characters)", 
			minSummaryLength)
	}


	// Either lease id or subnet id must be provided
	if msg.LeaseId == 0 && msg.SubnetId == 0 {
		return nil, errorsmod.Wrap(types.ErrInvalidRequest, 
			"must provide either leaseId or subnetId")
	}

	leaseId := msg.LeaseId

	// If lease id not set - fetch it from subnet id
	if leaseId == 0 {
		extractedLeaseId, found := k.leaseKeeper.GetFirstLeasedContractBySubnet(ctx, msg.SubnetId)
		if !found {
			return nil, errorsmod.Wrapf(types.ErrInvalidSubnetId, 
				"can not find lease for subnet id %d",
				msg.SubnetId)
		}
		leaseId = extractedLeaseId
	}

	// Extracting lease details
	lease, found := k.leaseKeeper.GetContract(ctx, leaseId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrInvalidLeaseId,
			"invalid lease id %d",
			leaseId)
	}

	// Check that subnet id from lease corresponds to provided subnet id (in case both were provided)
	if msg.SubnetId != lease.SubnetId {
		return nil, errorsmod.Wrapf(types.ErrInvalidSubnetId,
			"provided subnet id %d doesn't match lease record %d",
			msg.SubnetId,
			lease.SubnetId)

	}

	// Validated subnetId
	subnetId := lease.SubnetId

	// Check that incident time is after the lease start time
	if incidentTime.Before(*lease.ActivatedAt) {
        return nil, errorsmod.Wrapf(types.ErrInvalidIncidentTime, 
            "incident time %s is before contract start date %s", 
            incidentTime.Format(time.RFC3339), 
            lease.ActivatedAt.Format(time.RFC3339))
	}

	// Check that lease is in leased stats
	if lease.Status != leasetypes.ContractStatus_CONTRACT_STATUS_LEASED {
		return nil, errorsmod.Wrapf(types.ErrLeaseNotActive, 
			"lease %d has status %s", 
			msg.LeaseId, 
			lease.Status.String())
	}

	//Extracting subnet
	subnet, found := k.subnetKeeper.GetSubnet(ctx, subnetId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrInvalidSubnetId,
			"invalid subnet id %d",
			subnetId)
	}

	// Check that source IP is within the subnet
	hostBits := 32 - uint32(subnet.Mask)
	subnetSize := uint64(1 << hostBits)

	// Calculate broadcast address (last IP in subnet)
	broadcastAddr := subnet.NetworkAddress + subnetSize - 1

	// Check if source IP is within valid host range
	if msg.SourceIp < subnet.NetworkAddress+1 || msg.SourceIp > broadcastAddr {
		return nil, errorsmod.Wrapf(types.ErrInvalidIP, 
			"source IP %d is not a valid host IP in subnet %d/%d (valid range: %d-%d)", 
			msg.SourceIp, 
			subnet.NetworkAddress,
			subnet.Mask,
			subnet.NetworkAddress+1, 
			broadcastAddr)
	}

	// Validate lessor ticket reference
	const maxTicketRefLength = 30 
	if len(msg.LessorTicketRef) > maxTicketRefLength {
		return nil, errorsmod.Wrapf(types.ErrInvalidTicketRef, 
        	"lessor ticket reference too long (maximum %d characters)", 
			maxTicketRefLength)
	}

	// Validate evidence hash (IPFS)
	evidenceHash := strings.TrimSpace(msg.EvidenceHash)

	// Evidence hash is optional - complaints can be submitted without evidence
	if len(evidenceHash) > 0 {
		// IPFS hashes can be in two formats:
		// - CIDv0: starts with "Qm" (46 chars, base58)
		// - CIDv1: starts with "bafy" or "bafk" (59+ chars, base32)
		
		// Basic validation - check if it looks like an IPFS hash
		validIPFSv0 := regexp.MustCompile(`^Qm[1-9A-HJ-NP-Za-km-z]{44}$`)
		validIPFSv1 := regexp.MustCompile(`^ba[a-z2-7]{57,}$`)
		
		if !validIPFSv0.MatchString(evidenceHash) && !validIPFSv1.MatchString(evidenceHash) {
			return nil, errorsmod.Wrapf(types.ErrInvalidEvidenceHash, 
				"evidence hash does not appear to be a valid IPFS hash")
		}
	}

	// Check that the creator is admin of lessor org (subnet owner)
	if !k.subnetKeeper.VerifySubnetOwner(ctx, subnetId, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrUnauthorized,
			"address %s is not authorized to submit complaints for subnet %d",
        	msg.Creator,
        	subnetId)
	}

	// Create the complaint
	complaint := types.Complaint{
		LeaseId:            leaseId,
		SubnetId:           subnetId,
		LessorOrgId:        lease.OwnerOrg,
		LesseeOrgId:        lease.LesseeOrg,
		AbuseType:          abuseType,
		IncidentTime:       &incidentTime,
		SourceIp:           msg.SourceIp,
		LessorTicketRef:    msg.LessorTicketRef,
		EvidenceHash:       evidenceHash,
		Summary:            summary,
		State:              types.ComplaintState_STATE_SUBMITTED,
		SubmittedAt:        &currentTime,
		ResolutionDeadline: nil, // Will be set based on abuse type and SLA
		Thread:             []*types.ThreadMessage{}, // Empty thread initially
		LastActivityAt:     &currentTime,
		ClosedAt:           nil,
		ResolutionFailed:   false,
	}

	// Calculate resolution deadline based on abuse type (industry best practices)
	var deadlineDuration time.Duration
	switch abuseType {
	case types.AbuseType_ABUSE_TYPE_VIOLENCE_HATE_CSAM:
		deadlineDuration = 24 * time.Hour
	case types.AbuseType_ABUSE_TYPE_PHISHING_FRAUD:
		deadlineDuration = 24 * time.Hour
	case types.AbuseType_ABUSE_TYPE_MALWARE:
		deadlineDuration = 24 * time.Hour
	case types.AbuseType_ABUSE_TYPE_DOS_DDOS:
		deadlineDuration = 24 * time.Hour
	case types.AbuseType_ABUSE_TYPE_HACKING:
		deadlineDuration = 24 * time.Hour
	case types.AbuseType_ABUSE_TYPE_COPYRIGHT_DMCA:
		deadlineDuration = 72 * time.Hour
	case types.AbuseType_ABUSE_TYPE_SPAM:
		deadlineDuration = 72 * time.Hour
	case types.AbuseType_ABUSE_TYPE_BRUTEFORCE:
		deadlineDuration = 72 * time.Hour
	case types.AbuseType_ABUSE_TYPE_PORT_SCANNING:
		deadlineDuration = 5 * 24 * time.Hour
	case types.AbuseType_ABUSE_TYPE_OTHER:
		// Default: Other complaints
		deadlineDuration = 7 * 24 * time.Hour
	default:
		// Fallback: Standard business week
		deadlineDuration = 7 * 24 * time.Hour
	}
	
	resolutionDeadline := currentTime.Add(deadlineDuration)
	complaint.ResolutionDeadline = &resolutionDeadline

	// Store the complaint with auto-incrementing ID
	complaintId := k.AppendComplaint(ctx, complaint)

	return &types.MsgSubmitComplaintResponse{Id: complaintId}, nil
}
