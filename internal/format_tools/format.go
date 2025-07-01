package formattools

import (
	"strconv"
	"strings"
)

func FormatFloatTrimZero(val float64) string {
	str := strconv.FormatFloat(val, 'f', 3, 64)
	str = strings.TrimRight(str, "0")
	return str
}