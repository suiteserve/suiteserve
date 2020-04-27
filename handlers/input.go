package handlers

import (
	"strconv"
	"time"
)

func parseTime(s string) (time.Time, error) {
	var t time.Time
	if s != "" {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		t = time.Unix(i, 0)
	}
	return t, nil
}
