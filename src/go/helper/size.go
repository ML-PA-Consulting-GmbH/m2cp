package helper

import (
	"fmt"
	"math"
)

func SizeWithBinaryUnit(bytesize uint64) string {
	bf := float64(bytesize)

	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi"} {
		if math.Abs(bf) < 1024.0 {
			if unit == "" {
				return fmt.Sprintf("%.0f %sB", bf, unit)
			} else {
				return fmt.Sprintf("%3.1f %sB", bf, unit)
			}

		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1f EiB", bf)
}
