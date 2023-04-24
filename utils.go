package resound

// ChainEffects chains multiple effects for you automatically, returning the last chained effect.
// Example:
// sfxChain := resound.Chain(
//
//	resound.NewDelay(nil).SetWait(0.2).SetStrength(0.5),
//	resound.NewPan(nil),
//	resound.NewVolume(nil),
//
// )
// sfxChain at the end would be the Volume effect, which is being fed by the Pan effect, which is fed by the Delay effect.
func ChainEffects(effects ...IEffect) IEffect {
	for i := 1; i < len(effects); i++ {
		effects[i].setSource(effects[i-1])
	}
	return effects[len(effects)-1]
}

func clamp(v, min, max float64) float64 {
	if v > max {
		return max
	} else if v < min {
		return min
	}
	return v
}
