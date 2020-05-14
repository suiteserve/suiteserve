package handlers

import (
	"strconv"
)

func parseString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseUint(s string) (*uint, error) {
	if s == "" {
		return nil, nil
	}
	i, err := strconv.ParseUint(s, 10, 32)
	ui := uint(i)
	return &ui, err
}

func parseInt64(s string) (*int64, error) {
	if s == "" {
		return nil, nil
	}
	i, err := strconv.ParseInt(s, 10, 64)
	return &i, err
}

func parseBool(s string) (*bool, error) {
	if s == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(s)
	return &b, err
}
