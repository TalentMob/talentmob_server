package util

import (
	"strconv"
	"fmt"
)

func  ConvertStringToBool(s string) (bool, error) {
	b, err := strconv.ParseBool(s)

	if err != nil {
		return false, err
	}

	return b, nil
}
func  ConvertStringToInt(s string) (int, error) {
	c, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		return 0, err
	}

	return int(c), nil
}

func  ConvertToString(o interface{}) string {
	return fmt.Sprint(o)
}

func  UintToString(i uint) string {
	str := fmt.Sprint(i)
	return str
}