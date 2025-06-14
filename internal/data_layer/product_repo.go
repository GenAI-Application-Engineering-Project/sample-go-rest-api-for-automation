package datalayer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Product struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	ImageURL    string    `db:"image_url"`
	CategoryID  uuid.UUID `db:"category_id"`
	Price       float64   `db:"price"`
	Quantity    int       `db:"quantity"`
	CreatedAt   time.Time `db:"created_at"`
}

type ProductRepo struct {
	db *sqlx.DB
}

type ProductRepoInterface interface {
	GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error)
	ListProducts(ctx context.Context, createdAfter time.Time, limit int) ([]*Product, error)
	CreateProduct(ctx context.Context, category *Product) error
	UpdateProduct(ctx context.Context, category *Product) error
	DeleteProduct(ctx context.Context, id uuid.UUID) error
}

// NewProductRepository creates a new repository instance
func NewProductRepo(db *sqlx.DB) ProductRepoInterface {
	return &ProductRepo{db: db}
}

// GetProductByID fetches a product by its ID
func (r *ProductRepo) GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	const query = `
		SELECT id, name, description, image_url, category_id, price, quantity, created_at
		FROM products
		WHERE id = $1`

	var product Product
	err := r.db.GetContext(ctx, &product, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("getProductByID: %w: id `%s`", ErrNotFound, id)
		}
		return nil, fmt.Errorf("getProductByID: select query failed: %w", err)
	}

	return &product, nil
}

// ListProducts fetches all products from the database
func (r *ProductRepo) ListProducts(
	ctx context.Context,
	createdAfter time.Time, // pagination token
	limit int,
) ([]*Product, error) {
	limit = checkLimit(limit)
	args := map[string]any{
		"created_at": createdAfter,
		"limit":      limit,
	}

	const query = `
		SELECT id, name, description, image_url, category_id, price, quantity, created_at
		FROM products 
		WHERE created_at > :created_at 
		ORDER BY created_at ASC
		LIMIT :limit
	`

	stmt, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("listProducts: select query failed: %w", err)
	}
	defer stmt.Close()

	var products []*Product
	for stmt.Next() {
		var product Product
		if err := stmt.StructScan(&product); err != nil {
			return nil, fmt.Errorf("listProducts: scan failed: %w", err)
		}
		products = append(products, &product)
	}

	if len(products) == 0 {
		return []*Product{}, nil
	}

	return products, nil
}

// CreateProduct inserts a new product into the database
func (r *ProductRepo) CreateProduct(ctx context.Context, product *Product) error {
	const query = `
		INSERT INTO products(id, name, description, image_url, category_id, price, quantity, created_at) 
		VALUES(:id, :name, :description, :image_url, :category_id, :price, :quantity, :created_at)
	`
	result, err := r.db.NamedExecContext(ctx, query, product)
	if err != nil {
		return fmt.Errorf("createProduct: insert query failed: %w", err)
	}
	return checkRowsAffected(result, "createProduct")
}

// UpdateProduct modifies an existing product
func (r *ProductRepo) UpdateProduct(ctx context.Context, product *Product) error {
	const query = `
		UPDATE products
		SET name=:name, description=:description, image_url=:image_url,category_id=:category_id,
		price=:price, quantity=:quantity, created_at=:created_at
		WHERE id=:id
	`
	result, err := r.db.NamedExecContext(ctx, query, product)
	if err != nil {
		return fmt.Errorf("updateProduct: update query failed: %w", err)
	}
	return checkRowsAffected(result, "updateProduct")
}

// DeleteProduct removes a product by its ID
func (r *ProductRepo) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM products WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleteProduct: delete query failed: %w", err)
	}
	return checkRowsAffected(result, "deleteProduct")
}
