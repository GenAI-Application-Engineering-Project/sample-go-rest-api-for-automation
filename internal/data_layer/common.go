package datalayer

import (
	"database/sql"
	"errors"
	"fmt"
)

const (
	maxLimit = 1000
	minLimit = 1
	errMsg   = "%s failed: %w (db error: %v)"
)

var (
	errNotFound  = errors.New("record not found")
	errDBFailure = errors.New("database failure")
)

func checkLimit(limit int) int {
	if limit < minLimit {
		limit = minLimit
	} else if limit > maxLimit {
		limit = maxLimit
	}
	return limit
}

func checkRowsAffected(result sql.Result, funcName string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(errMsg, funcName, errDBFailure, err)
	} else if rows == 0 {
		return fmt.Errorf(errMsg, funcName, errNotFound, nil)
	}

	return nil
}
