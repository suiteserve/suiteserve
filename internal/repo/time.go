package repo

import (
	"strconv"
)

type Time int64

func ParseTime(s string) (Time, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	return Time(i), err
}
