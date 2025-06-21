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

var (
	categoryID      = uuid.MustParse("f2aa335f-6f91-4d4d-8057-53b0009bc376")
	testCategoryOne = datalayer.Category{
		ID:          uuid.MustParse("f2aa335f-6f91-4d4d-8057-53b0009bc376"),
		Name:        "Test Category A",
		Description: "Test category a description",
		CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	testCategoryTwo = datalayer.Category{
		ID:          uuid.MustParse("b12f2176-28ca-4acf-85b9-cc97ca1b3cf6"),
		Name:        "Test Category B",
		Description: "Test category B description",
		CreatedAt:   time.Date(2025, 10, 13, 0, 0, 0, 0, time.UTC),
	}
)

func TestGetCategory(t *testing.T) {
	const ctxTimeOut = 5 * time.Second
	const op = "CategoryHandler.GetCategory"

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

func TestListCategories(t *testing.T) {
	const ctxTimeOut = 5 * time.Second
	const testLimit = 10
	const op = "CategoryHandler.ListCategories"
	createdAfter := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("should respond with list of categories", func(t *testing.T) {
		listCategoriesResult := datalayer.ListCategoryResult{
			Categories: []*datalayer.Category{&testCategoryOne, &testCategoryTwo},
		}
		mockRepo := new(mocks.MockCategoryRepo)
		mockRepo.On("ListCategories", mock.Anything, createdAfter, testLimit).
			Return(listCategoriesResult)

		mockLogger := new(mocks.MockLogger)

		reqURL := "/categories?cursor=MjAyMy0wMS0wMVQwMDowMDowMFo&limit=10"
		req := httptest.NewRequest(http.MethodGet, reqURL, strings.NewReader(""))
		rw := httptest.NewRecorder()

		h := NewCategoryHandler(mockRepo, mockLogger, ctxTimeOut)
		router := mux.NewRouter()
		router.HandleFunc("/categories", h.ListCategories).Methods(http.MethodGet)
		router.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusOK, rw.Code)
		expectedResponse := `{
			"data": [
				{
					"id": "f2aa335f-6f91-4d4d-8057-53b0009bc376",
					"name": "Test Category A",
					"description": "Test category a description",
					"createdAt": "2023-01-01T00:00:00Z"
				},
				{
					"id": "b12f2176-28ca-4acf-85b9-cc97ca1b3cf6",
					"name": "Test Category B",
					"description": "Test category B description",
					"createdAt": "2025-10-13T00:00:00Z"
				}
			],
			"message": "Categories fetched successfully",
			"status": "success",
			"pagination": {
				"next_cursor": "MDAwMS0wMS0wMVQwMDowMDowMFo"
			}
		}`
		assert.JSONEq(t, expectedResponse, rw.Body.String())

		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should use default values if no request params", func(t *testing.T) {
		listCategoriesResult := datalayer.ListCategoryResult{
			Categories: []*datalayer.Category{&testCategoryOne, &testCategoryTwo},
		}
		mockRepo := new(mocks.MockCategoryRepo)
		mockRepo.On("ListCategories", mock.Anything, time.Time{}, 0).Return(listCategoriesResult)

		mockLogger := new(mocks.MockLogger)

		reqURL := "/categories"
		req := httptest.NewRequest(http.MethodGet, reqURL, strings.NewReader(""))
		rw := httptest.NewRecorder()

		h := NewCategoryHandler(mockRepo, mockLogger, ctxTimeOut)
		router := mux.NewRouter()
		router.HandleFunc("/categories", h.ListCategories).Methods(http.MethodGet)
		router.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusOK, rw.Code)
		expectedResponse := `{
			"data": [
				{
					"id": "f2aa335f-6f91-4d4d-8057-53b0009bc376",
					"name": "Test Category A",
					"description": "Test category a description",
					"createdAt": "2023-01-01T00:00:00Z"
				},
				{
					"id": "b12f2176-28ca-4acf-85b9-cc97ca1b3cf6",
					"name": "Test Category B",
					"description": "Test category B description",
					"createdAt": "2025-10-13T00:00:00Z"
				}
			],
			"message": "Categories fetched successfully",
			"status": "success",
			"pagination": {
				"next_cursor": "MDAwMS0wMS0wMVQwMDowMDowMFo"
			}
		}`
		assert.JSONEq(t, expectedResponse, rw.Body.String())

		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should respond with bad request if limit is invalid", func(t *testing.T) {
		mockRepo := new(mocks.MockCategoryRepo)
		mockLogger := new(mocks.MockLogger)
		const errMsg = "parse limit error"
		mockLogger.On("LogError", op, mock.Anything, errMsg).Return()

		reqURL := "/categories?cursor=MjAyMy0wMS0wMVQwMDowMDowMFo&limit=ab"
		req := httptest.NewRequest(http.MethodGet, reqURL, strings.NewReader(""))
		rw := httptest.NewRecorder()

		h := NewCategoryHandler(mockRepo, mockLogger, ctxTimeOut)
		router := mux.NewRouter()
		router.HandleFunc("/categories", h.ListCategories).Methods(http.MethodGet)
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

	t.Run("should respond with bad request if cursor is invalid", func(t *testing.T) {
		mockRepo := new(mocks.MockCategoryRepo)
		mockLogger := new(mocks.MockLogger)
		const errMsg = "parse cursor error"
		mockLogger.On("LogError", op, mock.Anything, errMsg).Return()

		reqURL := "/categories?cursor=MjAyMy0wMS0wMVQwMDowMDoweff&limit=10"
		req := httptest.NewRequest(http.MethodGet, reqURL, strings.NewReader(""))
		rw := httptest.NewRecorder()

		h := NewCategoryHandler(mockRepo, mockLogger, ctxTimeOut)
		router := mux.NewRouter()
		router.HandleFunc("/categories", h.ListCategories).Methods(http.MethodGet)
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

	t.Run("should respond with bad request if cursor token is invalid", func(t *testing.T) {
		mockRepo := new(mocks.MockCategoryRepo)
		mockLogger := new(mocks.MockLogger)
		const errMsg = "parse cursor error"
		mockLogger.On("LogError", op, mock.Anything, errMsg).Return()

		reqURL := "/categories?cursor=MjAyMy0wMS0wMVQ#MDow_Doweff&limit=10"
		req := httptest.NewRequest(http.MethodGet, reqURL, strings.NewReader(""))
		rw := httptest.NewRecorder()

		h := NewCategoryHandler(mockRepo, mockLogger, ctxTimeOut)
		router := mux.NewRouter()
		router.HandleFunc("/categories", h.ListCategories).Methods(http.MethodGet)
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

	t.Run("should respond with bad request if repo returns error", func(t *testing.T) {
		mockRepo := new(mocks.MockCategoryRepo)
		listCategoriesResult := datalayer.ListCategoryResult{
			Error: errors.New("db query error"),
		}
		mockRepo.On("ListCategories", mock.Anything, createdAfter, testLimit).
			Return(listCategoriesResult)

		mockLogger := new(mocks.MockLogger)
		const errMsg = "error fetching list of categories: createdAfter=`2023-01-01T00:00:00Z`, limit=`10`"
		mockLogger.On("LogError", op, mock.Anything, errMsg).Return()

		// reqURL := "/categories?cursor=MjAyMy0wMS0wMVQwMDowMDoweff&limit=10"
		reqURL := "/categories?cursor=MjAyMy0wMS0wMVQwMDowMDowMFo&limit=10"
		req := httptest.NewRequest(http.MethodGet, reqURL, strings.NewReader(""))
		rw := httptest.NewRecorder()

		h := NewCategoryHandler(mockRepo, mockLogger, ctxTimeOut)
		router := mux.NewRouter()
		router.HandleFunc("/categories", h.ListCategories).Methods(http.MethodGet)
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
