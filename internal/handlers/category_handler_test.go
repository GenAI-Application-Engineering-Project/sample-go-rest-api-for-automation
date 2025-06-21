package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"product-service/internal/mocks"

	datalayer "product-service/internal/data_layer"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCategory(t *testing.T) {
	const ctxTimeOut = 5 * time.Second
	const op = "CategoryHandler.GetCategory"

	categoryID := uuid.MustParse("f2aa335f-6f91-4d4d-8057-53b0009bc376")
	testCategoryOne := datalayer.Category{
		ID:          uuid.MustParse("f2aa335f-6f91-4d4d-8057-53b0009bc376"),
		Name:        "Test Category A",
		Description: "Test category a description",
		CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	t.Run("should respond with category", func(t *testing.T) {
		mockRepo := new(mocks.MockCategoryRepo)
		mockRepo.On("GetCategoryByID", mock.Anything, categoryID).Return(&testCategoryOne, nil)

		mockLogger := new(mocks.MockLogger)

		reqURL := "/categories/" + categoryID.String()
		req := httptest.NewRequest(http.MethodGet, reqURL, strings.NewReader(""))
		rw := httptest.NewRecorder()

		h := NewCategoryHandler(mockRepo, mockLogger, ctxTimeOut)
		router := mux.NewRouter()
		router.HandleFunc("/categories/{id}", h.GetCategory).Methods(http.MethodGet)
		router.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusOK, rw.Code)
		expectedResponse := `{
			"data": {
				"id": "f2aa335f-6f91-4d4d-8057-53b0009bc376",
				"name": "Test Category A",
				"description": "Test category a description",
				"createdAt": "2023-01-01T00:00:00Z"
			},
			"message": "Category fetched successfully",
			"status": "success"
		}`
		assert.JSONEq(t, expectedResponse, rw.Body.String())

		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should respond with bad request if id param is not valid", func(t *testing.T) {
		mockRepo := new(mocks.MockCategoryRepo)
		mockLogger := new(mocks.MockLogger)
		const errMsg = "error parsing `id` from uuid param"
		mockLogger.On("LogError", op, mock.Anything, errMsg).Return()

		reqURL := "/categories/1234" //  + categoryID.String()
		req := httptest.NewRequest(http.MethodGet, reqURL, strings.NewReader(""))
		rw := httptest.NewRecorder()

		h := NewCategoryHandler(mockRepo, mockLogger, ctxTimeOut)
		router := mux.NewRouter()
		router.HandleFunc("/categories/{id}", h.GetCategory).Methods(http.MethodGet)
		router.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusBadRequest, rw.Code)
		expectedResponse := `{
			"status":"error",
			"error": {
				"code": 1002,
				"message": "Invalid field format"
			}
		}`
		assert.JSONEq(t, expectedResponse, rw.Body.String())

		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should respond with bad request if category is not found", func(t *testing.T) {
		mockRepo := new(mocks.MockCategoryRepo)
		mockRepo.On("GetCategoryByID", mock.Anything, categoryID).Return(nil, datalayer.ErrNotFound)

		mockLogger := new(mocks.MockLogger)
		errMsg := "failed to fetch category from repo: id=`f2aa335f-6f91-4d4d-8057-53b0009bc376`"
		mockLogger.On("LogError", op, datalayer.ErrNotFound, errMsg)

		reqURL := "/categories/" + categoryID.String()
		req := httptest.NewRequest(http.MethodGet, reqURL, strings.NewReader(""))
		rw := httptest.NewRecorder()

		h := NewCategoryHandler(mockRepo, mockLogger, ctxTimeOut)
		router := mux.NewRouter()
		router.HandleFunc("/categories/{id}", h.GetCategory).Methods(http.MethodGet)
		router.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusBadRequest, rw.Code)
		expectedResponse := `{
			"status":"error",
			"error": {
				"code": 1300,
				"message": "Resource not found"
			}
		}`
		assert.JSONEq(t, expectedResponse, rw.Body.String())

		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should respond with internal server error if repo returns error", func(t *testing.T) {
		err := errors.New("db query error")
		mockRepo := new(mocks.MockCategoryRepo)
		mockRepo.On("GetCategoryByID", mock.Anything, categoryID).Return(nil, err)

		mockLogger := new(mocks.MockLogger)
		errMsg := "failed to fetch category from repo: id=`f2aa335f-6f91-4d4d-8057-53b0009bc376`"
		mockLogger.On("LogError", op, err, errMsg)

		reqURL := "/categories/" + categoryID.String()
		req := httptest.NewRequest(http.MethodGet, reqURL, strings.NewReader(""))
		rw := httptest.NewRecorder()

		h := NewCategoryHandler(mockRepo, mockLogger, ctxTimeOut)
		router := mux.NewRouter()
		router.HandleFunc("/categories/{id}", h.GetCategory).Methods(http.MethodGet)
		router.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		expectedResponse := `{
			"status":"error",
			"error": {
				"code": 1600,
				"message": "Internal server error"
			}
		}`
		assert.JSONEq(t, expectedResponse, rw.Body.String())

		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
