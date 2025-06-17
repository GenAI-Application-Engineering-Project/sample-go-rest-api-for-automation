package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockHTTPResponseWriter struct {
	mock.Mock
}

func (m *MockHTTPResponseWriter) Header() http.Header {
	args := m.Called()
	return args.Get(0).(http.Header)
}

func (m *MockHTTPResponseWriter) WriteHeader(statusCode int) {
	m.Called(statusCode)
}

func (m *MockHTTPResponseWriter) Write(p []byte) (int, error) {
	args := m.Called(p)
	return 0, args.Error(1)
}
