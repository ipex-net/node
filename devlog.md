# Subnet Module Development Log - December 20, 2024

## Core Design Decisions

### Identity and Ownership

- Subnet ID is required for private iteration
- Only organizations can own subnets
- Users can be assigned as subnet operators
- RIR changes track with owner changes automatically

### User Permissions

- Users can control:
  - AS number assignment
  - Subnet geolocation

### Org Permissions

- Orgs can:
  - Create subnets
  - Delete subnets(?)
  - When subnet is not leased(user is nil) AS and GEO updates are done in background mode and automatically. Owner doesn't control AS origin in the protocol as

### Contact Management

- RIR abuse contact email tracked at subnet level
- Owner and user abuse contacts derived from respective orgs

### Data Synchronization

- RIR sync processes full dataset (owner orgID, RIR, abuse email)
- Failed fetches trigger RIR lookup function
- Different sync intervals based on:
  - Staked amount
  - Organization age
  - IP space under administration

## Routing Registry Integration

### ARIN as Source of Truth

- ARIN IRR provides RPSL objects
- Accessible via:
  - NRTM (Near Real-Time Monitoring)
  - FTP
  - Whois (port 43)
- Free API with NRTM capability

### Origin Tracking

Three types of origins maintained:

1. Desired Origin
   - Set by user or derived from IRR if there's no user
2. Actual IRR Origin
   - Retrieved from ARIN
   - Mismatch triggers contract termination/penalties after some delay
3. RPKI Origin
   - Retrieved from ARIN
   - Mismatch triggers contract termination/penalties after some delay

## Future Product Features

### IP Reputation System

- Spam/blacklist tracking
- RPKI and IRR statistics
- IP database integration
- Blocklist monitoring
- Utilization metrics
- Geofeed discovery service
- Market price tracking for lease/sale

### Deferred Features

- NS management
- Reverse DNS pointers

## Data Sources

### Primary Truth Sources

1. RPKI Data (ARIN IRR DB)
2. IRR Objects (ARIN/RIR)
3. GEOFEED Client Data
4. RIR Ownership (WHOIS)

## Implementation Details

### Data Structure

```go
ignite scaffold type ipv4subnet {
    id:uint
    creationBlock:uint
    networkAddress:bytes
    mask:uint
    status:uint  // Bitmap for various states
    country:string
    countryChangeRequested:bool
    countryChangeRequestedBlock:uint64
    countryChangedBlock:uint64
    geofeedUrl:string
    geofeedSyncBlock:uint
    rir:string
    rirOwnerOrg:string
    rirAbuseEmail:string
    rirAllocationType:uint
    rirSyncBlock:uint64
    ownerOrgId:uint64
    userOrgId:uint64
    origin:string
    originTransferRequested:bool
    originTransferRequestedBlock:uint64
    originTransferedBlock:uint64
    irrOrigin:string
    irrMntBy:string
    irrCreated:string
    irrModified:string
    irrSyncBlock:uint64
    rpkiOrigin:string
    rpkiSyncBlock:uint
}
```
