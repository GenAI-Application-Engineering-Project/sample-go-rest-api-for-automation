package handlers

import (
	"net/http"
	"net/http/httptest"
	"product-service/internal/mocks"
	"strings"
	"testing"
	"time"

	datalayer "product-service/internal/data_layer"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCategory(t *testing.T) {
	const ctxTimeOut = 5 * time.Second

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
			"ID": "f2aa335f-6f91-4d4d-8057-53b0009bc376",
			"Name": "Test Category A",
			"Description": "Test category a description",
			"CreatedAt": "2023-01-01T00:00:00Z"
		}`
		assert.JSONEq(t, expectedResponse, rw.Body.String())

		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
