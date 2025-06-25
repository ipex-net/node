package keeper

import (
	"context"
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
        return nil, errorsmod.Wrap(types.ErrInvalidIP, "source IP cannot be 0")
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
    	return nil, errorsmod.Wrapf(types.ErrInvalidAbuseType, "got %d", msg.AbuseType)
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
			return nil, errorsmod.Wrapf(types.ErrInvalidSubnetId, "can not find lease for subnet id %d", msg.SubnetId)
		}
		leaseId = extractedLeaseId
	}

	// Extracting lease details
	lease, found := k.leaseKeeper.GetContract(ctx, leaseId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrInvalidLeaseId, "invalid lease id %d", leaseId)
	}

	// Check that subnet id from lease corresponds to provided subnet id (in case both were provided)
	if msg.SubnetId != lease.SubnetId {
		return nil, errorsmod.Wrapf(types.ErrInvalidSubnetId, "provided subnet id %d doesn't match lease record %d", msg.SubnetId, lease.SubnetId)

	}

	// Validated subnetId
	subnetId := lease.SubnetId

	// Check the contract start date
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
		return nil, errorsmod.Wrapf(types.ErrInvalidSubnetId, "invalid subnet id %d", subnetId)
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
	
	return &types.MsgSubmitComplaintResponse{}, nil
}
