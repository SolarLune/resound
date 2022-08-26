package resound

import (
	"math"
)

func addChannelValue(value, add int16) int16 {

	if add > 0 {
		if value > math.MaxInt16-add {
			return math.MaxInt16
		}
	} else {
		if value < math.MinInt16-add {
			return math.MinInt16
		}
	}

	return value + add

}
