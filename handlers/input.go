package handlers

import (
	"strconv"
)

func parseUint(s string) (uint, error) {
	i64, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(i64), nil
}