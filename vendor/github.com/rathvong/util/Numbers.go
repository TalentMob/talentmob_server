package util

import (
	"strconv"
	"bytes"
	"fmt"
	"math"
	"log"
)

func UniqueUint(input []uint) []uint {
	result := make([]uint, 0, len(input))
	newMap := make(map[uint]bool)

	for _, val := range input {
		if _, ok := newMap[val]; !ok {
			newMap[val] = true
			result = append(result, val)
		}
	}

	return result
}

func ConvertToUint(s string) (uint, error) {

	cs, err := strconv.ParseUint(s, 10, 64)

	if err != nil {
		return 0, err

	}

	return uint(cs), nil

}



func ConvertToUint64(s string) (uint64, error) {

	cs, err := strconv.ParseUint(s, 10, 64)

	if err != nil {
		return 0, err

	}

	return uint64(cs), nil

}

func FormatToArrayStringSQLQuery(array []uint) (sqlArray string) {
	buffer := bytes.NewBufferString("")

	for i, id := range array {

		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf("%v", id))
	}

	sqlArray = buffer.String()
	return
}

func FloatRound(input float64) float64 {
	if input < 0 {
		return math.Ceil(input - 0.5)
	}
	return math.Floor(input + 0.5)
}

func FloatRoundUp(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Ceil(digit)
	newVal = round / pow
	return
}

func FloatRoundDown(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Floor(digit)
	newVal = round / pow
	return
}

func ConvertStringToFloat64(s string) (float64, error) {
	i, err := strconv.ParseFloat(s, 64)

	if err != nil {
		return 0, err
	}

	return i, nil

}

func ConvertPageParamsToInt(s string) int {
	if s == "" {
		return 1
	}

	c, err := strconv.ParseInt(s, 10, 64)

	if err == nil {
		return int(c)
	}

	log.Println("convertStringToInt()")
	log.Println(err)

	defer func() { if c == 0 {c = 1} }()
	return 1

}

