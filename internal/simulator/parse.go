package simulator

import "strconv"

func parseHexInt64(v string) (int64, error) {
	if len(v) >= 2 && v[:2] == "0x" {
		return strconv.ParseInt(v[2:], 16, 64)
	}
	return strconv.ParseInt(v, 10, 64)
}
