package datalayer

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

var testCategoryOne = Category{
	ID:          uuid.MustParse("f2aa335f-6f91-4d4d-8057-53b0009bc376"),
	Name:        "Test Category A",
	Description: "Test category a description",
	CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
}

var testCategoryTwo = Category{
	ID:          uuid.MustParse("b12f2176-28ca-4acf-85b9-cc97ca1b3cf6"),
	Name:        "Test Category B",
	Description: "Test category B description",
	CreatedAt:   time.Date(2025, 10, 13, 0, 0, 0, 0, time.UTC),
}

func TestGetCategoryByID(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewCategoryRepo(db)
	ctx := context.Background()

	selectQuery := regexp.QuoteMeta(`SELECT id, name, description FROM categories WHERE id = $1`)
	t.Run("should return category", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "created_at"}).
			AddRow(testCategoryOne.ID, testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.CreatedAt)
		mock.ExpectQuery(selectQuery).WithArgs(testCategoryOne.ID).WillReturnRows(mockRows)
		category, err := repo.GetCategoryByID(ctx, testCategoryOne.ID)
		assert.NoError(t, err)
		assert.NotNil(t, category)
		assert.Equal(t, &testCategoryOne, category)
	})

	t.Run("should return error if select query error", func(t *testing.T) {
		dbErr := errors.New("query error")
		mock.ExpectQuery(selectQuery).WithArgs(testCategoryOne.ID).WillReturnError(dbErr)
		category, err := repo.GetCategoryByID(ctx, testCategoryOne.ID)
		assert.Error(t, err)
		assert.Nil(t, category)
		expectedErrMsg := "getCategoryByID: select query failed: query error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return error if no row", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "created_at"})
		mock.ExpectQuery(selectQuery).WithArgs(testCategoryOne.ID).WillReturnRows(mockRows)
		category, err := repo.GetCategoryByID(ctx, testCategoryOne.ID)
		assert.Nil(t, category)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
		expectedErrMsg := "getCategoryByID: not found: id `f2aa335f-6f91-4d4d-8057-53b0009bc376`"
		assert.Equal(t, expectedErrMsg, err.Error())
	})
}

func TestListCategories(t *testing.T) {
	var createdAfter time.Time
	limit := 10

	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewCategoryRepo(db)
	ctx := context.Background()

	selectQuery := regexp.QuoteMeta(`
			SELECT id, name, description, created_at
			FROM categories
			WHERE created_at > ?
			ORDER BY created_at ASC
			LIMIT ?
		`)

	t.Run("should return list of categories", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "created_at"}).
			AddRow(testCategoryOne.ID, testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.CreatedAt).
			AddRow(testCategoryTwo.ID, testCategoryTwo.Name, testCategoryTwo.Description, testCategoryTwo.CreatedAt)

		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, limit).WillReturnRows(mockRows)
		categories, err := repo.ListCategories(ctx, createdAfter, limit)

		assert.NoError(t, err)
		assert.NotNil(t, categories)
		assert.Equal(t, []*Category{&testCategoryOne, &testCategoryTwo}, categories)
	})

	t.Run("should return empty list if categories length is zero", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "created_at"})
		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, limit).WillReturnRows(mockRows)
		categories, err := repo.ListCategories(ctx, createdAfter, limit)

		assert.NoError(t, err)
		assert.NotNil(t, categories)
		assert.Equal(t, []*Category{}, categories)
	})

	t.Run("should return error if select query fails", func(t *testing.T) {
		dbErr := errors.New("query error")
		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, limit).WillReturnError(dbErr)
		categories, err := repo.ListCategories(ctx, createdAfter, limit)

		assert.Nil(t, categories)
		assert.Error(t, err)
		expectedErrMsg := "listCategories: select query failed: query error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return error if scan fails", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "createdAt"}).
			AddRow(testCategoryOne.ID, testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.CreatedAt).
			AddRow(testCategoryTwo.ID, testCategoryTwo.Name, testCategoryTwo.Description, testCategoryTwo.CreatedAt)

		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, limit).WillReturnRows(mockRows)
		categories, err := repo.ListCategories(ctx, createdAfter, limit)

		assert.Nil(t, categories)
		assert.Error(t, err)
		expectedErrMsg := "listCategories: scan failed: missing destination name createdAt in *datalayer.Category"
		assert.Equal(t, expectedErrMsg, err.Error())
	})
}
