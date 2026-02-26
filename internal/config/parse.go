package config

import "strconv"

func parseFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func parseInt64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}
