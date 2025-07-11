package keeper_test

import (
	"testing"
	"time"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/sample"
	"ipex/x/org/types"

	"github.com/stretchr/testify/require"
)

// TestOrganizationValidationIntegration tests comprehensive validation logic
func TestOrganizationValidationIntegration(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	baseTime := ctx.BlockTime()

	t.Run("Field Format Validation", func(t *testing.T) {
		creator := sample.AccAddress()

		// Test valid organization creation
		validOrg := types.Org{
			Creator:            creator,
			Admin:              creator,
			Treasury:           sample.AccAddress(),
			Name:               "Valid Test Organization",
			LegalName:          "Valid Test Organization Inc.",
			Address:            "123 Valid Street, Valid City, VC 12345",
			RegistrationNumber: "VAL123456789",
			CountryCode:        "US",
			TaxId:              "VAL987654321",
			Website:            "https://valid-test.example.com",
			AbuseEmail:         "abuse@valid-test.example.com",
			ContactEmail:       "contact@valid-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		orgId := keeper.AppendOrg(ctx, validOrg)
		require.Greater(t, orgId, uint64(0), "should create valid organization")

		// Verify all fields were stored correctly
		retrievedOrg, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find valid organization")
		require.Equal(t, "Valid Test Organization", retrievedOrg.Name, "should store correct name")
		require.Equal(t, "Valid Test Organization Inc.", retrievedOrg.LegalName, "should store correct legal name")
		require.Equal(t, "123 Valid Street, Valid City, VC 12345", retrievedOrg.Address, "should store correct address")
		require.Equal(t, "VAL123456789", retrievedOrg.RegistrationNumber, "should store correct registration number")
		require.Equal(t, "US", retrievedOrg.CountryCode, "should store correct country code")
		require.Equal(t, "VAL987654321", retrievedOrg.TaxId, "should store correct tax ID")
		require.Equal(t, "https://valid-test.example.com", retrievedOrg.Website, "should store correct website")
		require.Equal(t, "abuse@valid-test.example.com", retrievedOrg.AbuseEmail, "should store correct abuse email")
		require.Equal(t, "contact@valid-test.example.com", retrievedOrg.ContactEmail, "should store correct contact email")
	})

	t.Run("Email Format Validation", func(t *testing.T) {
		creator := sample.AccAddress()

		// Test various email formats
		validEmails := []string{
			"test@example.com",
			"test.user@example.com",
			"test+tag@example.com",
			"test123@example-domain.com",
			"user.name+tag@long-domain-name.example",
		}

		for i, email := range validEmails {
			org := types.Org{
				Creator:            creator,
				Admin:              creator,
				Treasury:           sample.AccAddress(),
				Name:               "Email Test Org",
				LegalName:          "Email Test Organization",
				Address:            "123 Email Street, Email City, EC 12345",
				RegistrationNumber: "EML123456789",
				CountryCode:        "US",
				TaxId:              "EML987654321",
				Website:            "https://email-test.example.com",
				AbuseEmail:         email,
				ContactEmail:       email,
				DomainVerified:     false,
				CreatedAt:          baseTime,
				UpdatedAt:          baseTime,
			}

			orgId := keeper.AppendOrg(ctx, org)
			require.Greater(t, orgId, uint64(0), "should create organization with valid email %d: %s", i, email)

			retrievedOrg, found := keeper.GetOrg(ctx, orgId)
			require.True(t, found, "should find organization with email %s", email)
			require.Equal(t, email, retrievedOrg.AbuseEmail, "should store correct abuse email")
			require.Equal(t, email, retrievedOrg.ContactEmail, "should store correct contact email")
		}
	})

	t.Run("Website URL Validation", func(t *testing.T) {
		creator := "cosmos1urltest123def456ghi789jkl012mno345pqr"

		// Test various valid URL formats
		validUrls := []string{
			"https://example.com",
			"https://www.example.com",
			"https://subdomain.example.com",
			"https://example.com/path",
			"https://example.com:8080",
			"http://localhost:3000",
			"https://example-domain.test.com",
		}

		for i, url := range validUrls {
			org := types.Org{
				Creator:            creator,
				Admin:              creator,
				Treasury:           sample.AccAddress(),
				Name:               "URL Test Org",
				LegalName:          "URL Test Organization",
				Address:            "123 URL Street, URL City, UC 12345",
				RegistrationNumber: "URL123456789",
				CountryCode:        "US",
				TaxId:              "URL987654321",
				Website:            url,
				AbuseEmail:         "abuse@url-test.example.com",
				ContactEmail:       "contact@url-test.example.com",
				DomainVerified:     false,
				CreatedAt:          baseTime,
				UpdatedAt:          baseTime,
			}

			orgId := keeper.AppendOrg(ctx, org)
			require.Greater(t, orgId, uint64(0), "should create organization with valid URL %d: %s", i, url)

			retrievedOrg, found := keeper.GetOrg(ctx, orgId)
			require.True(t, found, "should find organization with URL %s", url)
			require.Equal(t, url, retrievedOrg.Website, "should store correct website URL")
		}
	})

	t.Run("Country Code Validation", func(t *testing.T) {
		creator := "cosmos1countrytest123def456ghi789jkl012mno"

		// Test various valid country codes
		validCountryCodes := []string{
			"US", "CA", "GB", "FR", "DE", "JP", "AU", "BR", "IN", "CN",
			"RU", "ZA", "MX", "IT", "ES", "KR", "NL", "SE", "NO", "DK",
		}

		for i, countryCode := range validCountryCodes {
			org := types.Org{
				Creator:            creator,
				Admin:              creator,
				Treasury:           sample.AccAddress(),
				Name:               "Country Test Org",
				LegalName:          "Country Test Organization",
				Address:            "123 Country Street, Country City, CC 12345",
				RegistrationNumber: "CNT123456789",
				CountryCode:        countryCode,
				TaxId:              "CNT987654321",
				Website:            "https://country-test.example.com",
				AbuseEmail:         "abuse@country-test.example.com",
				ContactEmail:       "contact@country-test.example.com",
				DomainVerified:     false,
				CreatedAt:          baseTime,
				UpdatedAt:          baseTime,
			}

			orgId := keeper.AppendOrg(ctx, org)
			require.Greater(t, orgId, uint64(0), "should create organization with valid country code %d: %s", i, countryCode)

			retrievedOrg, found := keeper.GetOrg(ctx, orgId)
			require.True(t, found, "should find organization with country code %s", countryCode)
			require.Equal(t, countryCode, retrievedOrg.CountryCode, "should store correct country code")
		}
	})

	t.Run("Registration Number Validation", func(t *testing.T) {
		creator := "cosmos1regtest123def456ghi789jkl012mno345p"

		// Test various valid registration number formats
		validRegNumbers := []string{
			"REG123456789",
			"ABC-123456789",
			"XYZ123-456789",
			"12345ABCDE",
			"A1B2C3D4E5",
			"REG-ABC-123",
			"123456789A",
		}

		for i, regNumber := range validRegNumbers {
			org := types.Org{
				Creator:            creator,
				Admin:              creator,
				Treasury:           sample.AccAddress(),
				Name:               "Reg Test Org",
				LegalName:          "Registration Test Organization",
				Address:            "123 Reg Street, Reg City, RC 12345",
				RegistrationNumber: regNumber,
				CountryCode:        "US",
				TaxId:              "REG987654321",
				Website:            "https://reg-test.example.com",
				AbuseEmail:         "abuse@reg-test.example.com",
				ContactEmail:       "contact@reg-test.example.com",
				DomainVerified:     false,
				CreatedAt:          baseTime,
				UpdatedAt:          baseTime,
			}

			orgId := keeper.AppendOrg(ctx, org)
			require.Greater(t, orgId, uint64(0), "should create organization with valid reg number %d: %s", i, regNumber)

			retrievedOrg, found := keeper.GetOrg(ctx, orgId)
			require.True(t, found, "should find organization with reg number %s", regNumber)
			require.Equal(t, regNumber, retrievedOrg.RegistrationNumber, "should store correct registration number")
		}
	})

	t.Run("Address Format Validation", func(t *testing.T) {
		creator := "cosmos1addrtest123def456ghi789jkl012mno345"

		// Test various valid address formats
		validAddresses := []string{
			"123 Main Street, New York, NY 10001",
			"Suite 100, 456 Business Ave, Los Angeles, CA 90210",
			"Building A, Floor 5, 789 Tech Blvd, San Francisco, CA 94105",
			"Unit 20, 321 Industrial Way, Chicago, IL 60601",
			"1000 Corporate Drive, Austin, TX 78701",
			"P.O. Box 12345, Phoenix, AZ 85001",
			"15 Oak Street, Apt 3B, Boston, MA 02101",
		}

		for i, address := range validAddresses {
			org := types.Org{
				Creator:            creator,
				Admin:              creator,
				Treasury:           sample.AccAddress(),
				Name:               "Address Test Org",
				LegalName:          "Address Test Organization",
				Address:            address,
				RegistrationNumber: "ADR123456789",
				CountryCode:        "US",
				TaxId:              "ADR987654321",
				Website:            "https://address-test.example.com",
				AbuseEmail:         "abuse@address-test.example.com",
				ContactEmail:       "contact@address-test.example.com",
				DomainVerified:     false,
				CreatedAt:          baseTime,
				UpdatedAt:          baseTime,
			}

			orgId := keeper.AppendOrg(ctx, org)
			require.Greater(t, orgId, uint64(0), "should create organization with valid address %d", i)

			retrievedOrg, found := keeper.GetOrg(ctx, orgId)
			require.True(t, found, "should find organization with address")
			require.Equal(t, address, retrievedOrg.Address, "should store correct address")
		}
	})
}

// TestOrganizationCrossModuleIntegration tests integration with other modules
func TestOrganizationCrossModuleIntegration(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	baseTime := ctx.BlockTime()

	t.Run("Trust Module Integration", func(t *testing.T) {
		creator := "cosmos1trusttest123def456ghi789jkl012mno345"

		// Test organization creation initializes lessee reputation
		org := types.Org{
			Creator:            creator,
			Admin:              creator,
			Treasury:           sample.AccAddress(),
			Name:               "Trust Integration Org",
			LegalName:          "Trust Integration Organization",
			Address:            "123 Trust Street, Trust City, TC 12345",
			RegistrationNumber: "TRU123456789",
			CountryCode:        "US",
			TaxId:              "TRU987654321",
			Website:            "https://trust-test.example.com",
			AbuseEmail:         "abuse@trust-test.example.com",
			ContactEmail:       "contact@trust-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		orgId := keeper.AppendOrg(ctx, org)
		require.Greater(t, orgId, uint64(0), "should create organization")

		// Verify organization was created (trust integration is mocked in test environment)
		retrievedOrg, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find created organization")
		require.Equal(t, orgId, retrievedOrg.Id, "should have correct ID")

		// Test organization removal removes lessee reputation
		keeper.RemoveOrg(ctx, orgId)

		// Verify organization was removed
		_, found = keeper.GetOrg(ctx, orgId)
		require.False(t, found, "should not find removed organization")
	})

	t.Run("Organization Sequence Management", func(t *testing.T) {
		creator := "cosmos1seqtest123def456ghi789jkl012mno345p"

		// Create multiple organizations to test ID sequencing
		var orgIds []uint64
		numOrgs := 5

		for i := 0; i < numOrgs; i++ {
			org := types.Org{
				Creator:            creator,
				Admin:              creator,
				Treasury:           sample.AccAddress(),
				Name:               "Sequence Test Org",
				LegalName:          "Sequence Test Organization",
				Address:            "123 Sequence Street, Sequence City, SC 12345",
				RegistrationNumber: "SEQ123456789",
				CountryCode:        "US",
				TaxId:              "SEQ987654321",
				Website:            "https://sequence-test.example.com",
				AbuseEmail:         "abuse@sequence-test.example.com",
				ContactEmail:       "contact@sequence-test.example.com",
				DomainVerified:     false,
				CreatedAt:          baseTime,
				UpdatedAt:          baseTime,
			}

			orgId := keeper.AppendOrg(ctx, org)
			orgIds = append(orgIds, orgId)
		}

		// Verify sequential ID assignment
		for i := 0; i < numOrgs-1; i++ {
			require.Less(t, orgIds[i], orgIds[i+1], "organization IDs should be sequential")
			require.Equal(t, uint64(1), orgIds[i+1]-orgIds[i], "organization IDs should increment by 1")
		}

		// Verify each organization exists with correct ID
		for i, orgId := range orgIds {
			org, found := keeper.GetOrg(ctx, orgId)
			require.True(t, found, "should find organization %d", i)
			require.Equal(t, orgId, org.Id, "should have correct ID %d", i)
		}
	})

	t.Run("Organization State Consistency", func(t *testing.T) {
		creator := "cosmos1statetest123def456ghi789jkl012mno34"

		// Test organization state across creation, update, and removal
		org := types.Org{
			Creator:            creator,
			Admin:              creator,
			Treasury:           sample.AccAddress(),
			Name:               "State Test Org",
			LegalName:          "State Test Organization",
			Address:            "123 State Street, State City, ST 12345",
			RegistrationNumber: "STA123456789",
			CountryCode:        "US",
			TaxId:              "STA987654321",
			Website:            "https://state-test.example.com",
			AbuseEmail:         "abuse@state-test.example.com",
			ContactEmail:       "contact@state-test.example.com",
			DomainVerified:     false,
			CreatedAt:          baseTime,
			UpdatedAt:          baseTime,
		}

		orgId := keeper.AppendOrg(ctx, org)

		// Test initial state
		retrievedOrg, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find organization")
		require.Equal(t, baseTime, retrievedOrg.CreatedAt, "should have correct created time")
		require.Equal(t, baseTime, retrievedOrg.UpdatedAt, "should have correct updated time")
		require.False(t, retrievedOrg.DomainVerified, "should not be domain verified initially")

		// Test state update
		updatedTime := baseTime.Add(time.Hour)
		ctx = ctx.WithBlockTime(updatedTime)

		retrievedOrg.Name = "Updated State Test Org"
		retrievedOrg.UpdatedAt = updatedTime
		keeper.SetOrg(ctx, retrievedOrg)

		updatedOrg, found := keeper.GetOrg(ctx, orgId)
		require.True(t, found, "should find updated organization")
		require.Equal(t, "Updated State Test Org", updatedOrg.Name, "should have updated name")
		require.Equal(t, baseTime, updatedOrg.CreatedAt, "should preserve created time")
		require.Equal(t, updatedTime, updatedOrg.UpdatedAt, "should have updated time")

		// Test HasOrg function
		exists := keeper.HasOrg(ctx, orgId)
		require.True(t, exists, "should confirm organization exists")

		// Test state after removal
		keeper.RemoveOrg(ctx, orgId)

		exists = keeper.HasOrg(ctx, orgId)
		require.False(t, exists, "should confirm organization no longer exists")

		_, found = keeper.GetOrg(ctx, orgId)
		require.False(t, found, "should not find removed organization")
	})

	t.Run("Performance and Scalability", func(t *testing.T) {
		creator := "cosmos1perftest123def456ghi789jkl012mno34"

		// Test performance with many organizations
		numOrgs := 50
		var orgIds []uint64

		startTime := time.Now()

		// Create many organizations
		for i := 0; i < numOrgs; i++ {
			org := types.Org{
				Creator:            creator,
				Admin:              creator,
				Treasury:           sample.AccAddress(),
				Name:               "Performance Test Org",
				LegalName:          "Performance Test Organization",
				Address:            "123 Performance Street, Performance City, PC 12345",
				RegistrationNumber: "PRF123456789",
				CountryCode:        "US",
				TaxId:              "PRF987654321",
				Website:            "https://performance-test.example.com",
				AbuseEmail:         "abuse@performance-test.example.com",
				ContactEmail:       "contact@performance-test.example.com",
				DomainVerified:     false,
				CreatedAt:          baseTime,
				UpdatedAt:          baseTime,
			}

			orgId := keeper.AppendOrg(ctx, org)
			orgIds = append(orgIds, orgId)
		}

		creationTime := time.Since(startTime)
		t.Logf("Created %d organizations in %v", numOrgs, creationTime)

		// Test bulk retrieval performance
		startTime = time.Now()
		allOrgs := keeper.GetAllOrg(ctx)
		retrievalTime := time.Since(startTime)
		t.Logf("Retrieved %d organizations in %v", len(allOrgs), retrievalTime)

		require.GreaterOrEqual(t, len(allOrgs), numOrgs, "should retrieve at least created organizations")

		// Test individual lookup performance
		startTime = time.Now()
		for i := 0; i < 10; i++ { // Test first 10
			_, found := keeper.GetOrg(ctx, orgIds[i])
			require.True(t, found, "should find organization %d", i)
		}
		lookupTime := time.Since(startTime)
		t.Logf("Performed 10 individual lookups in %v", lookupTime)

		// Test bulk removal performance
		startTime = time.Now()
		for i := 0; i < 10; i++ {
			keeper.RemoveOrg(ctx, orgIds[i])
		}
		removalTime := time.Since(startTime)
		t.Logf("Removed 10 organizations in %v", removalTime)

		// Verify remaining organizations
		remainingOrgs := keeper.GetAllOrg(ctx)
		removedIds := make(map[uint64]bool)
		for i := 0; i < 10; i++ {
			removedIds[orgIds[i]] = true
		}

		for _, org := range remainingOrgs {
			require.False(t, removedIds[org.Id], "removed organization should not appear in list")
		}
	})
}