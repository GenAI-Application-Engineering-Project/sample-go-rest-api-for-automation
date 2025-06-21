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

const (
	testMinLimit = 10
	testMaxLimit = 1000
)

var testProductOne = Product{
	ID:          uuid.MustParse("f2aa335f-6f91-4d4d-8057-53b0009bc376"),
	Name:        "Test Product A",
	Description: "Test product a description",
	ImageURL:    "test/image/url",
	CategoryID:  uuid.MustParse("0c34eab4-2d9d-4755-8c4d-dbfbac6728e8"),
	Price:       234.85,
	Quantity:    20,
	CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
}

var testProductTwo = Product{
	ID:          uuid.MustParse("b12f2176-28ca-4acf-85b9-cc97ca1b3cf6"),
	Name:        "Test Product B",
	Description: "Test product B description",
	CategoryID:  uuid.MustParse("9fcceb36-8a46-404f-9ce6-047c3fb65617"),
	Price:       234.85,
	Quantity:    1543,
	CreatedAt:   time.Date(2025, 10, 13, 0, 0, 0, 0, time.UTC),
}

func TestGetProductByID(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewProductRepo(db, testMinLimit, testMaxLimit)
	ctx := t.Context()

	selectQuery := regexp.QuoteMeta(
		`SELECT id, name, description, image_url, category_id, price, quantity, created_at
		FROM products
		WHERE id = $1`,
	)
	t.Run("should return product", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "image_url", "category_id", "price", "quantity", "created_at"}).
			AddRow(testProductOne.ID, testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt)
		mock.ExpectQuery(selectQuery).WithArgs(testProductOne.ID).WillReturnRows(mockRows)
		product, err := repo.GetProductByID(ctx, testProductOne.ID)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, &testProductOne, product)
	})

	t.Run("should return error if select query error", func(t *testing.T) {
		dbErr := errors.New("query error")
		mock.ExpectQuery(selectQuery).WithArgs(testProductOne.ID).WillReturnError(dbErr)
		product, err := repo.GetProductByID(ctx, testProductOne.ID)
		assert.Error(t, err)
		assert.Nil(t, product)
		expectedErrMsg := "getProductByID: select query failed: query error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return error if no row", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "created_at"})
		mock.ExpectQuery(selectQuery).WithArgs(testProductOne.ID).WillReturnRows(mockRows)
		product, err := repo.GetProductByID(ctx, testProductOne.ID)
		assert.Nil(t, product)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
		expectedErrMsg := "getProductByID: resource not found: id `f2aa335f-6f91-4d4d-8057-53b0009bc376`"
		assert.Equal(t, expectedErrMsg, err.Error())
	})
}

