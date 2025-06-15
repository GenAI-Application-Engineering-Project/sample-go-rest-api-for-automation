package datalayer

import (
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
	ctx := t.Context()

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
	ctx := t.Context()

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

	t.Run("should use minimum limit if limit is less than minimum limit", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "created_at"}).
			AddRow(testCategoryOne.ID, testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.CreatedAt).
			AddRow(testCategoryTwo.ID, testCategoryTwo.Name, testCategoryTwo.Description, testCategoryTwo.CreatedAt)

		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, 1).WillReturnRows(mockRows)
		categories, err := repo.ListCategories(ctx, createdAfter, -1)

		assert.NoError(t, err)
		assert.NotNil(t, categories)
		assert.Equal(t, []*Category{&testCategoryOne, &testCategoryTwo}, categories)
	})

	t.Run("should use maximum limit if limit is greater than maximum limit", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "created_at"}).
			AddRow(testCategoryOne.ID, testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.CreatedAt).
			AddRow(testCategoryTwo.ID, testCategoryTwo.Name, testCategoryTwo.Description, testCategoryTwo.CreatedAt)

		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, 1000).WillReturnRows(mockRows)
		categories, err := repo.ListCategories(ctx, createdAfter, 100009)

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

func TestCreateCategory(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewCategoryRepo(db)
	ctx := t.Context()

	insertQuery := regexp.QuoteMeta(
		`INSERT INTO categories(id, name, description, created_at) VALUES(?, ?, ?, ?)`,
	)

	t.Run("should create valid category", func(t *testing.T) {
		mock.ExpectExec(insertQuery).
			WithArgs(testCategoryOne.ID, testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateCategory(ctx, &testCategoryOne)
		assert.NoError(t, err)
	})

	t.Run("should return error if insert query fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mock.ExpectExec(insertQuery).
			WithArgs(testCategoryOne.ID, testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.CreatedAt).
			WillReturnError(dbErr)

		err := repo.CreateCategory(ctx, &testCategoryOne)
		assert.Error(t, err)
		expectedErrMsg := "createCategory: insert query failed: database error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return not found if no rows affected", func(t *testing.T) {
		mock.ExpectExec(insertQuery).
			WithArgs(testCategoryOne.ID, testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.CreatedAt).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.CreateCategory(ctx, &testCategoryOne)
		assert.Error(t, err)
		expectedErrMsg := "createCategory: no rows affected: not found"
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return error if rows affected fails", func(t *testing.T) {
		dbErr := errors.New("rows affected error")
		mock.ExpectExec(insertQuery).
			WithArgs(testCategoryOne.ID, testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.CreatedAt).
			WillReturnResult(sqlmock.NewErrorResult(dbErr))

		err := repo.CreateCategory(ctx, &testCategoryOne)
		assert.Error(t, err)
		expectedErrMsg := "createCategory: failed to get rows affected: rows affected error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})
}

func TestUpdateCategory(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewCategoryRepo(db)
	ctx := t.Context()

	updateQuery := regexp.QuoteMeta(`UPDATE categories SET name=?, description=? WHERE id=?`)

	t.Run("should update valid category", func(t *testing.T) {
		mock.ExpectExec(updateQuery).
			WithArgs(testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateCategory(ctx, &testCategoryOne)
		assert.NoError(t, err)
	})

	t.Run("should return error if update query fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mock.ExpectExec(updateQuery).
			WithArgs(testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.ID).
			WillReturnError(dbErr)

		err := repo.UpdateCategory(ctx, &testCategoryOne)
		assert.Error(t, err)
		expectedErrMsg := "updateCategory: update query failed: database error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return not found if no rows affected", func(t *testing.T) {
		mock.ExpectExec(updateQuery).
			WithArgs(testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateCategory(ctx, &testCategoryOne)
		assert.Error(t, err)
		expectedErrMsg := "updateCategory: no rows affected: not found"
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return error if rows affected fails", func(t *testing.T) {
		dbErr := errors.New("rows affected error")
		mock.ExpectExec(updateQuery).
			WithArgs(testCategoryOne.Name, testCategoryOne.Description, testCategoryOne.ID).
			WillReturnResult(sqlmock.NewErrorResult(dbErr))

		err := repo.UpdateCategory(ctx, &testCategoryOne)
		assert.Error(t, err)
		expectedErrMsg := "updateCategory: failed to get rows affected: rows affected error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})
}

func TestDeleteCategory(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewCategoryRepo(db)
	ctx := t.Context()

	deleteQuery := regexp.QuoteMeta(`DELETE FROM categories WHERE id = $1`)

	t.Run("should delete valid category", func(t *testing.T) {
		mock.ExpectExec(deleteQuery).
			WithArgs(testCategoryOne.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.DeleteCategory(ctx, testCategoryOne.ID)
		assert.NoError(t, err)
	})

	t.Run("should return error if delete query fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mock.ExpectExec(deleteQuery).WithArgs(testCategoryOne.ID).WillReturnError(dbErr)

		err := repo.DeleteCategory(ctx, testCategoryOne.ID)
		assert.Error(t, err)
		expectedErrMsg := "deleteCategory: delete query failed: database error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return not found if no rows affected", func(t *testing.T) {
		mock.ExpectExec(deleteQuery).
			WithArgs(testCategoryOne.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteCategory(ctx, testCategoryOne.ID)
		assert.Error(t, err)
		expectedErrMsg := "deleteCategory: no rows affected: not found"
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return error if rows affected fails", func(t *testing.T) {
		dbErr := errors.New("rows affected error")
		mock.ExpectExec(deleteQuery).
			WithArgs(testCategoryOne.ID).WillReturnResult(sqlmock.NewErrorResult(dbErr))

		err := repo.DeleteCategory(ctx, testCategoryOne.ID)
		assert.Error(t, err)
		expectedErrMsg := "deleteCategory: failed to get rows affected: rows affected error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})
}
