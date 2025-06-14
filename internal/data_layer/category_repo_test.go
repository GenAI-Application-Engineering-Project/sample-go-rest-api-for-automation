package datalayer

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

var testCategory = Category{
	ID:          uuid.MustParse("f2aa335f-6f91-4d4d-8057-53b0009bc376"),
	Name:        "Test Category A",
	Description: "Test category a description",
}

func TestGetCategoryByID(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewCategoryRepo(db)
	ctx := context.Background()

	selectQuery := regexp.QuoteMeta(`SELECT id, name, description FROM categories WHERE id = $1`)
	t.Run("should return category id", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description"}).
			AddRow(testCategory.ID, testCategory.Name, testCategory.Description)
		mock.ExpectQuery(selectQuery).WithArgs(testCategory.ID).WillReturnRows(mockRows)
		category, err := repo.GetCategoryByID(ctx, testCategory.ID)
		assert.NoError(t, err)
		assert.NotNil(t, category)
		assert.Equal(t, &testCategory, category)
	})

	t.Run("should fail if query error", func(t *testing.T) {
		dbErr := errors.New("query error")
		mock.ExpectQuery(selectQuery).WithArgs(testCategory.ID).WillReturnError(dbErr)
		category, err := repo.GetCategoryByID(ctx, testCategory.ID)
		assert.Error(t, err)
		assert.Nil(t, category)
		expectedErrMsg := "getCategoryByID: select query failed: query error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should fail if no row", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description"})
		mock.ExpectQuery(selectQuery).WithArgs(testCategory.ID).WillReturnRows(mockRows)
		category, err := repo.GetCategoryByID(ctx, testCategory.ID)
		assert.Nil(t, category)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
		expectedErrMsg := "getCategoryByID: not found: id `f2aa335f-6f91-4d4d-8057-53b0009bc376`"
		assert.Equal(t, expectedErrMsg, err.Error())
	})
}
