package datalayer

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("resource not found")

func checkLimit(limit int, minLimit, maxLimit int) int {
	if limit < minLimit {
		limit = minLimit
	} else if limit > maxLimit {
		limit = maxLimit
	}
	return limit
}

func checkRowsAffected(result sql.Result, op string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: no rows affected: %w", op, ErrNotFound)
	}
	return nil
}
