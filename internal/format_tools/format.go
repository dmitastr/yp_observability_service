package formattools

import (
	"regexp"
	"strconv"
	_ "strings"
)

func FormatFloatTrimZero(val float64) string {
	re := regexp.MustCompile(`(.?0+)$`)
	str := strconv.FormatFloat(val, 'f', 3, 64)
	str = re.ReplaceAllString(str, "")
	// str = strings.TrimRight(str, "0")
	return str
}