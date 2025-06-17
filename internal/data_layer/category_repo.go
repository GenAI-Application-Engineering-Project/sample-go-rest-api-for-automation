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

type Category struct {
	ID          uuid.UUID `json:"id"          db:"id"`
	Name        string    `json:"name"        db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt"   db:"created_at"`
}

type CategoryRepo struct {
	db *sqlx.DB
}

type CategoryRepoInterface interface {
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error)
	ListCategories(ctx context.Context, createdAfter time.Time, limit int) ([]*Category, error)
	CreateCategory(ctx context.Context, category *Category) error
	UpdateCategory(ctx context.Context, category *Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

// NewCategoryRepo creates a new repository instance
func NewCategoryRepo(db *sqlx.DB) CategoryRepoInterface {
	return &CategoryRepo{db: db}
}

// GetCategoryByID fetches a category by its ID
func (r *CategoryRepo) GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	const query = `SELECT id, name, description FROM categories WHERE id = $1`

	var category Category
	err := r.db.GetContext(ctx, &category, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("getCategoryByID: %w: id `%s`", ErrNotFound, id)
		}
		return nil, fmt.Errorf("getCategoryByID: select query failed: %w", err)
	}

	return &category, nil
}

// ListCategories fetches all categories from the database
func (r *CategoryRepo) ListCategories(
	ctx context.Context,
	createdAfter time.Time, // pagination cursor
	limit int,
) ([]*Category, error) {
	limit = checkLimit(limit)
	args := map[string]any{
		"created_at": createdAfter,
		"limit":      limit,
	}

	const query = `
		SELECT id, name, description, created_at
		FROM categories
		WHERE created_at > :created_at
		ORDER BY created_at ASC
		LIMIT :limit
	`

	stmt, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("listCategories: select query failed: %w", err)
	}
	defer stmt.Close()

	var categories []*Category
	for stmt.Next() {
		var category Category
		if err := stmt.StructScan(&category); err != nil {
			return nil, fmt.Errorf("listCategories: scan failed: %w", err)
		}
		categories = append(categories, &category)
	}

	if len(categories) == 0 {
		return []*Category{}, nil
	}

	return categories, nil
}

// CreateCategory inserts a new category into the database
func (r *CategoryRepo) CreateCategory(ctx context.Context, category *Category) error {
	const query = `INSERT INTO categories(id, name, description, created_at) VALUES(:id, :name, :description, :created_at)`
	result, err := r.db.NamedExecContext(ctx, query, category)
	if err != nil {
		return fmt.Errorf("createCategory: insert query failed: %w", err)
	}
	return checkRowsAffected(result, "createCategory")
}

// UpdateCategory modifies an existing category
func (r *CategoryRepo) UpdateCategory(ctx context.Context, category *Category) error {
	const query = `UPDATE categories SET name=:name, description=:description WHERE id=:id`
	result, err := r.db.NamedExecContext(ctx, query, category)
	if err != nil {
		return fmt.Errorf("updateCategory: update query failed: %w", err)
	}
	return checkRowsAffected(result, "updateCategory")
}

// DeleteCategory removes a category by its ID
func (r *CategoryRepo) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM categories WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleteCategory: delete query failed: %w", err)
	}
	return checkRowsAffected(result, "deleteCategory")
}