func TestListProducts(t *testing.T) {
	var createdAfter time.Time
	limit := 10

	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewProductRepo(db, testMinLimit, testMaxLimit)
	ctx := t.Context()

	selectQuery := regexp.QuoteMeta(`
			SELECT id, name, description, image_url, category_id, price, quantity, created_at
			FROM products
			WHERE created_at > ?
			ORDER BY created_at ASC
			LIMIT ?
		`)

	t.Run("should return list of products", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "image_url", "category_id", "price", "quantity", "created_at"}).
			AddRow(testProductOne.ID, testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt).
			AddRow(testProductTwo.ID, testProductTwo.Name, testProductTwo.Description, testProductTwo.ImageURL, testProductTwo.CategoryID, testProductTwo.Price, testProductTwo.Quantity, testProductTwo.CreatedAt)

		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, limit).WillReturnRows(mockRows)
		products, err := repo.ListProducts(ctx, createdAfter, limit)

		assert.NoError(t, err)
		assert.NotNil(t, products)
		assert.Equal(t, []*Product{&testProductOne, &testProductTwo}, products)
	})

	t.Run("should use minimum limit if limit is less than minimum limit", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "image_url", "category_id", "price", "quantity", "created_at"}).
			AddRow(testProductOne.ID, testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt).
			AddRow(testProductTwo.ID, testProductTwo.Name, testProductTwo.Description, testProductTwo.ImageURL, testProductTwo.CategoryID, testProductTwo.Price, testProductTwo.Quantity, testProductTwo.CreatedAt)

		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, 10).WillReturnRows(mockRows)
		products, err := repo.ListProducts(ctx, createdAfter, -1)

		assert.NoError(t, err)
		assert.NotNil(t, products)
		assert.Equal(t, []*Product{&testProductOne, &testProductTwo}, products)
	})

	t.Run("should use maximum limit if limit is greater than maximum limit", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "image_url", "category_id", "price", "quantity", "created_at"}).
			AddRow(testProductOne.ID, testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt).
			AddRow(testProductTwo.ID, testProductTwo.Name, testProductTwo.Description, testProductTwo.ImageURL, testProductTwo.CategoryID, testProductTwo.Price, testProductTwo.Quantity, testProductTwo.CreatedAt)

		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, 1000).WillReturnRows(mockRows)
		products, err := repo.ListProducts(ctx, createdAfter, 100009)

		assert.NoError(t, err)
		assert.NotNil(t, products)
		assert.Equal(t, []*Product{&testProductOne, &testProductTwo}, products)
	})

	t.Run("should return empty list if products length is zero", func(t *testing.T) {
		mockRows := sqlmock.NewRows(
			[]string{
				"id",
				"name",
				"description",
				"image_url",
				"category_id",
				"price",
				"quantity",
				"created_at",
			},
		)
		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, limit).WillReturnRows(mockRows)
		products, err := repo.ListProducts(ctx, createdAfter, limit)

		assert.NoError(t, err)
		assert.NotNil(t, products)
		assert.Equal(t, []*Product{}, products)
	})

	t.Run("should return error if select query fails", func(t *testing.T) {
		dbErr := errors.New("query error")
		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, limit).WillReturnError(dbErr)
		products, err := repo.ListProducts(ctx, createdAfter, limit)

		assert.Nil(t, products)
		assert.Error(t, err)
		expectedErrMsg := "listProducts: select query failed: query error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return error if scan fails", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"id", "name", "description", "createdAt"}).
			AddRow(testProductOne.ID, testProductOne.Name, testProductOne.Description, testProductOne.CreatedAt).
			AddRow(testProductTwo.ID, testProductTwo.Name, testProductTwo.Description, testProductTwo.CreatedAt)

		mock.ExpectQuery(selectQuery).WithArgs(createdAfter, limit).WillReturnRows(mockRows)
		products, err := repo.ListProducts(ctx, createdAfter, limit)

		assert.Nil(t, products)
		assert.Error(t, err)
		expectedErrMsg := "listProducts: scan failed: missing destination name createdAt in *datalayer.Product"
		assert.Equal(t, expectedErrMsg, err.Error())
	})
}

