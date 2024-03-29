package effects

import (
	"io"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/solarlune/resound"
	"github.com/tanema/gween/ease"
)

// Volume is an effect that changes the overall volume of the incoming audio byte stream.
type Volume struct {
	strength      float64
	normalization float64
	active        bool
	Source        io.ReadSeeker
}

// NewVolume creates a new Volume effect. source is the source stream to apply this effect to.
// If you add this effect to a DSPChannel, source can be nil, as it will take effect for whatever
// streams are played through the DSPChannel.
func NewVolume(source io.ReadSeeker) *Volume {
	volume := &Volume{Source: source, strength: 1, active: true, normalization: 1}
	return volume
}

// Clone clones the effect, returning an resound.IEffect.
func (volume *Volume) Clone() resound.IEffect {
	return &Volume{
		strength: volume.strength,
		active:   volume.active,
		Source:   volume.Source,
	}
}

func (volume *Volume) Read(p []byte) (n int, err error) {

	if n, err = volume.Source.Read(p); err != nil {
		return
	}

	volume.ApplyEffect(p, n)

	return
}

func (volume *Volume) ApplyEffect(p []byte, bytesRead int) {

	// If the effect isn't active, then we can return early.
	if !volume.active {
		return
	}

	perc := volume.strength
	if volume.strength <= 1 {
		perc = float64(ease.InSine(float32(volume.strength), 0, 1, 1))
	}

	perc *= volume.normalization

	// Make an audio buffer for easy stream manipulation.
	audio := resound.AudioBuffer(p)

	// Loop through all frames in the stream that are available to be read.

	// We use bytesRead / 4 here because the size of the byte buffer can be larger
	// than the amount of bytes actually read, whoops
	for i := 0; i < bytesRead/4; i++ {

		// Get the audio value:
		l, r := audio.Get(i)

		// Multiply it by the volume strength:
		l *= perc
		r *= perc

		// Set it back, and you're done.
		audio.Set(i, l, r)
	}

}

