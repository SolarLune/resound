package resound

import (
	"io"
	"math"
	"strconv"
)

// IEffect indicates an effect that implements io.ReadSeeker and generally takes effect on an existing audio stream.
// It represents the result of applying an effect to an audio stream, and is playable in its own right.
type IEffect interface {
	io.ReadSeeker
	ApplyEffect(data []byte, bytesRead int) // This function is called when sound data goes through an effect. The effect should modify the data byte buffer.
	SetSource(io.ReadSeeker)                // This function allows an effect's source to be dynamically altered; this allows for easy chaining with resound.ChainEffects().
}

// AudioBuffer wraps a []byte of audio data and provides handy functions to get
// and set values for a specific position in the buffer.
type AudioBuffer []byte

func (ab AudioBuffer) Len() int {
	return len(ab) / 4
}

// Get returns the values for the left and right audio channels at the specified stream sample index.
// The values returned for the left and right audio channels range from 0 to 1.
func (ab AudioBuffer) Get(i int) (l, r float64) {
	lc := float64(int16(ab[i*4]) | int16(ab[i*4+1])<<8)
	rc := float64(int16(ab[i*4+2]) | int16(ab[i*4+3])<<8)
	lc /= math.MaxInt16
	rc /= math.MaxInt16
	return lc, rc
}

// Set sets the left and right audio channel values at the specified stream sample index.
// The values should range from 0 to 1.
func (ab AudioBuffer) Set(i int, l, r float64) {

	max := float64(math.MaxInt16)

	l = clamp(l*math.MaxInt16, -max, max)
	r = clamp(r*math.MaxInt16, -max, max)

	lcc := int16(l)
	rcc := int16(r)

	ab[(i * 4)] = byte(lcc)
	ab[(i*4)+1] = byte(lcc >> 8)
	ab[(i*4)+2] = byte(rcc)
	ab[(i*4)+3] = byte(rcc >> 8)
}

func (ab AudioBuffer) String() string {
	s := "{ "
	for i := 0; i < ab.Len(); i++ {
		l, r := ab.Get(i)
		ls := strconv.FormatFloat(l, 'f', 6, 64)
		rs := strconv.FormatFloat(r, 'f', 6, 64)
		s += "( " + ls + ", " + rs + " ) "
	}
	s += " }"
	return s
}

// ChainEffects chains multiple effects for you automatically, returning the last chained effect.
// Example:
// sfxChain := resound.Chain(sourceSound,
//
//	resound.NewDelay(nil).SetWait(0.2).SetStrength(0.5),
//	resound.NewPan(nil),
//	resound.NewVolume(nil),
//
// )
// sfxChain at the end would be the Volume effect, which is being fed by the Pan effect, which is fed by the Delay effect.
func ChainEffects(source io.ReadSeeker, effects ...IEffect) IEffect {
	effects[0].SetSource(source)
	for i := 1; i < len(effects); i++ {
		effects[i].SetSource(effects[i-1])
	}
	return effects[len(effects)-1]
}
