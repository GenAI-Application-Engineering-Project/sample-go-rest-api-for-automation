package datalayer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	getProductByID = "GetProductByID"
	listProducts   = "ListProducts"
	createProduct  = "CreateProduct"
	updateProduct  = "UpdateProduct"
	deletedProduct = "DeleteProduct"
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
	ListProducts(ctx context.Context, id uuid.UUID, limit int) ([]*Product, error)
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
	query := `SELECT id, name, description, image_url, category_id, price, quantity, created_at FROM products WHERE id = :id`
	params := map[string]any{"id": id}

	stmt, err := r.db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf(errMsg, getProductByID, errDBFailure, err)
	}
	defer stmt.Close()

	var product Product
	if stmt.Next() {
		if err := stmt.StructScan(&product); err != nil {
			return nil, fmt.Errorf(errMsg, getProductByID, errDBFailure, err)
		}
		return &product, nil
	}

	// Product not found
	return nil, fmt.Errorf(errMsg, getProductByID, errNotFound, nil)
}

// ListProducts fetches all products from the database
func (r *ProductRepo) ListProducts(
	ctx context.Context,
	id uuid.UUID,
	limit int,
) ([]*Product, error) {
	limit = checkLimit(limit)
	args := map[string]any{
		"id":    id,
		"limit": limit,
	}

	query := `SELECT id, name, description, image_url, category_id, price, quantity, created_at FROM products WHERE id > :id ORDER BY id ASC LIMIT :limit`
	stmt, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf(errMsg, listProducts, errDBFailure, err)
	}
	defer stmt.Close()

	var products []*Product
	for stmt.Next() {
		var product Product
		if err := stmt.StructScan(&product); err != nil {
			return nil, fmt.Errorf(errMsg, listProducts, errDBFailure, err)
		}
		products = append(products, &product)
	}
	return products, nil
}

// CreateProduct inserts a new product into the database
func (r *ProductRepo) CreateProduct(ctx context.Context, product *Product) error {
	query := `INSERT INTO products(id, name, description, image_url, category_id, price, quantity, created_at) VALUES(:id, :name, :description, :image_url, :category_id, :price, :quantity, :created_at)`
	result, err := r.db.NamedExecContext(ctx, query, product)
	if err != nil {
		return fmt.Errorf(errMsg, createProduct, errDBFailure, err)
	}
	return checkRowsAffected(result, createCategory)
}

// UpdateProduct modifies an existing product
func (r *ProductRepo) UpdateProduct(ctx context.Context, product *Product) error {
	query := `UPDATE products SET name=:name, description=:description, image_url=:image_url,category_id=:category_id, price=:price, quantity=:quantity, created_at=:created_at WHERE id=:id`
	result, err := r.db.NamedExecContext(ctx, query, product)
	if err != nil {
		return fmt.Errorf(errMsg, updateProduct, errDBFailure, err)
	}
	return checkRowsAffected(result, updateCategory)
}

// DeleteProduct removes a product by its ID
func (r *ProductRepo) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id=:id`
	params := map[string]any{
		"id": id,
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf(errMsg, deletedProduct, errDBFailure, err)
	}
	return checkRowsAffected(result, deletedProduct)
}
