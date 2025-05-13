package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

func TestUpsertAccountMapping(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer db.Close()

	t.Run("Insert new account mapping", func(t *testing.T) {
		// Create a new mapping
		am := &models.AccountMapping{
			ExternalName: "test_account_1",
			LunchMoneyId: 100,
			IsPlaid:      true,
		}

		err := db.UpsertAccountMapping(am)
		assert.NoError(t, err)

		// Verify it was inserted correctly
		result, err := db.GetAccountMapping("test_account_1")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, am.ExternalName, result.ExternalName)
		assert.Equal(t, am.LunchMoneyId, result.LunchMoneyId)
		assert.Equal(t, am.IsPlaid, result.IsPlaid)
	})

	t.Run("Update existing account mapping", func(t *testing.T) {
		// First create a mapping
		am := &models.AccountMapping{
			ExternalName: "test_account_2",
			LunchMoneyId: 200,
			IsPlaid:      false,
		}

		err := db.UpsertAccountMapping(am)
		assert.NoError(t, err)

		// Now update it
		updatedAm := &models.AccountMapping{
			ExternalName: "test_account_2",
			LunchMoneyId: 201,
			IsPlaid:      true,
		}

		err = db.UpsertAccountMapping(updatedAm)
		assert.NoError(t, err)

		// Verify it was updated
		result, err := db.GetAccountMapping("test_account_2")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, updatedAm.ExternalName, result.ExternalName)
		assert.Equal(t, updatedAm.LunchMoneyId, result.LunchMoneyId)
		assert.Equal(t, updatedAm.IsPlaid, result.IsPlaid)
	})

	t.Run("Ignore account mapping with LunchMoneyId = -1", func(t *testing.T) {
		// Create an ignored account mapping
		am := &models.AccountMapping{
			ExternalName: "test_ignored_account",
			LunchMoneyId: -1,
			IsPlaid:      false,
		}

		err := db.UpsertAccountMapping(am)
		assert.NoError(t, err)
		assert.Less(t, am.LunchMoneyId, int64(0), "LunchMoneyId should be negative")

		// Verify it was saved as an ignored account
		result, err := db.GetAccountMapping("test_ignored_account")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, am.ExternalName, result.ExternalName)
		assert.Equal(t, am.LunchMoneyId, result.LunchMoneyId)
		assert.Equal(t, am.IsPlaid, result.IsPlaid)

		// Verify it's in the ignored_external_accounts table
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM ignored_external_accounts WHERE external_name = ?", am.ExternalName).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Update ignored account mapping", func(t *testing.T) {
		// First create an ignored mapping
		am := &models.AccountMapping{
			ExternalName: "test_ignored_to_normal",
			LunchMoneyId: -1,
			IsPlaid:      false,
		}

		err := db.UpsertAccountMapping(am)
		assert.NoError(t, err)

		// Now update it to non-ignored
		updatedAm := &models.AccountMapping{
			ExternalName: "test_ignored_to_normal",
			LunchMoneyId: 300,
			IsPlaid:      true,
		}

		err = db.UpsertAccountMapping(updatedAm)
		assert.NoError(t, err)

		// Verify it was updated
		result, err := db.GetAccountMapping("test_ignored_to_normal")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, updatedAm.ExternalName, result.ExternalName)
		assert.Equal(t, updatedAm.LunchMoneyId, result.LunchMoneyId)
		assert.Equal(t, updatedAm.IsPlaid, result.IsPlaid)
	})
}
