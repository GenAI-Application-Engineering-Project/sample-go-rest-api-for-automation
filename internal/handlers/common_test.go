package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	datalayer "product-service/internal/data_layer"
	"product-service/internal/mocks"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWriteResponse(t *testing.T) {
	t.Run("should write success response", func(t *testing.T) {
		data := datalayer.Category{
			ID:          uuid.MustParse("f2aa335f-6f91-4d4d-8057-53b0009bc376"),
			Name:        "Test Category A",
			Description: "Test category a description",
			CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		rw := httptest.NewRecorder()
		WriteSuccessResponse(rw, 200, "success", data, nil, nil, nil)

		expectedResponse := `{
			"data": {
				"id": "f2aa335f-6f91-4d4d-8057-53b0009bc376",
				"name": "Test Category A",
				"description": "Test category a description",
				"createdAt": "2023-01-01T00:00:00Z"
			},
			"message": "success",
			"status": "success"
		}`
		assert.JSONEq(t, expectedResponse, rw.Body.String())
	})

	t.Run("should respond with internal server error if encoding fails", func(t *testing.T) {
		type Node struct {
			Value string
			Next  *Node
		}
		data := &Node{Value: "A"}
		data.Next = data

		mockLogger := new(mocks.MockLogger)
		const errMsg = "error encoding json response"
		mockLogger.On("LogError", mock.Anything, errMsg).Return()

		rw := httptest.NewRecorder()
		WriteSuccessResponse(rw, 200, "success", data, nil, nil, mockLogger)

		expectedResponse := `{
			"status":"error",
			"error": {
				"code": 1600,
				"message": "Internal server error"
			}
		}`
		assert.JSONEq(t, expectedResponse, rw.Body.String())

		mockLogger.AssertExpectations(t)
	})

	t.Run("should respond with internal server error if encoding fails", func(t *testing.T) {
		data := map[string]string{"message": "hello"}
		err := errors.New("writer error")

		mockLogger := new(mocks.MockLogger)
		const errMsg = "error writing response to client"
		mockLogger.On("LogError", err, errMsg).Return().Once()

		mockResponseWriter := new(mocks.MockHTTPResponseWriter)
		mockResponseWriter.On("Write", mock.Anything).Return(0, err)
		mockResponseWriter.On("Header").Return(http.Header{})
		mockResponseWriter.On("WriteHeader", 200).Return()

		WriteSuccessResponse(mockResponseWriter, 200, "success", data, nil, nil, mockLogger)

		mockResponseWriter.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestReadUUIdParam(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	const param = "id"

	t.Run("should parse valid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = mux.SetURLVars(req, map[string]string{param: validUUID})

		actualUUID, err := ParseUUIDParam(req, param)
		assert.Equal(t, validUUID, actualUUID.String())
		assert.NoError(t, err)
	})

	t.Run("should return error if uuid param is not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		actualUUID, err := ParseUUIDParam(req, param)
		assert.Equal(t, uuid.Nil, actualUUID)
		assert.Error(t, err)
		assert.Equal(t, "param not found: `id`", err.Error())
	})

	t.Run("should return error if uuid param is nil", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = mux.SetURLVars(req, map[string]string{param: uuid.Nil.String()})

		actualUUID, err := ParseUUIDParam(req, param)
		assert.Equal(t, uuid.Nil, actualUUID)
		assert.Error(t, err)
		assert.Equal(t, "invalid id param: `00000000-0000-0000-0000-000000000000`", err.Error())
	})
}
