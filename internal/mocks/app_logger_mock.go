package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (l *MockLogger) LogError(err error, msg string) {
	l.Called(err, msg)
}
