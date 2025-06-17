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

func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	// Parse id param
	id, err := ParseUUIDParam(r, "id")
	if err != nil {
		h.appLogger.LogError(err, "getCategory: error parsing `id` from uuid param")
		WriteErrorResponse(
			w,
			http.StatusBadRequest,
			ErrCodeInvalidFieldFormat,
			ErrMessageInvalidFieldFormat,
			nil,
			h.appLogger,
		)
		return
	}

	// Fetch category from repo
	ctx, cancel := context.WithTimeout(r.Context(), h.ctxTimeout)
	defer cancel()

	category, err := h.repo.GetCategoryByID(ctx, id)
	if err != nil {
		msg := fmt.Sprintf("getCategory: failed to fetch category from repo: id=`%s`", id)
		h.appLogger.LogError(err, msg)

		if errors.Is(err, datalayer.ErrNotFound) {
			WriteErrorResponse(
				w,
				http.StatusBadRequest,
				ErrCodeResourceNotFound,
				ErrMessageResourceNotFound,
				nil,
				h.appLogger,
			)
		} else {
			WriteErrorResponse(w, http.StatusInternalServerError, ErrCodeInternalServerError, ErrMessageInternalServerError, nil, h.appLogger)
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
		h.appLogger,
	)
}
