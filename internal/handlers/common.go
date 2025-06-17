package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	applogger "product-service/internal/app_logger"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	// Error codes
	ErrCodeInternalServerError = 1600
	ErrCodeInvalidFieldFormat  = 1002
	ErrCodeResourceNotFound    = 1300

	// Error code messages
	ErrMessageInvalidFieldFormat  = "Invalid field format"
	ErrMessageResourceNotFound    = "Resource not found"
	ErrMessageInternalServerError = "Internal server error"

	StatusSuccess = "success"
	StatusError   = "error"
)

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type HTTPSuccessResponse struct {
	Status     string      `json:"status"`
	Data       any         `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Meta       any         `json:"meta,omitempty"`
	Message    string      `json:"message"`
}

type HTTPErrorResponse struct {
	Status string `json:"status"`
	Error  Error  `json:"error"`
}

func WriteResponse(
	w http.ResponseWriter,
	statusCode int,
	details any,
	logger applogger.LoggerInterface,
) {
	// Encode response to buffer
	var buf bytes.Buffer
	if details != nil {
		err := json.NewEncoder(&buf).Encode(details)
		if err != nil {
			logger.LogError(err, "error encoding json response")
			WriteErrorResponse(
				w,
				http.StatusInternalServerError,
				ErrCodeInternalServerError,
				ErrMessageInternalServerError,
				nil,
				logger,
			)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Write response body
	if buf.Len() > 0 {
		if _, err := buf.WriteTo(w); err != nil {
			logger.LogError(err, "error writing response to client")
		}
	}
}

func WriteErrorResponse(
	w http.ResponseWriter,
	statusCode int,
	code int,
	message string,
	details any,
	logger applogger.LoggerInterface,
) {
	resp := HTTPErrorResponse{
		Status: StatusError,
		Error: Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	WriteResponse(w, statusCode, resp, logger)
}

func WriteSuccessResponse(
	w http.ResponseWriter,
	statusCode int,
	message string,
	data any,
	pagination *Pagination,
	meta any,
	logger applogger.LoggerInterface,
) {
	resp := HTTPSuccessResponse{
		Status:     StatusSuccess,
		Data:       data,
		Pagination: pagination,
		Meta:       meta,
		Message:    message,
	}

	WriteResponse(w, statusCode, resp, logger)
}

func ParseUUIDParam(r *http.Request, param string) (uuid.UUID, error) {
	vars := mux.Vars(r)
	uuidStrVal, exists := vars[param]
	if !exists || uuidStrVal == "" {
		return uuid.Nil, fmt.Errorf("param not found: `%s`", param)
	}

	uuidVal, err := uuid.Parse(uuidStrVal)
	if err != nil {
		return uuid.Nil, err
	}

	if uuidVal == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid id param: `%s`", uuidVal)
	}

	return uuidVal, nil
}
