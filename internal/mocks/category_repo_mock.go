package mocks

import (
	"context"
	"time"

	datalayer "product-service/internal/data_layer"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockCategoryRepo struct {
	mock.Mock
}

func (m *MockCategoryRepo) GetCategoryByID(
	ctx context.Context,
	id uuid.UUID,
) (*datalayer.Category, error) {
	args := m.Called(ctx, id)
	if cat, ok := args.Get(0).(*datalayer.Category); ok {
		return cat, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCategoryRepo) ListCategories(
	ctx context.Context,
	createdAfter time.Time,
	limit int,
) datalayer.ListCategoryResult {
	args := m.Called(ctx, createdAfter, limit)
	return args.Get(0).(datalayer.ListCategoryResult)
}

func (m *MockCategoryRepo) CreateCategory(ctx context.Context, category *datalayer.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepo) UpdateCategory(ctx context.Context, category *datalayer.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepo) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
