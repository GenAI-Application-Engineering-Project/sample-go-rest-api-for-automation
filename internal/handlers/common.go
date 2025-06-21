package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	applogger "product-service/internal/app_logger"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	// Error codes
	ErrCodeInternalServerError = 1600
	ErrCodeInvalidFieldFormat  = 1002
	ErrCodeResourceNotFound    = 1300

	LimitParam = "limit"
	CursorParm = "cursor"

	// Error code messages
	ErrMessageInvalidFieldFormat  = "Invalid field format"
	ErrMessageResourceNotFound    = "Resource not found"
	ErrMessageInternalServerError = "Internal server error"

	StatusSuccess = "success"
	StatusError   = "error"
)

type Pagination struct {
	Page       int       `json:"page,omitempty"`
	PerPage    int       `json:"per_page,omitempty"`
	Total      int       `json:"total,omitempty"`
	TotalPages int       `json:"total_pages,omitempty"`
	HasMore    bool      `json:"has_more,omitempty"`
	NextCursor time.Time `json:"next_cursor,omitempty"`
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
	op string,
	logger applogger.LoggerInterface,
) {
	// Encode response to buffer
	var buf bytes.Buffer
	if details != nil {
		err := json.NewEncoder(&buf).Encode(details)
		if err != nil {
			logger.LogError(op, err, "error encoding json response")
			WriteErrorResponse(
				w,
				http.StatusInternalServerError,
				ErrCodeInternalServerError,
				ErrMessageInternalServerError,
				nil,
				op,
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
			logger.LogError(op, err, "error writing response to client")
		}
	}
}

func WriteErrorResponse(
	w http.ResponseWriter,
	statusCode int,
	code int,
	message string,
	details any,
	op string,
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

	WriteResponse(w, statusCode, resp, op, logger)
}

func WriteSuccessResponse(
	w http.ResponseWriter,
	statusCode int,
	message string,
	data any,
	pagination *Pagination,
	meta any,
	op string,
	logger applogger.LoggerInterface,
) {
	resp := HTTPSuccessResponse{
		Status:     StatusSuccess,
		Data:       data,
		Pagination: pagination,
		Meta:       meta,
		Message:    message,
	}

	WriteResponse(w, statusCode, resp, op, logger)
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

// DecodeCursorToTime decodes a base64 URL-safe string back into a time.Time
func DecodeCursorToTime(cursor string) (time.Time, error) {
	decodedBytes, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cursor encoding: %s", cursor)
	}

	t, err := time.Parse(time.RFC3339Nano, string(decodedBytes))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cursor time format: %s", cursor)
	}
	return t, nil
}

// EncodeTimeToCursor encodes a time.Time into a base64 URL-safe string
func EncodeTimeToCursor(t time.Time) string {
	timeStr := t.UTC().Format(time.RFC3339Nano)
	return base64.RawURLEncoding.EncodeToString([]byte(timeStr))
}

func ParseCursor(r *http.Request) (time.Time, error) {
	cursorStr := r.URL.Query().Get(CursorParm)
	if cursorStr == "" {
		return time.Time{}, nil
	}

	createdAfter, err := DecodeCursorToTime(cursorStr)
	if err != nil {
		return time.Time{}, err
	}
	return createdAfter, nil
}

func ParseLimit(r *http.Request) (int, error) {
	limitStr := r.URL.Query().Get(LimitParam)
	if limitStr == "" {
		return 0, nil
	}

	val, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		return 0, err
	}

	return int(val), nil
}

func ParseAndValidatePagination(
	r *http.Request,
	op string,
	logger applogger.LoggerInterface,
) (time.Time, int, bool) {
	cursor, err := ParseCursor(r)
	if err != nil {
		logger.LogError(op, err, "parse cursor error")
		return time.Time{}, 0, false
	}
	limit, err := ParseLimit(r)
	if err != nil {
		logger.LogError(op, err, "parse limit error")
		return time.Time{}, 0, false
	}
	return cursor, limit, true
}
