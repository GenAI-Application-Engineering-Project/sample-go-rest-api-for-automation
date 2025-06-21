package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (l *MockLogger) LogError(op string, err error, msg string) {
	l.Called(op, err, msg)
}
