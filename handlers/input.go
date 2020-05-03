package handlers

import (
	"strconv"
)

func parseUint(s string) (uint, bool, error) {
	if s == "" {
		return 0, false, nil
	}
	i, err := strconv.ParseUint(s, 10, 32)
	return uint(i), err == nil, err
}

func parseInt64(s string) (int64, bool, error) {
	if s == "" {
		return 0, false, nil
	}
	i, err := strconv.ParseInt(s, 10, 64)
	return i, err == nil, err
}

func parseBool(s string) (bool, bool, error) {
	if s == "" {
		return false, false, nil
	}
	b, err := strconv.ParseBool(s)
	return b, err == nil, err
}
