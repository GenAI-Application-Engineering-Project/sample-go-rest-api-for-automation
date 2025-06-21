package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	applogger "product-service/internal/app_logger"
	datalayer "product-service/internal/data_layer"
)

type CategoryHandler struct {
	repo       datalayer.CategoryRepoInterface
	appLogger  applogger.LoggerInterface
	ctxTimeout time.Duration
}

func NewCategoryHandler(
	repo datalayer.CategoryRepoInterface,
	appLogger applogger.LoggerInterface,
	ctxTimeout time.Duration,
) *CategoryHandler {
	return &CategoryHandler{
		repo:       repo,
		appLogger:  appLogger,
		ctxTimeout: ctxTimeout,
	}
}

// GetCategory  handles HTTP GET request to fetch a category.
//
// @Summary     Get a category by ID
// @Description Retrieves a category by its ID from the database
// @Tags        Categories
// @Accept      json
// @Produce     json
// @Param       id path string true "Category UUID"
// @Success     200 {object} CategoryResponse
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /categories/{id} [get]
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	const op = "CategoryHandler.GetCategory"

	// Parse id param
	id, err := ParseUUIDParam(r, "id")
	if err != nil {
		h.appLogger.LogError(op, err, "error parsing `id` from uuid param")
		WriteErrorResponse(
			w,
			http.StatusBadRequest,
			ErrCodeInvalidFieldFormat,
			ErrMessageInvalidFieldFormat,
			nil,
			op,
			h.appLogger,
		)
		return
	}

	// Fetch category from repo
	ctx, cancel := context.WithTimeout(r.Context(), h.ctxTimeout)
	defer cancel()

	category, err := h.repo.GetCategoryByID(ctx, id)
	if err != nil {
		msg := fmt.Sprintf("failed to fetch category from repo: id=`%s`", id)
		h.appLogger.LogError(op, err, msg)

		if errors.Is(err, datalayer.ErrNotFound) {
			WriteErrorResponse(
				w,
				http.StatusBadRequest,
				ErrCodeResourceNotFound,
				ErrMessageResourceNotFound,
				nil,
				op,
				h.appLogger,
			)
		} else {
			WriteErrorResponse(w, http.StatusInternalServerError, ErrCodeInternalServerError, ErrMessageInternalServerError, nil, op, h.appLogger)
		}
		return
	}

	// Write http response
	WriteSuccessResponse(
		w,
		http.StatusOK,
		"Category fetched successfully",
		category,
		nil,
		nil,
		op,
		h.appLogger,
	)
}

// ListCategories handles HTTP GET requests to fetch a paginated list of categories.
//
// @Summary      List categories
// @Description  Retrieves a paginated list of category resources from the database.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        cursor  query     string false "Pagination cursor (RFC3339 timestamp)"
// @Param        limit   query     int    false "Max number of categories to return (e.g. 50)"
// @Success      200     {object}  ListCategoriesResponse
// @Failure      400     {object}  ErrorResponse "Invalid cursor or limit"
// @Failure      500     {object}  ErrorResponse "Internal server error"
// @Router       /categories [get]
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	const op = "CategoryHandler.ListCategories"

	createdAfter, limit, isValid := ParseAndValidatePagination(r, op, h.appLogger)
	if !isValid {
		WriteErrorResponse(
			w,
			http.StatusBadRequest,
			ErrCodeInvalidFieldFormat,
			ErrMessageInvalidFieldFormat,
			nil,
			op,
			h.appLogger,
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.ctxTimeout)
	defer cancel()

	result := h.repo.ListCategories(ctx, createdAfter, limit)
	if result.Error != nil {
		errMsg := fmt.Sprintf(
			"error fetching list of categories: createdAfter=`%s`, limit=`%d`",
			createdAfter.Format(time.RFC3339),
			limit,
		)
		h.appLogger.LogError(op, result.Error, errMsg)
		WriteErrorResponse(
			w,
			http.StatusInternalServerError,
			ErrCodeInternalServerError,
			ErrMessageInternalServerError,
			nil,
			op,
			h.appLogger,
		)
		return
	}

	pagination := Pagination{
		HasMore:    result.HasMore,
		NextCursor: EncodeTimeToCursor(result.NextCursor),
	}

	WriteSuccessResponse(
		w,
		http.StatusOK,
		"Categories fetched successfully",
		result.Categories,
		&pagination,
		nil,
		op,
		h.appLogger,
	)
}
