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
	db       *sqlx.DB
	minLimit int
	maxLimit int
}

type ListCategoryResult struct {
	Categories []*Category
	NextCursor time.Time
	HasMore    bool
	Error      error
}

type CategoryRepoInterface interface {
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error)
	ListCategories(ctx context.Context, createdAfter time.Time, limit int) ListCategoryResult
	CreateCategory(ctx context.Context, category *Category) error
	UpdateCategory(ctx context.Context, category *Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

// NewCategoryRepo creates a new repository instance
func NewCategoryRepo(db *sqlx.DB, minLimit, maxLimit int) CategoryRepoInterface {
	return &CategoryRepo{
		db:       db,
		minLimit: minLimit,
		maxLimit: maxLimit,
	}
}

// GetCategoryByID retrieves a single category from the database by its unique UUID.
//
// It executes a parameterized SQL query to prevent SQL injection, selecting the
// `id`, `name`, and `description` fields from the `categories` table where the
// `id` matches the given parameter.
//
// Parameters:
//   - ctx: A context used for request scoping, cancellation, and timeout.
//   - id: A UUID representing the unique identifier of the category.
//
// Returns:
//   - A pointer to the Category struct if found.
//   - An error if the query fails or the category does not exist.
//   - If no category is found, the returned error wraps ErrNotFound,
//     allowing callers to check with errors.Is(err, ErrNotFound).
func (r *CategoryRepo) GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	const getCategoryByIDQuery = `
		SELECT id, name, description, createdAt 
		FROM categories 
		WHERE id = $1
	`

	var category Category
	err := r.db.GetContext(ctx, &category, getCategoryByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("getCategoryByID: %w: id `%s`", ErrNotFound, id)
		}
		return nil, fmt.Errorf("getCategoryByID: select query failed: %w", err)
	}

	return &category, nil
}

// ListCategories retrieves a paginated list of categories from the database,
// ordered by creation time and ID in ascending order.
//
// Pagination is controlled using a time-based cursor (`createdAfter`) and a limit.
// To support cursor-based pagination, the query fetches one extra record beyond
// the specified limit to determine if more results are available.
//
// Parameters:
//   - ctx: the context for managing request lifetime and cancellation.
//   - createdAfter: only categories created at or after this time will be returned.
//   - limit: the maximum number of categories to return (enforced via checkLimit).
//
// Returns:
//   - ListCategoryResult: a struct containing the following:
//   - Categories: the list of retrieved categories.
//   - NextCursor: the timestamp of the next item for pagination, if more exist.
//   - HasMore: a boolean indicating if more results are available.
//   - Error: any error that occurred during the operation.
func (r *CategoryRepo) ListCategories(
	ctx context.Context,
	createdAfter time.Time,
	limit int,
) ListCategoryResult {
	limit = checkLimit(limit, r.minLimit, r.maxLimit)
	fetchLimit := limit + 1
	args := map[string]any{
		"created_at": createdAfter,
		"limit":      fetchLimit,
	}

	const query = `
		SELECT id, name, description, created_at
		FROM categories
		WHERE created_at >= :created_at
		ORDER BY created_at ASC, id ASC
		LIMIT :limit
	`

	stmt, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return ListCategoryResult{
			Error: fmt.Errorf("listCategories: select query failed: %w", err),
		}
	}
	defer stmt.Close()

	var categories []*Category
	for stmt.Next() {
		var category Category
		if err := stmt.StructScan(&category); err != nil {
			return ListCategoryResult{
				Error: fmt.Errorf("listCategories: scan failed: %w", err),
			}
		}
		categories = append(categories, &category)
	}

	if len(categories) == 0 {
		return ListCategoryResult{
			Categories: []*Category{},
			NextCursor: time.Time{},
			HasMore:    false,
		}
	}

	hasMore := false
	var nextCursor time.Time
	if len(categories) == fetchLimit {
		hasMore = true
		nextCursor = categories[limit].CreatedAt
		categories = categories[:limit]
	}

	return ListCategoryResult{
		Categories: categories,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
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
