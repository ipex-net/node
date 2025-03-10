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
	OrgRegisteredEventAdmin              = "admin"
	OrgRegisteredEventLegalName          = "legal_name"
	OrgRegisteredEventAddress            = "address"
	OrgRegisteredEventRegistrationNumber = "registration_number"
	OrgRegisteredEventCountryCode        = "country_code"
	OrgRegisteredEventTaxId              = "tax_id"
	OrgRegisteredEventWebsite            = "website"
	OrgRegisteredEventAbuseEmail         = "abuse_email"
	OrgRegisteredEventContactEmail       = "contact_email"
	OrgRegisteredEventTreasury			 = "treasury"
)

// Org updated
const (
	OrgUpdatedEventType               = "org_updated"
	OrgUpdatedEventCreator            = "creator"
	OrgUpdatedEventId                 = "org_id"
	OrgUpdatedEventAdmin              = "admin"
	OrgUpdatedEventLegalName          = "legal_name"
	OrgUpdatedEventAddress            = "address"
	OrgUpdatedEventRegistrationNumber = "registration_number"
	OrgUpdatedEventCountryCode        = "country_code"
	OrgUpdatedEventTaxId              = "tax_id"
	OrgUpdatedEventWebsite            = "website"
	OrgUpdatedEventAbuseEmail         = "abuse_email"
	OrgUpdatedEventContactEmail       = "contact_email"
)
