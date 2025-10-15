package formattools

import (
	"strconv"
	_ "strings"
)

func FormatFloatTrimZero(val float64) string {
	str := strconv.FormatFloat(val, 'f', -1, 64)
	return str
}
