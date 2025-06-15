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
	applogger  applogger.LoggerInterface
	ctxTimeout time.Duration
}

func NewCategoryHandler(
	repo datalayer.CategoryRepoInterface,
	applogger applogger.LoggerInterface,
	ctxTimeout time.Duration,
) *CategoryHandler {
	return &CategoryHandler{
		repo:       repo,
		applogger:  applogger,
		ctxTimeout: ctxTimeout,
	}
}

func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	// Parse id param
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		h.applogger.LogError(err, "getCategory: error parsing `id` from uuid param")
		http.Error(w, StatusBadRequestMessage, http.StatusBadRequest)
		return
	}

	// Fetch category from repo
	ctx, cancel := context.WithTimeout(r.Context(), h.ctxTimeout)
	defer cancel()

	category, err := h.repo.GetCategoryByID(ctx, id)
	if err != nil {
		msg := fmt.Sprintf("getCategory: failed to fetch category from repo: id=`%s`", id)
		h.applogger.LogError(err, msg)

		if errors.Is(err, datalayer.ErrNotFound) {
			http.Error(w, StatusBadRequestMessage, http.StatusBadRequest)
		} else {
			http.Error(w, StatusInternalServerErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	// Write http response
	writeHTTPResponse(w, category, h.applogger)
}
