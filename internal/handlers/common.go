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
	StatusBadRequestMessage          = "Bad Request"
	StatusInternalServerErrorMessage = "Internal Server Error"
)

func parseUUIDParam(r *http.Request, param string) (uuid.UUID, error) {
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

func writeHTTPResponse(
	w http.ResponseWriter, responseBody any, appLogger applogger.LoggerInterface,
) {
	// Encode response to buffer
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(responseBody)
	if err != nil {
		appLogger.LogError(err, "error encoding json response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set headers and write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write response body
	_, err = buf.WriteTo(w)
	if err != nil {
		msg := "error writing json response: " + buf.String()
		appLogger.LogError(err, msg)
	}
}
