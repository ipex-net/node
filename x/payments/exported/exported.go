package exported
type TransactionType int32

const (
	TransactionType_TRANSACTION_TYPE_UNSPECIFIED            TransactionType = 0
	TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT          TransactionType = 1
	TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT TransactionType = 2
	TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_REFUND  TransactionType = 3
	TransactionType_TRANSACTION_TYPE_REFUND                 TransactionType = 4
	TransactionType_TRANSACTION_TYPE_SUBNET_SALE            TransactionType = 5
	TransactionType_TRANSACTION_TYPE_ASN_SALE               TransactionType = 6
)

var TransactionType_name = map[int32]string{
	0: "TRANSACTION_TYPE_UNSPECIFIED",
	1: "TRANSACTION_TYPE_LEASE_PAYMENT",
	2: "TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT",
	3: "TRANSACTION_TYPE_SAFETY_DEPOSIT_REFUND",
	4: "TRANSACTION_TYPE_REFUND",
	5: "TRANSACTION_TYPE_SUBNET_SALE",
	6: "TRANSACTION_TYPE_ASN_SALE",
}

var TransactionType_value = map[string]int32{
	"TRANSACTION_TYPE_UNSPECIFIED":            0,
	"TRANSACTION_TYPE_LEASE_PAYMENT":          1,
	"TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT": 2,
	"TRANSACTION_TYPE_SAFETY_DEPOSIT_REFUND":  3,
	"TRANSACTION_TYPE_REFUND":                 4,
	"TRANSACTION_TYPE_SUBNET_SALE":            5,
	"TRANSACTION_TYPE_ASN_SALE":               6,
}


type ModuleID int32

const (
	ModuleID_MODULE_ID_UNSPECIFIED ModuleID = 0
	ModuleID_MODULE_ID_LEASE       ModuleID = 1
	ModuleID_MODULE_ID_SUBNET_SALE ModuleID = 2
	ModuleID_MODULE_ID_ASN_SALE    ModuleID = 3
)

var ModuleID_name = map[int32]string{
	0: "MODULE_ID_UNSPECIFIED",
	1: "MODULE_ID_LEASE",
	2: "MODULE_ID_SUBNET_SALE",
	3: "MODULE_ID_ASN_SALE",
}

var ModuleID_value = map[string]int32{
	"MODULE_ID_UNSPECIFIED": 0,
	"MODULE_ID_LEASE":       1,
	"MODULE_ID_SUBNET_SALE": 2,
	"MODULE_ID_ASN_SALE":    3,
}
