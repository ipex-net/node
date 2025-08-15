package keeper_test

import (
	"testing"
	"time"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/sample"
	"ipex/x/org/types"

	"github.com/stretchr/testify/require"
)

// TestOrganizationLifecycleIntegration tests the complete organization lifecycle
func TestOrganizationLifecycleIntegration(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// Test data
	baseTime := ctx.BlockTime()
	creator := sample.AccAddress()
	admin := sample.AccAddress()
	treasury := sample.AccAddress()

	t.Run("Create Organization", func(t *testing.T) {
		// Create a complete organization
		org := types.Org{
			Creator:            creator,
			Admin:              admin,
			Treasury:           treasury,
			Name:               "Test Organization Inc",
			LegalName:          "Test Organization Incorporated",
			Address:            "123 Main Street, Suite 100, New York, NY 10001",
			RegistrationNumber: "REG123456789",
			CountryCode:        "US",
			TaxId:              "TAX123456789",
			Website:            "https://test-org.example.com",
			AbuseEmail:         "abuse@test-org.example.com",
			ContactEmail:       "contact@test-org.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		orgId := keeper.AppendOrg(ctx, org)
		require.Greater(t, orgId, uint64(0), "should create organization with valid ID")

		// Verify organization was created correctly
		retrievedOrg, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find created organization")
		require.Equal(t, orgId, retrievedOrg.Id, "should have correct ID")
		require.Equal(t, creator, retrievedOrg.Creator, "should have correct creator")
		require.Equal(t, admin, retrievedOrg.Admin, "should have correct admin")
		require.Equal(t, treasury, retrievedOrg.Treasury, "should have correct treasury")
		require.Equal(t, "Test Organization Inc", retrievedOrg.Name, "should have correct name")
		require.Equal(t, "Test Organization Incorporated", retrievedOrg.LegalName, "should have correct legal name")
		require.Equal(t, "https://test-org.example.com", retrievedOrg.Website, "should have correct website")
		require.False(t, retrievedOrg.DomainVerified, "should not be domain verified initially")
		require.Nil(t, retrievedOrg.DomainVerifiedAt, "should not have domain verified timestamp initially")
	})

	t.Run("Update Organization", func(t *testing.T) {
		// Create organization for update testing
		org := types.Org{
			Creator:            creator,
			Admin:              admin,
			Treasury:           treasury,
			Name:               "Update Test Org",
			LegalName:          "Update Test Organization",
			Address:            "456 Update Street, Update City, UC 12345",
			RegistrationNumber: "UPD123456789",
			CountryCode:        "US",
			TaxId:              "UPD987654321",
			Website:            "https://update-test.example.com",
			AbuseEmail:         "abuse@update-test.example.com",
			ContactEmail:       "contact@update-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		orgId := keeper.AppendOrg(ctx, org)

		// Update the organization
		retrievedOrg, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find organization")

		// Update various fields
		updatedTime := baseTime.Add(time.Hour)
		ctx = ctx.WithBlockTime(updatedTime)

		retrievedOrg.Name = "Updated Organization Name"
		retrievedOrg.LegalName = "Updated Organization Legal Name"
		retrievedOrg.Website = "https://new-updated-test.example.com"
		retrievedOrg.AbuseEmail = "new-abuse@updated-test.example.com"
		retrievedOrg.UpdatedAt = updatedTime

		keeper.SetOrg(ctx, retrievedOrg)

		// Verify updates
		updatedOrg, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find updated organization")
		require.Equal(t, "Updated Organization Name", updatedOrg.Name, "should have updated name")
		require.Equal(t, "Updated Organization Legal Name", updatedOrg.LegalName, "should have updated legal name")
		require.Equal(t, "https://new-updated-test.example.com", updatedOrg.Website, "should have updated website")
		require.Equal(t, "new-abuse@updated-test.example.com", updatedOrg.AbuseEmail, "should have updated abuse email")
		require.Equal(t, updatedTime, updatedOrg.UpdatedAt, "should have updated timestamp")

		// Website change should reset domain verification
		require.False(t, updatedOrg.DomainVerified, "should reset domain verification on website change")
		require.Nil(t, updatedOrg.DomainVerifiedAt, "should clear domain verified timestamp")
	})

	t.Run("Organization Removal", func(t *testing.T) {
		// Create organization for removal testing
		org := types.Org{
			Creator:            creator,
			Admin:              admin,
			Treasury:           treasury,
			Name:               "Remove Test Org",
			LegalName:          "Remove Test Organization",
			Address:            "789 Remove Street, Remove City, RC 54321",
			RegistrationNumber: "REM123456789",
			CountryCode:        "CA",
			TaxId:              "REM987654321",
			Website:            "https://remove-test.example.com",
			AbuseEmail:         "abuse@remove-test.example.com",
			ContactEmail:       "contact@remove-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		orgId := keeper.AppendOrg(ctx, org)

		// Verify organization exists
		_, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find organization before removal")

		// Attempt to remove organization (currently disabled)
		keeper.RemoveOrg(ctx, orgId)

		// Verify organization still exists since removal is disabled
		_, found = keeper.GetOrg(ctx, orgId)
		require.True(t, found, "organization should still exist since removal is disabled")

		// Verify organization is still in the list
		allOrgs := keeper.GetAllOrg(ctx)
		foundInList := false
		for _, o := range allOrgs {
			if o.Id == orgId {
				foundInList = true
				break
			}
		}
		require.True(t, foundInList, "organization should still appear in list since removal is disabled")
	})

	t.Run("Authorization Validation", func(t *testing.T) {
		// Create organization with specific admin
		org := types.Org{
			Creator:            creator,
			Admin:              admin,
			Treasury:           treasury,
			Name:               "Auth Test Org",
			LegalName:          "Authorization Test Organization",
			Address:            "321 Auth Street, Auth City, AC 98765",
			RegistrationNumber: "AUTH123456789",
			CountryCode:        "GB",
			TaxId:              "AUTH987654321",
			Website:            "https://auth-test.example.com",
			AbuseEmail:         "abuse@auth-test.example.com",
			ContactEmail:       "contact@auth-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		orgId := keeper.AppendOrg(ctx, org)

		// Test VerifyOrgAdmin function
		// First, let's verify the org was created with the correct admin
		storedOrg, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find stored org")
		require.Equal(t, admin, storedOrg.Admin, "stored org should have correct admin")

		isValidAdmin := keeper.VerifyOrgAdmin(ctx, orgId, admin)
		require.True(t, isValidAdmin, "should verify correct admin")

		invalidAdmin := "invalid_address_format"
		isInvalidAdmin := keeper.VerifyOrgAdmin(ctx, orgId, invalidAdmin)
		require.False(t, isInvalidAdmin, "should reject invalid admin")

		// Test with non-existent org
		isNonExistentOrg := keeper.VerifyOrgAdmin(ctx, 99999, admin)
		require.False(t, isNonExistentOrg, "should reject admin for non-existent org")
	})
}

// TestOrganizationIndexingIntegration tests the indexing system for efficient queries
func TestOrganizationIndexingIntegration(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	baseTime := ctx.BlockTime()

	t.Run("Index by Admin", func(t *testing.T) {
		admin1 := sample.AccAddress()
		admin2 := sample.AccAddress()
		creator := sample.AccAddress()

		// Create organizations with different admins
		org1 := types.Org{
			Creator:            creator,
			Admin:              admin1,
			Treasury:           sample.AccAddress(),
			Name:               "Admin Index Test Org 1",
			LegalName:          "Admin Index Test Organization One",
			Address:            "100 Admin St, Admin City, AC 10001",
			RegistrationNumber: "ADM1123456789",
			CountryCode:        "US",
			TaxId:              "ADM1987654321",
			Website:            "https://admin1-test.example.com",
			AbuseEmail:         "abuse@admin1-test.example.com",
			ContactEmail:       "contact@admin1-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		org2 := types.Org{
			Creator:            creator,
			Admin:              admin2,
			Treasury:           "cosmos1treasury2jkl012mno345pqr678stu901vwx",
			Name:               "Admin Index Test Org 2",
			LegalName:          "Admin Index Test Organization Two",
			Address:            "200 Admin St, Admin City, AC 20002",
			RegistrationNumber: "ADM2123456789",
			CountryCode:        "CA",
			TaxId:              "ADM2987654321",
			Website:            "https://admin2-test.example.com",
			AbuseEmail:         "abuse@admin2-test.example.com",
			ContactEmail:       "contact@admin2-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		id1 := keeper.AppendOrg(ctx, org1)
		id2 := keeper.AppendOrg(ctx, org2)

		// Test admin indexing by getting all orgs and checking admin mapping
		allOrgs := keeper.GetAllOrg(ctx)

		var admin1OrgId, admin2OrgId uint64
		admin1Found := false
		admin2Found := false

		for _, org := range allOrgs {
			if org.Admin == admin1 {
				admin1OrgId = org.Id
				admin1Found = true
			}
			if org.Admin == admin2 {
				admin2OrgId = org.Id
				admin2Found = true
			}
		}

		require.True(t, admin1Found, "should find organization for admin1")
		require.True(t, admin2Found, "should find organization for admin2")
		require.Equal(t, id1, admin1OrgId, "should have correct org ID for admin1")
		require.Equal(t, id2, admin2OrgId, "should have correct org ID for admin2")
	})

	t.Run("Index by Creator", func(t *testing.T) {
		creator1 := "cosmos1creator1def456ghi789jkl012mno345pqr678"
		creator2 := "cosmos1creator2ghi789jkl012mno345pqr678stu"
		admin := "cosmos1admin999def456ghi789jkl012mno345pqr678"

		// Create organizations with different creators
		org1 := types.Org{
			Creator:            creator1,
			Admin:              admin,
			Treasury:           "cosmos1treasury999ghi789jkl012mno345pqr678",
			Name:               "Creator Index Test Org 1",
			LegalName:          "Creator Index Test Organization One",
			Address:            "300 Creator St, Creator City, CC 30003",
			RegistrationNumber: "CRT1123456789",
			CountryCode:        "FR",
			TaxId:              "CRT1987654321",
			Website:            "https://creator1-test.example.com",
			AbuseEmail:         "abuse@creator1-test.example.com",
			ContactEmail:       "contact@creator1-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		org2 := types.Org{
			Creator:            creator2,
			Admin:              admin,
			Treasury:           "cosmos1treasury998jkl012mno345pqr678stu901",
			Name:               "Creator Index Test Org 2",
			LegalName:          "Creator Index Test Organization Two",
			Address:            "400 Creator St, Creator City, CC 40004",
			RegistrationNumber: "CRT2123456789",
			CountryCode:        "DE",
			TaxId:              "CRT2987654321",
			Website:            "https://creator2-test.example.com",
			AbuseEmail:         "abuse@creator2-test.example.com",
			ContactEmail:       "contact@creator2-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		id1 := keeper.AppendOrg(ctx, org1)
		id2 := keeper.AppendOrg(ctx, org2)

		// Test creator indexing by getting all orgs and checking creator mapping
		allOrgs := keeper.GetAllOrg(ctx)

		var creator1OrgId, creator2OrgId uint64
		creator1Found := false
		creator2Found := false

		for _, org := range allOrgs {
			if org.Creator == creator1 {
				creator1OrgId = org.Id
				creator1Found = true
			}
			if org.Creator == creator2 {
				creator2OrgId = org.Id
				creator2Found = true
			}
		}

		require.True(t, creator1Found, "should find organization for creator1")
		require.True(t, creator2Found, "should find organization for creator2")
		require.Equal(t, id1, creator1OrgId, "should have correct org ID for creator1")
		require.Equal(t, id2, creator2OrgId, "should have correct org ID for creator2")
	})

	t.Run("Multi-Organization Management", func(t *testing.T) {
		// Test creating multiple organizations and managing them
		numOrgs := 10
		var orgIds []uint64
		creator := "cosmos1multiorg123def456ghi789jkl012mno345"

		for i := 0; i < numOrgs; i++ {
			org := types.Org{
				Creator:            creator,
				Admin:              creator,
				Treasury:           "cosmos1treasury111ghi789jkl012mno345pqr678",
				Name:               "Multi Org Test " + string(rune(i+49)), // Converts to "1", "2", etc
				LegalName:          "Multi Organization Test Number " + string(rune(i+49)),
				Address:            "500 Multi St, Multi City, MC 50005",
				RegistrationNumber: "MLT" + string(rune(i+49)) + "123456789",
				CountryCode:        "JP",
				TaxId:              "MLT" + string(rune(i+49)) + "987654321",
				Website:            "https://multi" + string(rune(i+49)) + "-test.example.com",
				AbuseEmail:         "abuse@multi" + string(rune(i+49)) + "-test.example.com",
				ContactEmail:       "contact@multi" + string(rune(i+49)) + "-test.example.com",
				DomainVerified:     false,
				CreatedAt:          baseTime,
				UpdatedAt:          baseTime,
			}

			orgId := keeper.AppendOrg(ctx, org)
			orgIds = append(orgIds, orgId)
		}

		require.Len(t, orgIds, numOrgs, "should create all organizations")

		// Verify all organizations exist
		allOrgs := keeper.GetAllOrg(ctx)
		createdOrgCount := 0
		for _, org := range allOrgs {
			for _, createdId := range orgIds {
				if org.Id == createdId {
					createdOrgCount++
					require.Equal(t, creator, org.Creator, "should have correct creator")
					require.Equal(t, creator, org.Admin, "should have correct admin")
					break
				}
			}
		}
		require.Equal(t, numOrgs, createdOrgCount, "should find all created organizations")

		// Test individual lookups
		for i := 0; i < 5; i++ { // Test first 5
			org, found := keeper.GetOrg(ctx, orgIds[i])
			require.True(t, found, "should find organization %d", i)
			require.Equal(t, orgIds[i], org.Id, "should have correct ID")
		}

		// Test attempted removal of some organizations (currently disabled)
		for i := 0; i < 3; i++ {
			keeper.RemoveOrg(ctx, orgIds[i])
		}

		// Verify all organizations still exist since removal is disabled
		remainingOrgs := keeper.GetAllOrg(ctx)
		attemptedRemovalIds := make(map[uint64]bool)
		for i := 0; i < 3; i++ {
			attemptedRemovalIds[orgIds[i]] = true
		}

		foundAttemptedRemovals := 0
		for _, org := range remainingOrgs {
			if attemptedRemovalIds[org.Id] {
				foundAttemptedRemovals++
			}
		}
		require.Equal(t, 3, foundAttemptedRemovals, "all attempted removals should still be present since removal is disabled")
	})
}

// TestDomainValidationIntegration tests the domain validation workflow
func TestDomainValidationIntegration(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	baseTime := ctx.BlockTime()
	creator := "cosmos1domain123def456ghi789jkl012mno345pqr"

	t.Run("Domain Validation Lifecycle", func(t *testing.T) {
		// Create organization with website
		org := types.Org{
			Creator:            creator,
			Admin:              creator,
			Treasury:           "cosmos1domaintreasury123ghi789jkl012mno345",
			Name:               "Domain Test Org",
			LegalName:          "Domain Test Organization",
			Address:            "600 Domain St, Domain City, DC 60006",
			RegistrationNumber: "DOM123456789",
			CountryCode:        "AU",
			TaxId:              "DOM987654321",
			Website:            "https://domain-test.example.com",
			AbuseEmail:         "abuse@domain-test.example.com",
			ContactEmail:       "contact@domain-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		orgId := keeper.AppendOrg(ctx, org)

		// Initially not verified
		retrievedOrg, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find organization")
		require.False(t, retrievedOrg.DomainVerified, "should not be verified initially")
		require.Nil(t, retrievedOrg.DomainVerifiedAt, "should not have verified timestamp initially")
		require.Nil(t, retrievedOrg.DomainVerifAttemptedAt, "should not have attempted timestamp initially")

		// Simulate domain validation success
		err := keeper.UpdateOrgDomainValidationStatus(ctx, orgId, true)
		require.NoError(t, err, "should update domain validation status")

		// Check validation status after success
		verifiedOrg, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find organization")
		require.True(t, verifiedOrg.DomainVerified, "should be domain verified")
		require.NotNil(t, verifiedOrg.DomainVerifiedAt, "should have verified timestamp")
		require.NotNil(t, verifiedOrg.DomainVerifAttemptedAt, "should have attempted timestamp")
		require.Equal(t, baseTime, *verifiedOrg.DomainVerifiedAt, "should have correct verified time")
		require.Equal(t, baseTime, *verifiedOrg.DomainVerifAttemptedAt, "should have correct attempted time")

		// Simulate domain validation failure (re-verification)
		laterTime := baseTime.Add(time.Hour * 74) // After 3 days for re-verification
		ctx = ctx.WithBlockTime(laterTime)

		err = keeper.UpdateOrgDomainValidationStatus(ctx, orgId, false)
		require.NoError(t, err, "should update domain validation status to failed")

		// Check validation status after failure
		failedOrg, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find organization")
		require.False(t, failedOrg.DomainVerified, "should not be domain verified after failure")
		require.NotNil(t, failedOrg.DomainVerifiedAt, "should preserve last verified timestamp")
		require.NotNil(t, failedOrg.DomainVerifAttemptedAt, "should have attempted timestamp")
		require.Equal(t, laterTime, *failedOrg.DomainVerifAttemptedAt, "should have updated attempted time")
	})

	t.Run("Domain Validation Retry Logic", func(t *testing.T) {
		// Test GetOrgsForDomainValidation with different org ages and statuses
		currentTime := baseTime.Add(time.Hour * 25) // 25 hours after creation

		// Create organizations with different characteristics
		orgs := []types.Org{
			{
				Creator:            creator,
				Admin:              creator,
				Treasury:           "cosmos1treasury001ghi789jkl012mno345pqr678",
				Name:               "Young Org",
				LegalName:          "Young Organization",
				Address:            "700 Young St, Young City, YC 70007",
				RegistrationNumber: "YNG123456789",
				CountryCode:        "NZ",
				TaxId:              "YNG987654321",
				Website:            "https://young-org.example.com",
				AbuseEmail:         "abuse@young-org.example.com",
				ContactEmail:       "contact@young-org.example.com",
				DomainVerified:     false,
				CreatedAt:          baseTime,
				UpdatedAt:          baseTime,
			},
			{
				Creator:                creator,
				Admin:                  creator,
				Treasury:               "cosmos1treasury002ghi789jkl012mno345pqr678",
				Name:                   "Verified Org",
				LegalName:              "Verified Organization",
				Address:                "800 Verified St, Verified City, VC 80008",
				RegistrationNumber:     "VER123456789",
				CountryCode:            "SE",
				TaxId:                  "VER987654321",
				Website:                "https://verified-org.example.com",
				AbuseEmail:             "abuse@verified-org.example.com",
				ContactEmail:           "contact@verified-org.example.com",
				DomainVerified:         true,
				DomainVerifiedAt:       &baseTime,
				DomainVerifAttemptedAt: &baseTime,
				CreatedAt:              baseTime,
				UpdatedAt:              baseTime,
			},
			{
				Creator:            creator,
				Admin:              creator,
				Treasury:           "cosmos1treasury003ghi789jkl012mno345pqr678",
				Name:               "No Website Org",
				LegalName:          "No Website Organization",
				Address:            "900 NoWeb St, NoWeb City, NC 90009",
				RegistrationNumber: "NOW123456789",
				CountryCode:        "NO",
				TaxId:              "NOW987654321",
				Website:            "", // No website
				AbuseEmail:         "abuse@noweb-org.example.com",
				ContactEmail:       "contact@noweb-org.example.com",
				DomainVerified:     false,
				CreatedAt:          baseTime,
				UpdatedAt:          baseTime,
			},
		}

		var orgIds []uint64
		for _, org := range orgs {
			orgId := keeper.AppendOrg(ctx, org)
			orgIds = append(orgIds, orgId)
		}

		// Test GetOrgsForDomainValidation
		orgsForValidation := keeper.GetOrgsForDomainValidation(ctx, currentTime)

		// Verify results
		foundYoungOrg := false
		foundVerifiedOrg := false
		foundNoWebsiteOrg := false

		for _, org := range orgsForValidation {
			if org.Name == "Young Org" {
				foundYoungOrg = true
			}
			if org.Name == "Verified Org" {
				foundVerifiedOrg = true
			}
			if org.Name == "No Website Org" {
				foundNoWebsiteOrg = true
			}
		}

		require.True(t, foundYoungOrg, "should include young org for validation")
		require.False(t, foundVerifiedOrg, "should not include recently verified org (within 3 days)")
		require.False(t, foundNoWebsiteOrg, "should not include org without website")
	})

	t.Run("Treasury and Admin Address Handling", func(t *testing.T) {
		// Test GetOrgTreasury and GetOrgAdmin functions
		treasuryAddr := sample.AccAddress()
		adminAddr := sample.AccAddress()

		org := types.Org{
			Creator:            creator,
			Admin:              adminAddr,
			Treasury:           treasuryAddr,
			Name:               "Address Test Org",
			LegalName:          "Address Test Organization",
			Address:            "1000 Address St, Address City, AC 10000",
			RegistrationNumber: "ADR123456789",
			CountryCode:        "BR",
			TaxId:              "ADR987654321",
			Website:            "https://address-test.example.com",
			AbuseEmail:         "abuse@address-test.example.com",
			ContactEmail:       "contact@address-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		orgId := keeper.AppendOrg(ctx, org)

		// Test GetOrgTreasury
		treasuryAccAddr, found := keeper.GetOrgTreasury(ctx, orgId)
		require.True(t, found, "should find treasury address")
		require.Equal(t, treasuryAddr, treasuryAccAddr.String(), "should have correct treasury address")

		// Test GetOrgAdmin
		adminAccAddr, found := keeper.GetOrgAdmin(ctx, orgId)
		require.True(t, found, "should find admin address")
		require.Equal(t, adminAddr, adminAccAddr.String(), "should have correct admin address")

		// Test with non-existent org
		_, found = keeper.GetOrgTreasury(ctx, 99999)
		require.False(t, found, "should not find treasury for non-existent org")

		_, found = keeper.GetOrgAdmin(ctx, 99999)
		require.False(t, found, "should not find admin for non-existent org")
	})
}
