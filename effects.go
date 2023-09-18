package resound

import (
	"io"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/tanema/gween/ease"
)

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

// IEffect indicates an effect that implements io.ReadSeeker and generally takes effect on an existing audio stream.
// It represents the result of applying an effect to an audio stream, and is playable in its own right.
type IEffect interface {
	io.ReadSeeker
	ApplyEffect(data []byte) // This function is called when sound data goes through an effect. The effect should modify the source data buffer.
	SetSource(io.ReadSeeker) // This function allows an effect's source to be dynamically altered; this allows for easy chaining with resound.ChainEffects().
}

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

// Clone clones the effect, returning an IEffect.
func (volume *Volume) Clone() IEffect {
	return &Volume{
		strength: volume.strength,
		active:   volume.active,
		Source:   volume.Source,
	}
}

func (volume *Volume) Read(p []byte) (n int, err error) {

	n, err = volume.Source.Read(p)
	if err != nil {
		return 0, err
	}

	volume.ApplyEffect(p)

	return n, nil

}

func (volume *Volume) ApplyEffect(p []byte) {

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
	audio := AudioBuffer(p)

	// Loop through all frames in the stream that are available to be read.
	for i := 0; i < audio.Len(); i++ {

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
func (volume *Volume) SetActive(active bool) {
	volume.active = active
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

// // Clone clones the effect, returning an IEffect.
// func (loop *Loop) Clone() IEffect {
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
	strength float64
	active   bool
	Source   io.ReadSeeker
}

// NewPan creates a new Pan effect. source is the source stream to apply the
// effect on. Panning defaults to 0.
// If you add this effect to a DSPChannel, source can be nil, as it will take effect for whatever
// streams are played through the DSPChannel.
func NewPan(source io.ReadSeeker) *Pan {

	pan := &Pan{Source: source, active: true}
	return pan

}

// Clone clones the effect, returning an IEffect.
func (pan *Pan) Clone() IEffect {
	return &Pan{
		strength: pan.strength,
		active:   pan.active,
		Source:   pan.Source,
	}
}

func (pan *Pan) Read(p []byte) (n int, err error) {

	_, err = pan.Source.Read(p)
	if err != nil {
		return 0, err
	}

	pan.ApplyEffect(p)

	return len(p), nil

}

func (pan *Pan) ApplyEffect(p []byte) {

	if !pan.active {
		return
	}

	if pan.strength < -1 {
		pan.strength = -1
	} else if pan.strength > 1 {
		pan.strength = 1
	}

	// This implementation uses a linear scale, ranging from -1 to 1, for stereo or mono sounds.
	// If pan = 0.0, the balance for the sound in each speaker is at 100% left and 100% right.
	// When pan is -1.0, only the left channel of the stereo sound is audible, when pan is 1.0,
	// only the right channel of the stereo sound is audible.
	// https://docs.unity3d.com/ScriptReference/AudioSource-panStereo.html
	ls := math.Min(pan.strength*-1+1, 1)
	rs := math.Min(pan.strength+1, 1)

	audio := AudioBuffer(p)

	for i := 0; i < audio.Len(); i++ {

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
	pan.strength = panPercent
	return pan
}

// Pan returns the panning value for the pan effect in a percentage, ranging from -1 (hard left) to 1 (hard right).
func (pan *Pan) Pan() float64 {
	return pan.strength
}

func (pan *Pan) SetSource(source io.ReadSeeker) {
	pan.Source = source
}

// Delay is an effect that adds a delay to the sound.
type Delay struct {
	wait         float64
	strength     float64
	feedbackLoop bool
	Source       io.ReadSeeker

	active bool
	buffer [][2]float64
}

// NewDelay creates a new Delay effect.
// If you add this effect to a DSPChannel, source can be nil, as it will take effect for whatever
// streams are played through the DSPChannel.
func NewDelay(source io.ReadSeeker) *Delay {

	return &Delay{
		Source:       source,
		wait:         0.1,
		strength:     0.75,
		feedbackLoop: false,
		buffer:       [][2]float64{},
		active:       true,
	}

}

// Clone creates a clone of the Delay effect.
func (delay *Delay) Clone() IEffect {
	return &Delay{
		wait:         delay.wait,
		strength:     delay.strength,
		feedbackLoop: delay.feedbackLoop,
		Source:       delay.Source,

		active: delay.active,
	}
}

func (delay *Delay) Read(p []byte) (n int, err error) {

	_, err = delay.Source.Read(p)
	if err != nil {
		return 0, err
	}

	delay.ApplyEffect(p)

	return len(p), nil

}

func (delay *Delay) ApplyEffect(p []byte) {

	sampleRate := audio.CurrentContext().SampleRate()

	audio := AudioBuffer(p)

	for i := 0; i < audio.Len(); i++ {

		l, r := audio.Get(i)

		if delay.feedbackLoop {

			if len(delay.buffer) > 0 {

				l += delay.buffer[0][0] * delay.strength
				r += delay.buffer[0][1] * delay.strength

			}

			delay.buffer = append(delay.buffer, [2]float64{l, r})

		} else {

			delay.buffer = append(delay.buffer, [2]float64{l, r})
			l += delay.buffer[0][0] * delay.strength
			r += delay.buffer[0][1] * delay.strength

		}

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

// SetFeedbackLoop sets the feedback loop of the delay. If set to on, the delay's results feed back into itself.
func (delay *Delay) SetFeedbackLoop(on bool) *Delay {
	delay.feedbackLoop = on
	return delay
}

// FeedbackLoop returns if the delay's results feed back into itself or not.
func (delay *Delay) FeedbackLoop() bool {
	return delay.feedbackLoop
}

func (delay *Delay) SetSource(source io.ReadSeeker) {
	delay.Source = source
}

// Distort distorts the stream that plays through it, clipping the signal.
type Distort struct {
	Source   io.ReadSeeker
	strength float64
	active   bool
}

// NewDistort creates a new Distort effect. source is the source stream to
// apply the effect to.
// If you add this effect to a DSPChannel, you can pass nil as the source, as
// it will take effect for whatever streams are played through the DSPChannel.
func NewDistort(source io.ReadSeeker) *Distort {

	return &Distort{
		Source:   source,
		strength: 0,
		active:   true,
	}

}

// Clone clones the effect, returning an IEffect.
func (distort *Distort) Clone() IEffect {
	return &Distort{
		strength: distort.strength,
		Source:   distort.Source,
		active:   distort.active,
	}
}

func (distort *Distort) Read(p []byte) (n int, err error) {

	_, err = distort.Source.Read(p)
	if err != nil {
		return 0, err
	}

	distort.ApplyEffect(p)

	return len(p), nil

}

func (distort *Distort) ApplyEffect(p []byte) {

	if !distort.active || distort.strength <= 0 {
		return
	}

	audio := AudioBuffer(p)

	for i := 0; i < audio.Len(); i++ {

		l, r := audio.Get(i)

		if math.Abs(l) < distort.strength {
			l = math.Round(l)
		}

		if math.Abs(r) < distort.strength {
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

// Strength returns the strength of the Distort effect.
func (distort *Distort) Strength() float64 {
	return distort.strength
}

// SetStrength sets the overall strength of the Distort effect. 0 is the minimum value.
func (distort *Distort) SetStrength(strength float64) *Distort {
	strength = clamp(strength, 0, 1)
	distort.strength = strength
	return distort
}

func (distort *Distort) SetSource(source io.ReadSeeker) {
	distort.Source = source
}

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

// Clone clones the effect, returning an IEffect.
func (lpf *LowpassFilter) Clone() IEffect {
	return &LowpassFilter{
		strength: lpf.strength,
		Source:   lpf.Source,
		active:   lpf.active,
	}
}

func (lpf *LowpassFilter) Read(p []byte) (n int, err error) {

	_, err = lpf.Source.Read(p)
	if err != nil {
		return 0, err
	}

	lpf.ApplyEffect(p)

	return len(p), nil

}

func (lpf *LowpassFilter) ApplyEffect(p []byte) {

	if !lpf.active {
		return
	}

	alpha := math.Sin(lpf.strength * math.Pi / 2)
	audio := AudioBuffer(p)

	for i := 0; i < audio.Len(); i++ {

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

func (lpf *LowpassFilter) SetSource(source io.ReadSeeker) {
	lpf.Source = source
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
	bitcrush := &Bitcrush{Source: source, active: true, strength: 1}
	return bitcrush
}

// Clone clones the effect, returning an IEffect.
func (bitcrush *Bitcrush) Clone() IEffect {
	return &Bitcrush{
		strength: bitcrush.strength,
		active:   bitcrush.active,
		Source:   bitcrush.Source,
	}
}

func (bitcrush *Bitcrush) Read(p []byte) (n int, err error) {

	_, err = bitcrush.Source.Read(p)
	if err != nil {
		return 0, err
	}

	bitcrush.ApplyEffect(p)

	return len(p), nil
}

func (bitcrush *Bitcrush) ApplyEffect(p []byte) {

	if !bitcrush.active || bitcrush.strength == 0 {
		return
	}

	audio := AudioBuffer(p)

	s := ease.InExpo(float32(bitcrush.strength), 0, 1, 1)

	str := float64(s) * 1000

	// str := (bitcrush.strength) * 1000

	for i := 0; i < audio.Len(); i++ {

		ri := int(math.Round(float64(i)/str) * str)

		if ri >= audio.Len() {
			ri = audio.Len() - 1
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

func (bitcrush *Bitcrush) SetSource(source io.ReadSeeker) {
	bitcrush.Source = source
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

// func (reverb *Reverb) Clone() IEffect {
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
