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
}

// AudioBuffer wraps a []byte of audio data and provides handy functions to get
// and set values for a specific position in the buffer.
type AudioBuffer []byte

func (ab AudioBuffer) Len() int {
	// We divide by 4 because it's L16 PCM audio at 2 channels, with int16s composing 2 bytes per sample, per channel (2 bytes * 2 channels = 4).
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
