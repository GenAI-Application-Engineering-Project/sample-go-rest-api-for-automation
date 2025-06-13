package datalayer

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	getCategoryByID = "GetCategoryByID"
	listCategories  = "ListCategories"
	createCategory  = "CreateCategory"
	updateCategory  = "UpdateCategory"
	deleteCategory  = "DeleteCategory"
)

type Category struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

type CategoryRepo struct {
	db *sqlx.DB
}

type CategoryRepoInterface interface {
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error)
	ListCategories(ctx context.Context, id uuid.UUID, limit int) ([]*Category, error)
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
	query := `SELECT id, name, description FROM categories WHERE id = :id`
	params := map[string]any{"id": id}

	stmt, err := r.db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf(errMsg, getCategoryByID, errDBFailure, err)
	}
	defer stmt.Close()

	var category Category
	if stmt.Next() {
		if err := stmt.StructScan(&category); err != nil {
			return nil, fmt.Errorf(errMsg, getCategoryByID, errDBFailure, err)
		}
		return &category, nil
	}
	return nil, fmt.Errorf(errMsg, getCategoryByID, errNotFound, nil)
}

// ListCategories fetches all categories from the database
func (r *CategoryRepo) ListCategories(
	ctx context.Context,
	id uuid.UUID,
	limit int,
) ([]*Category, error) {
	limit = checkLimit(limit)
	args := map[string]any{
		"id":    id,
		"limit": limit,
	}

	query := `SELECT id, name, description FROM categories WHERE id > :id ORDER BY id ASC LIMIT :limit`
	stmt, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf(errMsg, listCategories, errDBFailure, err)
	}
	defer stmt.Close()

	var categories []*Category
	for stmt.Next() {
		var category Category
		if err := stmt.StructScan(&category); err != nil {
			return nil, fmt.Errorf(errMsg, listCategories, errDBFailure, err)
		}
		categories = append(categories, &category)
	}
	return categories, nil
}

// CreateCategory inserts a new category into the database
func (r *CategoryRepo) CreateCategory(ctx context.Context, category *Category) error {
	query := `INSERT INTO categories(id, name, description) VALUES(:id, :name, :description)`
	result, err := r.db.NamedExecContext(ctx, query, category)
	if err != nil {
		return fmt.Errorf(errMsg, createCategory, errDBFailure, err)
	}
	return checkRowsAffected(result, createCategory)
}

// UpdateCategory modifies an existing category
func (r *CategoryRepo) UpdateCategory(ctx context.Context, category *Category) error {
	query := `UPDATE categories SET name=:name, description=:description WHERE id=:id`
	result, err := r.db.NamedExecContext(ctx, query, category)
	if err != nil {
		return fmt.Errorf(errMsg, updateCategory, errDBFailure, err)
	}
	return checkRowsAffected(result, updateCategory)
}

// DeleteCategory removes a category by its ID
func (r *CategoryRepo) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM categories WHERE id=:id`
	params := map[string]any{"id": id}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf(errMsg, deleteCategory, errDBFailure, err)
	}
	return checkRowsAffected(result, deleteCategory)
}