func TestCreateProduct(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewProductRepo(db, testMinLimit, testMaxLimit)
	ctx := t.Context()

	insertQuery := regexp.QuoteMeta(
		`INSERT INTO products(id, name, description, image_url, category_id, price, quantity, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	t.Run("should create valid product", func(t *testing.T) {
		mock.ExpectExec(insertQuery).
			WithArgs(testProductOne.ID, testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateProduct(ctx, &testProductOne)
		assert.NoError(t, err)
	})

	t.Run("should return error if insert query fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mock.ExpectExec(insertQuery).
			WithArgs(testProductOne.ID, testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt).
			WillReturnError(dbErr)

		err := repo.CreateProduct(ctx, &testProductOne)
		assert.Error(t, err)
		expectedErrMsg := "createProduct: insert query failed: database error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return not found if no rows affected", func(t *testing.T) {
		mock.ExpectExec(insertQuery).
			WithArgs(testProductOne.ID, testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.CreateProduct(ctx, &testProductOne)
		assert.Error(t, err)
		expectedErrMsg := "createProduct: no rows affected: resource not found"
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return error if rows affected fails", func(t *testing.T) {
		dbErr := errors.New("rows affected error")
		mock.ExpectExec(insertQuery).
			WithArgs(testProductOne.ID, testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt).
			WillReturnResult(sqlmock.NewErrorResult(dbErr))

		err := repo.CreateProduct(ctx, &testProductOne)
		assert.Error(t, err)
		expectedErrMsg := "createProduct: failed to get rows affected: rows affected error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})
}

func TestUpdateProduct(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewProductRepo(db, testMinLimit, testMaxLimit)
	ctx := t.Context()

	updateQuery := regexp.QuoteMeta(
		`UPDATE products SET name=?, description=?, image_url=?,category_id=?, price=?, quantity=?, created_at=? WHERE id=?`,
	)

	t.Run("should update valid product", func(t *testing.T) {
		mock.ExpectExec(updateQuery).
			WithArgs(testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt, testProductOne.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateProduct(ctx, &testProductOne)
		assert.NoError(t, err)
	})

	t.Run("should return error if update query fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mock.ExpectExec(updateQuery).
			WithArgs(testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt, testProductOne.ID).
			WillReturnError(dbErr)

		err := repo.UpdateProduct(ctx, &testProductOne)
		assert.Error(t, err)
		expectedErrMsg := "updateProduct: update query failed: database error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return not found if no rows affected", func(t *testing.T) {
		mock.ExpectExec(updateQuery).
			WithArgs(testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt, testProductOne.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateProduct(ctx, &testProductOne)
		assert.Error(t, err)
		expectedErrMsg := "updateProduct: no rows affected: resource not found"
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return error if rows affected fails", func(t *testing.T) {
		dbErr := errors.New("rows affected error")
		mock.ExpectExec(updateQuery).
			WithArgs(testProductOne.Name, testProductOne.Description, testProductOne.ImageURL, testProductOne.CategoryID, testProductOne.Price, testProductOne.Quantity, testProductOne.CreatedAt, testProductOne.ID).
			WillReturnResult(sqlmock.NewErrorResult(dbErr))

		err := repo.UpdateProduct(ctx, &testProductOne)
		assert.Error(t, err)
		expectedErrMsg := "updateProduct: failed to get rows affected: rows affected error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})
}

func TestDeleteProduct(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewProductRepo(db, testMinLimit, testMaxLimit)
	ctx := t.Context()

	deleteQuery := regexp.QuoteMeta(`DELETE FROM products WHERE id = $1`)

	t.Run("should delete valid product", func(t *testing.T) {
		mock.ExpectExec(deleteQuery).
			WithArgs(testProductOne.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.DeleteProduct(ctx, testProductOne.ID)
		assert.NoError(t, err)
	})

	t.Run("should return error if delete query fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mock.ExpectExec(deleteQuery).WithArgs(testProductOne.ID).WillReturnError(dbErr)

		err := repo.DeleteProduct(ctx, testProductOne.ID)
		assert.Error(t, err)
		expectedErrMsg := "deleteProduct: delete query failed: database error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return not found if no rows affected", func(t *testing.T) {
		mock.ExpectExec(deleteQuery).
			WithArgs(testProductOne.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteProduct(ctx, testProductOne.ID)
		assert.Error(t, err)
		expectedErrMsg := "deleteProduct: no rows affected: resource not found"
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("should return error if rows affected fails", func(t *testing.T) {
		dbErr := errors.New("rows affected error")
		mock.ExpectExec(deleteQuery).
			WithArgs(testProductOne.ID).WillReturnResult(sqlmock.NewErrorResult(dbErr))

		err := repo.DeleteProduct(ctx, testProductOne.ID)
		assert.Error(t, err)
		expectedErrMsg := "deleteProduct: failed to get rows affected: rows affected error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})
}