func (volume *Volume) Seek(offset int64, whence int) (int64, error) {
	if volume.Source == nil {
		return 0, nil
	}
	return volume.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (volume *Volume) SetActive(active bool) *Volume {
	volume.active = active
	return volume
}

// Active returns if the effect is active.
func (volume *Volume) Active() bool {
	return volume.active
}

// SetNormalizationFactor sets the normalization factor for the Volume effect.
// This should be obtained from an AudioProperties Analysis.
func (volume *Volume) SetNormalizationFactor(normalization float64) {
	volume.normalization = normalization
}

// SetStrength sets the strength of the Volume effect to the specified percentage.
// The lowest possible value is 0.0, with 1.0 taking a 100% effect.
// The volume is altered on a sine-based easing curve.
// At over 100% volume, the sound is clipped as necessary.
func (volume *Volume) SetStrength(strength float64) *Volume {
	if strength < 0 {
		strength = 0
	}
	volume.strength = strength
	return volume
}

// Strength returns the strength of the Volume effect as a percentage.
func (volume *Volume) Strength() float64 {
	return volume.strength
}

// SetSource sets the active source for the effect.
func (volume *Volume) SetSource(source io.ReadSeeker) {
	volume.Source = source
}

// // Loop is an effect that loops an incoming audio byte stream.
// type Loop struct {
// 	loopCount       int
// 	activeLoopIndex int
// 	active          bool
// 	Source          io.ReadSeeker
// }

// // NewLoop creates a new Loop effect. source is the source stream to apply this effect to.
// // If you add this effect to a DSPChannel, source can be nil, as it will take effect for whatever
// // streams are played through the DSPChannel.
// func NewLoop(source io.ReadSeeker) *Loop {
// 	volume := &Loop{Source: source, loopCount: -1}
// 	return volume
// }

// // Clone clones the effect, returning an resound.IEffect.
// func (loop *Loop) Clone() resound.IEffect {
// 	return &Loop{
// 		loopCount:       loop.loopCount,
// 		activeLoopIndex: loop.activeLoopIndex,
// 	}
// }

// func (loop *Loop) Read(p []byte) (n int, err error) {

// 	n, err = loop.Source.Read(p)
// 	if err != nil {
// 		if loop.Source != nil {
// 			loop.Seek(0, io.SeekStart)
// 		} else {
// 			return 0, err
// 		}
// 	}

// 	return n, nil

// }

// func (loop *Loop) ApplyEffect(p []byte) {
// 	// The loop effect doesn't actually do anything to the source audio.
// }

// func (loop *Loop) Seek(offset int64, whence int) (int64, error) {
// 	if loop.Source == nil {
// 		return 0, nil
// 	}
// 	return loop.Source.Seek(offset, whence)
// }

// // SetActive sets the effect to be active.
// func (loop *Loop) SetActive(active bool) {
// 	loop.active = active
// }

// // Active returns if the effect is active.
// func (loop *Loop) Active() bool {
// 	return loop.active
// }

// func (loop *Loop) SetSource(source io.ReadSeeker) {
// 	loop.Source = source
// }

// Pan is a panning effect, handling panning the sound between the left and right channels.
type Pan struct {
	pan    float64
	active bool
	Source io.ReadSeeker
}

// NewPan creates a new Pan effect. source is the source stream to apply the
// effect on. Panning defaults to 0.
// If you add this effect to a DSPChannel, source can be nil, as it will take effect for whatever
// streams are played through the DSPChannel.
func NewPan(source io.ReadSeeker) *Pan {

	pan := &Pan{Source: source, active: true}
	return pan

}

// Clone clones the effect, returning an resound.IEffect.
func (pan *Pan) Clone() resound.IEffect {
	return &Pan{
		pan:    pan.pan,
		active: pan.active,
		Source: pan.Source,
	}
}

func (pan *Pan) Read(p []byte) (n int, err error) {

	if n, err = pan.Source.Read(p); err != nil {
		return
	}

	pan.ApplyEffect(p, n)

	return
}

func (pan *Pan) ApplyEffect(p []byte, bytesRead int) {

	if !pan.active {
		return
	}

	if pan.pan < -1 {
		pan.pan = -1
	} else if pan.pan > 1 {
		pan.pan = 1
	}

	// This implementation uses a linear scale, ranging from -1 to 1, for stereo or mono sounds.
	// If pan = 0.0, the balance for the sound in each speaker is at 100% left and 100% right.
	// When pan is -1.0, only the left channel of the stereo sound is audible, when pan is 1.0,
	// only the right channel of the stereo sound is audible.
	// https://docs.unity3d.com/ScriptReference/AudioSource-panStereo.html
	ls := math.Min(pan.pan*-1+1, 1)
	rs := math.Min(pan.pan+1, 1)

	audio := resound.AudioBuffer(p)

	for i := 0; i < bytesRead/4; i++ {

		l, r := audio.Get(i)

		l *= ls
		r *= rs

		audio.Set(i, l, r)

	}

}

func (pan *Pan) Seek(offset int64, whence int) (int64, error) {
	if pan.Source == nil {
		return 0, nil
	}
	return pan.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (pan *Pan) SetActive(active bool) *Pan {
	pan.active = active
	return pan
}

// Active returns if the effect is active.
func (pan *Pan) Active() bool {
	return pan.active
}

// SetPan sets the panning percentage for the pan effect.
// The possible values range from -1 (hard left) to 1 (hard right).
func (pan *Pan) SetPan(panPercent float64) *Pan {
	if panPercent > 1 {
		panPercent = 1
	} else if panPercent < -1 {
		panPercent = -1
	}
	pan.pan = panPercent
	return pan
}

// Pan returns the panning value for the pan effect in a percentage, ranging from -1 (hard left) to 1 (hard right).
func (pan *Pan) Pan() float64 {
	return pan.pan
}

// SetSource sets the active source for the effect.
func (pan *Pan) SetSource(source io.ReadSeeker) {
	pan.Source = source
}

// Delay is an effect that adds a delay to the sound.
type Delay struct {
	wait     float64
	strength float64
	feedback float64
	Source   io.ReadSeeker

	active bool
	buffer [][2]float64
}

// NewDelay creates a new Delay effect.
// If you add this effect to a DSPChannel, source can be nil, as it will take effect for whatever
// streams are played through the DSPChannel.
func NewDelay(source io.ReadSeeker) *Delay {

	return &Delay{
		Source:   source,
		wait:     0.1,
		strength: 1.0,
		feedback: 0.5,
		buffer:   [][2]float64{},
		active:   true,
	}

}

// Clone creates a clone of the Delay effect.
func (delay *Delay) Clone() resound.IEffect {
	return &Delay{
		wait:     delay.wait,
		strength: delay.strength,
		Source:   delay.Source,
		feedback: delay.feedback,
		active:   delay.active,
	}
}

func (delay *Delay) Read(p []byte) (n int, err error) {

	if n, err = delay.Source.Read(p); err != nil {
		return
	}

	delay.ApplyEffect(p, n)

	return
}

func (delay *Delay) ApplyEffect(p []byte, bytesRead int) {

	sampleRate := audio.CurrentContext().SampleRate()

	audio := resound.AudioBuffer(p)

	for i := 0; i < bytesRead/4; i++ {

		l, r := audio.Get(i)

		bl := l
		br := r

		if len(delay.buffer) > 0 {

			bl += delay.buffer[0][0] * delay.feedback
			br += delay.buffer[0][1] * delay.feedback
			// l = bl
			// r = br
			l = mix(l, bl, delay.strength)
			r = mix(r, br, delay.strength)

		}

		delay.buffer = append(delay.buffer, [2]float64{bl, br})

		// 44100 For example
		if len(delay.buffer) > int(float64(sampleRate)*delay.wait) {
			delay.buffer = delay.buffer[1:]
		}

		if delay.active {
			audio.Set(i, l, r)
		}

	}

}

func (delay *Delay) Seek(offset int64, whence int) (int64, error) {
	if delay.Source == nil {
		return 0, nil
	}
	return delay.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (delay *Delay) SetActive(active bool) *Delay {
	delay.active = active
	return delay
}

// Active returns if the effect is active.
func (delay *Delay) Active() bool {
	return delay.active
}

// SetWait sets the overall wait time of the Delay effect in seconds as it's added on top of the original signal.
// 0 is the minimum value.
func (delay *Delay) SetWait(waitTime float64) *Delay {
	if waitTime < 0 {
		waitTime = 0
	}
	delay.wait = waitTime
	return delay
}

// Wait returns the wait time of the Delay effect.
func (delay *Delay) Wait() float64 {
	return delay.wait
}

// SetStrength sets the overall volume of the Delay effect as it's added on top of the original signal.
// 0 is the minimum value.
func (delay *Delay) SetStrength(strength float64) *Delay {
	if strength < 0 {
		strength = 0
	}
	delay.strength = strength
	return delay
}

// Strength returns the strength of the Delay effect.
func (delay *Delay) Strength() float64 {
	return delay.strength
}

// SetFeedback sets the feedback percentage of the delay. Each echo's volume is modulated by this percentage.
func (delay *Delay) SetFeedback(feedbackPercentage float64) *Delay {
	delay.feedback = clamp(feedbackPercentage, 0, 1)
	return delay
}

// Feedback returns the feedback percentage of the delay.
func (delay *Delay) Feedback() float64 {
	return delay.feedback
}

// SetSource sets the active source for the effect.
func (delay *Delay) SetSource(source io.ReadSeeker) {
	delay.Source = source
}

// Distort distorts the stream that plays through it, clipping the signal.
type Distort struct {
	Source          io.ReadSeeker
	crushPercentage float64
	active          bool
}

// NewDistort creates a new Distort effect. source is the source stream to
// apply the effect to.
// If you add this effect to a DSPChannel, you can pass nil as the source, as
// it will take effect for whatever streams are played through the DSPChannel.
func NewDistort(source io.ReadSeeker) *Distort {

	return &Distort{
		Source:          source,
		crushPercentage: 0,
		active:          true,
	}

}

// Clone clones the effect, returning an resound.IEffect.
func (distort *Distort) Clone() resound.IEffect {
	return &Distort{
		crushPercentage: distort.crushPercentage,
		Source:          distort.Source,
		active:          distort.active,
	}
}

func (distort *Distort) Read(p []byte) (n int, err error) {

	if n, err = distort.Source.Read(p); err != nil {
		return
	}

	distort.ApplyEffect(p, n)

	return
}

func (distort *Distort) ApplyEffect(p []byte, bytesRead int) {

	if !distort.active || distort.crushPercentage <= 0 {
		return
	}

	audio := resound.AudioBuffer(p)

	for i := 0; i < bytesRead/4; i++ {

		l, r := audio.Get(i)

		if math.Abs(l) < distort.crushPercentage {
			l = math.Round(l)
		}

		if math.Abs(r) < distort.crushPercentage {
			r = math.Round(r)
		}

		audio.Set(i, l, r)

	}

}

func (distort *Distort) Seek(offset int64, whence int) (int64, error) {
	if distort.Source == nil {
		return 0, nil
	}
	return distort.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (distort *Distort) SetActive(active bool) *Distort {
	distort.active = active
	return distort
}

// Active returns if the effect is active.
func (distort *Distort) Active() bool {
	return distort.active
}

// CrushPercentage returns the crush percentage of the Distort effect.
func (distort *Distort) CrushPercentage() float64 {
	return distort.crushPercentage
}

// SetCrushPercentage sets the overall crush percentage of the Distort effect.
// Any value below this in percentage amplitude is rounded off.
// 0 is the minimum value.
func (distort *Distort) SetCrushPercentage(strength float64) *Distort {
	strength = clamp(strength, 0, 1)
	distort.crushPercentage = strength
	return distort
}

// SetSource sets the active source for the effect.
func (distort *Distort) SetSource(source io.ReadSeeker) {
	distort.Source = source
}

// LowpassFilter represents a low-pass filter for a source audio stream.
type LowpassFilter struct {
	Source    io.ReadSeeker
	active    bool
	prevLeft  float64
	prevRight float64
	strength  float64
}

// NewLowpassFilter creates a new low-pass filter for the given source stream.
// If you add this effect to a DSPChannel, there's no need to pass a source, as
// it will take effect for whatever streams are played through the DSPChannel.
func NewLowpassFilter(source io.ReadSeeker) *LowpassFilter {

	return &LowpassFilter{
		Source:   source,
		strength: 0.5,
		active:   true,
	}

}

// Clone clones the effect, returning an resound.IEffect.
func (lpf *LowpassFilter) Clone() resound.IEffect {
	return &LowpassFilter{
		strength: lpf.strength,
		Source:   lpf.Source,
		active:   lpf.active,
	}
}

func (lpf *LowpassFilter) Read(p []byte) (n int, err error) {

	if n, err = lpf.Source.Read(p); err != nil {
		return
	}

	lpf.ApplyEffect(p, n)

	return

}

func (lpf *LowpassFilter) ApplyEffect(p []byte, bytesRead int) {

	if !lpf.active {
		return
	}

	alpha := math.Sin(lpf.strength * math.Pi / 2)
	audio := resound.AudioBuffer(p)

	// TODO: Make low-pass / high-pass filters better quality.
	for i := 0; i < bytesRead/4; i++ {

		l, r := audio.Get(i)

		l = (1-alpha)*l + (lpf.prevLeft * alpha)
		r = (1-alpha)*r + (lpf.prevRight * alpha)

		lpf.prevLeft = l
		lpf.prevRight = r

		audio.Set(i, l, r)

	}

}

func (lpf *LowpassFilter) Seek(offset int64, whence int) (int64, error) {
	if lpf.Source == nil {
		return 0, nil
	}
	return lpf.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (lpf *LowpassFilter) SetActive(active bool) *LowpassFilter {
	lpf.active = active
	return lpf
}

// Active returns if the effect is active.
func (lpf *LowpassFilter) Active() bool {
	return lpf.active
}

func (lpf *LowpassFilter) Strength() float64 {
	return lpf.strength
}

func (lpf *LowpassFilter) SetStrength(strength float64) *LowpassFilter {
	strength = clamp(strength, 0, 1)
	lpf.strength = strength
	return lpf
}

// SetSource sets the active source for the effect.
func (lpf *LowpassFilter) SetSource(source io.ReadSeeker) {
	lpf.Source = source
}

// HighpassFilter represents a highpass filter for an audio stream.
type HighpassFilter struct {
	Source   io.ReadSeeker
	active   bool
	prev     [2]float64
	strength float64
}

// NewHighpassFilter creates a new high-pass filter for the given source stream.
// If you add this effect to a DSPChannel, there's no need to pass a source, as
// it will take effect for whatever streams are played through the DSPChannel.
func NewHighpassFilter(source io.ReadSeeker) *HighpassFilter {

	return &HighpassFilter{
		Source:   source,
		strength: 0.8,
		active:   true,
	}

}

// Clone clones the effect, returning an resound.IEffect.
func (h *HighpassFilter) Clone() resound.IEffect {
	return &HighpassFilter{
		strength: h.strength,
		Source:   h.Source,
		active:   h.active,
	}
}

func (h *HighpassFilter) Read(p []byte) (n int, err error) {

	if n, err = h.Source.Read(p); err != nil {
		return
	}

	h.ApplyEffect(p, n)

	return

}

func (h *HighpassFilter) ApplyEffect(p []byte, bytesRead int) {

	if !h.active {
		return
	}

	alpha := math.Sin(h.strength * math.Pi / 2)
	audio := resound.AudioBuffer(p)

	for i := 0; i < bytesRead/4; i++ {

		l, r := audio.Get(i)

		nl := (1-alpha)*l + ((l - h.prev[0]) * alpha)
		nr := (1-alpha)*r + ((r - h.prev[1]) * alpha)

		// l = (1-alpha)*l + (h.prev[0] * alpha)
		// r = (1-alpha)*r + (h.prev[1] * alpha)

		// fmt.Println(l, r, h.prev)

		audio.Set(i, nl, nr)

		h.prev[0] = l
		h.prev[1] = r

	}

}

func (h *HighpassFilter) Seek(offset int64, whence int) (int64, error) {
	if h.Source == nil {
		return 0, nil
	}
	return h.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (h *HighpassFilter) SetActive(active bool) *HighpassFilter {
	h.active = active
	return h
}

// Active returns if the effect is active.
func (h *HighpassFilter) Active() bool {
	return h.active
}

// SetStrength sets the strength of the HighpassFilter.
// The values are clamped from 0 to 1 (100%).
func (h *HighpassFilter) SetStrength(strength float64) *HighpassFilter {
	h.strength = clamp(strength, 0, 1)
	return h
}

// Strength returns the strength of the HighpassFilter.
func (h *HighpassFilter) Strength() float64 {
	return h.strength
}

// SetSource sets the active source for the effect.
func (h *HighpassFilter) SetSource(source io.ReadSeeker) {
	h.Source = source
}

// Bitcrush is an effect that changes the pitch of the incoming audio byte stream.
type Bitcrush struct {
	strength float64
	active   bool
	Source   io.ReadSeeker
}

// NewBitcrush creates a new Bitcrush effect.
// source is the source stream to apply this effect to.
// If you add this effect to a DSPChannel, there's no need to pass a source, as
// it will take effect for whatever streams are played through the DSPChannel.
func NewBitcrush(source io.ReadSeeker) *Bitcrush {
	bitcrush := &Bitcrush{Source: source, active: true, strength: 0.1}
	return bitcrush
}

// Clone clones the effect, returning an resound.IEffect.
func (bitcrush *Bitcrush) Clone() resound.IEffect {
	return &Bitcrush{
		strength: bitcrush.strength,
		active:   bitcrush.active,
		Source:   bitcrush.Source,
	}
}

func (bitcrush *Bitcrush) Read(p []byte) (n int, err error) {

	if n, err = bitcrush.Source.Read(p); err != nil {
		return
	}

	bitcrush.ApplyEffect(p, n)

	return
}

func (bitcrush *Bitcrush) ApplyEffect(p []byte, bytesRead int) {

	if !bitcrush.active || bitcrush.strength == 0 {
		return
	}

	audio := resound.AudioBuffer(p)

	s := ease.InExpo(float32(bitcrush.strength), 0, 1, 1)

	str := float64(s) * 1000

	bufferSize := bytesRead / 4

	// str := (bitcrush.strength) * 1000

	for i := 0; i < bufferSize; i++ {

		ri := int(math.Round(float64(i)/str) * str)

		if ri >= bufferSize {
			ri = bufferSize - 1
		}

		l, r := audio.Get(ri)
		audio.Set(i, l, r)

	}

}

func (bitcrush *Bitcrush) Seek(offset int64, whence int) (int64, error) {
	if bitcrush.Source == nil {
		return 0, nil
	}
	return bitcrush.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (bitcrush *Bitcrush) SetActive(active bool) *Bitcrush {
	bitcrush.active = active
	return bitcrush
}

// Active returns if the effect is active.
func (bitcrush *Bitcrush) Active() bool {
	return bitcrush.active
}

// Strength returns the strength of the Bitcrush effect as a percentage.
func (bitcrush *Bitcrush) Strength() float64 {
	return bitcrush.strength
}

// SetStrength sets the strength of the Bitcrush effect to the specified percentage.
func (bitcrush *Bitcrush) SetStrength(bitcrushFactor float64) *Bitcrush {
	bitcrush.strength = clamp(bitcrushFactor, 0, 1)
	return bitcrush
}

// SetSource sets the active source for the effect.
func (bitcrush *Bitcrush) SetSource(source io.ReadSeeker) {
	bitcrush.Source = source
}

type circularBuffer struct {
	buffer     [][2]float64
	maxSize    int
	readIndex  float64
	writeIndex int
}

func newCircularBuffer(maxSize int) circularBuffer {
	return circularBuffer{
		maxSize: maxSize,
		buffer:  make([][2]float64, maxSize),
	}
}

func (c *circularBuffer) write(l, r float64) {

	c.buffer[c.writeIndex][0] = l
	c.buffer[c.writeIndex][1] = r
	c.writeIndex++

	if c.writeIndex >= c.maxSize {
		c.writeIndex = 0
	}

}

// readWriteDistance returns the distance between the write index and read index; note that this is the
// shortest distance.
func (c circularBuffer) readWriteDistance() float64 {
	r := c.readIndex
	w := float64(c.writeIndex)
	m := float64(c.maxSize)
	distance := math.Abs(r - w)
	if distance > m/2 {
		return math.Min(math.Abs(m-r)+w, math.Abs(m-w)+r)
	}
	return distance
}

func (c *circularBuffer) incrementRead(value float64) {
	c.readIndex += value
	bufSize := float64(len(c.buffer))
	if c.readIndex >= bufSize {
		c.readIndex -= bufSize
	}
}

func (c circularBuffer) read(offset int) (l, r float64) {
	if !c.BufferFull() {
		return 0, 0
	}
	readIndex := int(c.readIndex) + offset
	if readIndex >= c.maxSize {
		readIndex -= c.maxSize
	}
	return c.buffer[readIndex][0], c.buffer[readIndex][1]
}

func (c circularBuffer) BufferFull() bool {
	return len(c.buffer) == c.maxSize
}

// PitchShift is an effect that changes the pitch of the incoming audio stream.
type PitchShift struct {
	strength float64
	pitch    float64
	active   bool
	Source   io.ReadSeeker

	pitchBuffer circularBuffer
}

// −12log2(t1/t2) = how many semitones

// NewPitchShift creates a new PitchShift effect.
// source is the source stream to apply this effect to and bufferSize is the size of the buffer the pitch shift effect operates on.
// The larger the buffer, the smoother it will sound, but the more echoing there will be as the effect runs through the buffer.
// A buffer size of 1024, 2048, or 4096 are good starting points.
// If you add this effect to a DSPChannel, source can be nil, as it will take effect for whatever
// streams are played through the DSPChannel.
func NewPitchShift(source io.ReadSeeker, bufferSize int) *PitchShift {
	pitchShift := &PitchShift{
		Source:      source,
		strength:    1,
		active:      true,
		pitch:       1,
		pitchBuffer: newCircularBuffer(bufferSize),
	}
	return pitchShift
}

// Clone clones the effect, returning an resound.IEffect.
func (p *PitchShift) Clone() resound.IEffect {
	return &PitchShift{
		strength: p.strength,
		pitch:    p.pitch,
		active:   p.active,
		Source:   p.Source,
	}
}

func (p *PitchShift) Read(byteSlice []byte) (n int, err error) {

	if n, err = p.Source.Read(byteSlice); err != nil {
		return
	}

	p.ApplyEffect(byteSlice, n)

	return
}

func (p *PitchShift) ApplyEffect(byteSlice []byte, bytesRead int) {

	// If the effect isn't active, then we can return early.
	if !p.active {
		return
	}

	audio := resound.AudioBuffer(byteSlice)
	bufferLength := bytesRead / 4

	for i := 0; i < bufferLength; i++ {
		// Get the audio value:
		l, r := audio.Get(i)

		// Write the unaltered audio to the pitch buffer.
		p.pitchBuffer.write(l, r)

		// Reading from the buffer slower or faster than 1 per frame will give us a pitched result.
		pitchedL, pitchedR := p.pitchBuffer.read(0)

		// After we do this, we could just increment the read index by pitch (so higher pitch values increment
		// faster and lower values slower, giving higher pitch and lower pitch), but this alone would give
		// a crackling sound.

		// To keep from having crackling audio (which can happen when the buffer reads old values
		// or when the write head passes the read head), we will read the audio from the circular pitch buffer
		// twice and then mix the result. We read once where the read index is, and once from the opposite side.

		// By cross-fading between these two points based on the distance between the read and
		// write indices, we can avoid crackling.

		// For more information, see the following (extremely) helpful sites:
		// https://en.wikipedia.org/wiki/Circular_buffer
		// https://schaumont.dyn.wpi.edu/ece4703b22/lab5x.html
		// https://people.ece.cornell.edu/land/courses/ece5760/FinalProjects/s2017/jmt329_swc63_gzm3/jmt329_swc63_gzm3/PitchShifter/index.html

		pitchedL2, pitchedR2 := p.pitchBuffer.read(p.pitchBuffer.maxSize / 2)
		cross := p.pitchBuffer.readWriteDistance() / float64(p.pitchBuffer.maxSize/2)
		cross2 := 1 - cross

		fl := pitchedL*cross + pitchedL2*cross2
		fr := pitchedR*cross + pitchedR2*cross2

		audio.Set(i, mix(l, fl, p.strength), mix(r, fr, p.strength))

		p.pitchBuffer.incrementRead(p.pitch)

	}

}

func (p *PitchShift) Seek(offset int64, whence int) (int64, error) {
	if p.Source == nil {
		return 0, nil
	}
	return p.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (p *PitchShift) SetActive(active bool) *PitchShift {
	p.active = active
	return p
}

// Active returns if the effect is active.
func (p *PitchShift) Active() bool {
	return p.active
}

// SetStrength sets the strength of the PitchShift effect to the specified percentage.
// The lowest possible value is 0.0, with 1.0 being the maximum and taking a 100% effect.
func (p *PitchShift) SetStrength(strength float64) *PitchShift {
	if strength < 0 {
		strength = 0
	}
	if strength > 1 {
		strength = 1
	}
	p.strength = strength
	return p
}

// Strength returns the strength of the PitchShift effect as a percentage. The value ranges from 0 to 1.
func (p *PitchShift) Strength() float64 {
	return p.strength
}

// SetSource sets the active source for the effect.
func (p *PitchShift) SetSource(source io.ReadSeeker) {
	p.Source = source
}

// SetPitch sets the target pitch of the PitchShift effect to the specified percentage.
// The lowest possible value is 0.0, with 1.0 being 100% pitch.
func (p *PitchShift) SetPitch(pitchFactor float64) *PitchShift {
	if pitchFactor < 0 {
		pitchFactor = 0
	}
	p.pitch = pitchFactor
	return p
}

// Pitch returns the pitch of the PitchShift effect as a percentage.
func (p *PitchShift) Pitch() float64 {
	return p.pitch
}

// type Reverb struct {
// 	FeedbackLoop bool
// 	Source       io.ReadSeeker

// 	active     bool
// 	buffer     [][2]int16
// 	bufferSize int
// }

// func NewReverb(source io.ReadSeeker, reverbBufferSize int) *Reverb {

// 	return &Reverb{
// 		Source:     source,
// 		buffer:     make([][2]int16, reverbBufferSize),
// 		bufferSize: reverbBufferSize,
// 		active:     true,
// 	}

// }

// func (reverb *Reverb) Clone() resound.IEffect {
// 	return &Reverb{
// 		FeedbackLoop: reverb.FeedbackLoop,
// 		Source:       reverb.Source,

// 		active: reverb.active,
// 	}
// }

// func (reverb *Reverb) Read(p []byte) (n int, err error) {

// 	_, err = reverb.Source.Read(p)
// 	if err != nil {
// 		return 0, err
// 	}

// 	reverb.ApplyEffect(p)

// 	return len(p), nil

// }

// func (reverb *Reverb) ApplyEffect(p []byte) {

// 	for i := 0; i < len(p); i += 4 {
// 		lc := int16(p[i]) | int16(p[i+1])<<8
// 		rc := int16(p[i+2]) | int16(p[i+3])<<8

// 		reverb.buffer = append(reverb.buffer, [2]int16{lc, rc})

// 		// 44100 For example
// 		if len(reverb.buffer) > reverb.bufferSize {
// 			reverb.buffer = reverb.buffer[1:]
// 		}

// 		l := 0.0
// 		r := 0.0

// 		for i := 0; i < len(reverb.buffer); i++ {
// 			l += float64(reverb.buffer[i][0])
// 			r += float64(reverb.buffer[i][1])
// 		}

// 		l /= float64(len(reverb.buffer))
// 		r /= float64(len(reverb.buffer))

// 		if reverb.active {

// 			lc = int16(l)
// 			rc = int16(r)

// 		}

// 		p[i] = byte(lc)
// 		p[i+1] = byte(lc >> 8)
// 		p[i+2] = byte(rc)
// 		p[i+3] = byte(rc >> 8)

// 	}

// }

// func (reverb *Reverb) Seek(offset int64, whence int) (int64, error) {
// 	if reverb.Source == nil {
// 		return 0, nil
// 	}
// 	return reverb.Source.Seek(offset, whence)
// }

// func (reverb *Reverb) SetActive(active bool) {
// 	reverb.active = active
// }

// func (reverb *Reverb) Active() bool {
// 	return reverb.active
// }

func clamp(v, min, max float64) float64 {
	if v > max {
		return max
	} else if v < min {
		return min
	}
	return v
}

func mix(v1, v2, perc float64) float64 {
	return v1 + ((v2 - v1) * perc)
}
