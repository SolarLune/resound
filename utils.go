package resound

func clamp(v, min, max float64) float64 {
	if v > max {
		return max
	} else if v < min {
		return min
	}
	return v
}
