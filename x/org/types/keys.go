package types

const (
	// ModuleName defines the module name
	ModuleName = "org"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_org"
)

var (
	ParamsKey = []byte("p_org")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	OrgKey             = "Org/value/"
	OrgCountKey        = "Org/count/"
	PermissionKey      = "Permission/value/"
	PermissionCountKey = "Permission/count/"

	// Collection keys for secondary indexes
	OrgByAdminKey   = "Org/by_admin/"
	OrgByCreatorKey = "Org/by_creator/"
	OrgSequenceKey  = "Org/sequence/"
)

// Events
// Org deleted
const (
	OrgDeletedEventType       = "org_deleted"
	OrgDeletedEventCreator    = "creator"
	OrgDeletedEventContractId = "org_id"
)

// Org registered
const (
	OrgRegisteredEventType               = "org_registered"
	OrgRegisteredEventCreator            = "creator"
	OrgRegisteredEventId                 = "org_id"
	OrgRegisteredEventName               = "name"
	OrgRegisteredEventLegalName          = "legal_name"
	OrgRegisteredEventAddress            = "address"
	OrgRegisteredEventRegistrationNumber = "registration_number"
	OrgRegisteredEventCountryCode        = "country_code"
	OrgRegisteredEventTaxId              = "tax_id"
	OrgRegisteredEventWebsite            = "website"
	OrgRegisteredEventAbuseEmail         = "abuse_email"
	OrgRegisteredEventContactEmail       = "contact_email"
)

// Org updated
const (
	OrgUpdatedEventType               = "org_updated"
	OrgUpdatedEventCreator            = "creator"
	OrgUpdatedEventId                 = "org_id"
	OrgUpdatedEventLegalName          = "legal_name"
	OrgUpdatedEventAddress            = "address"
	OrgUpdatedEventRegistrationNumber = "registration_number"
	OrgUpdatedEventCountryCode        = "country_code"
	OrgUpdatedEventTaxId              = "tax_id"
	OrgUpdatedEventWebsite            = "website"
	OrgUpdatedEventAbuseEmail         = "abuse_email"
	OrgUpdatedEventContactEmail       = "contact_email"
)

// Admin changed
const (
	OrgAdminChangedEventType = "org_admin_changed"
	OrgAdminChangedCreator   = "creator"
	OrgAdminChangedId        = "org_id"
	OrgAdminChangedNewAdmin  = "new_admin"
)

// Treasury changed
const (
	OrgTreasuryChangedEventType 	= "org_treasury_changed"
	OrgTreasuryChangedCreator   	= "creator"
	OrgTreasuryChangedId       		= "org_id"
	OrgTreasuryChangedNewTreasury  	= "new_treasury"
)
