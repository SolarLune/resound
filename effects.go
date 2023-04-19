package resound

import (
	"io"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/tanema/gween/ease"
)

// IEffect indicates an effect that implements io.ReadSeeker and generally takes effect on an existing audio stream.
// It represents the result of applying an effect to an audio stream, and is playable in its own right.
type IEffect interface {
	io.ReadSeeker
	applyEffect(data []byte)
	Active() bool          // Active returns if the IEffect is active or not. When inactive, it applies no effect to the incoming audio stream.
	SetActive(active bool) // SetActive sets the Effect's active state.
	Clone() IEffect        // Clone returns a clone of the IEffect.
}

// Volume is an effect that changes the overall volume of the incoming audio byte stream.
type Volume struct {
	strength float64
	active   bool
	Source   io.ReadSeeker
}

// NewVolume creates a new Volume effect. source is the source stream to apply this
// effect to, and percent is the strength percentage, ranging from 0 to 1, to indicate how
// strongly the volume should be altered. You can over-amplify the sound by pushing
// the volume above 1 - otherwise, the volume is altered on a sine-based easing curve.
// If you add this effect to a DSPChannel, there's no need to pass a source, as
// it will take effect for whatever streams are played through the DSPChannel.
func NewVolume(source io.ReadSeeker, strength float64) *Volume {

	if strength < 0 {
		strength = 0
	}

	return &Volume{Source: source, strength: strength, active: true}

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

	volume.applyEffect(p)

	return n, nil

}

func (volume *Volume) applyEffect(p []byte) {

	if !volume.active {
		return
	}

	perc := volume.strength
	if volume.strength <= 1 {
		perc = float64(ease.InSine(float32(volume.strength), 0, 1, 1))
	}

	for i := 0; i < len(p); i += 4 {
		lc := float64(int16(p[i])|int16(p[i+1])<<8) * perc
		rc := float64(int16(p[i+2])|int16(p[i+3])<<8) * perc

		if lc > math.MaxInt16 {
			lc = math.MaxInt16
		}

		if rc > math.MaxInt16 {
			rc = math.MaxInt16
		}

		lcc := int16(lc)
		rcc := int16(rc)

		p[i] = byte(lcc)
		p[i+1] = byte(lcc >> 8)
		p[i+2] = byte(rcc)
		p[i+3] = byte(rcc >> 8)
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

// Strength returns the strength of the Volume effect as a percentage.
func (volume *Volume) Strength() float64 {
	return volume.strength
}

// SetStrength sets the strength of the Volume effect to the specified percentage.
// The lowest possible value is 0.0, with 1.0 taking a 100% effect.
// At over 100% volume, the sound is clipped as necessary.
func (volume *Volume) SetStrength(strength float64) {
	if strength < 0 {
		strength = 0
	}
	volume.strength = strength
}

// Pan is a panning effect, handling panning the sound between the left and right channels.
type Pan struct {
	strength float64
	active   bool
	Source   io.ReadSeeker
}

// NewPan creates a new Pan effect. source is the source stream to apply the
// effect on, and panPercentage is the percentage of the panning effect in percentage,
// ranging from -1 (left channel only), to 1 (right channel only) the pan should take effect over.
// If you add this effect to a DSPChannel, there's no need to pass a source, as
// it will take effect for whatever streams are played through the DSPChannel.
func NewPan(source io.ReadSeeker, panPercentage float64) *Pan {

	pan := &Pan{Source: source, strength: panPercentage, active: true}
	pan.SetPan(panPercentage)
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

	pan.applyEffect(p)

	return len(p), nil

}

func (pan *Pan) applyEffect(p []byte) {

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
	for i := 0; i < len(p); i += 4 {
		lc := int16(float64(int16(p[i])|int16(p[i+1])<<8) * ls)
		rc := int16(float64(int16(p[i+2])|int16(p[i+3])<<8) * rs)

		p[i] = byte(lc)
		p[i+1] = byte(lc >> 8)
		p[i+2] = byte(rc)
		p[i+3] = byte(rc >> 8)
	}

}

func (pan *Pan) Seek(offset int64, whence int) (int64, error) {
	if pan.Source == nil {
		return 0, nil
	}
	return pan.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (pan *Pan) SetActive(active bool) {
	pan.active = active
}

// Active returns if the effect is active.
func (pan *Pan) Active() bool {
	return pan.active
}

// Pan returns the panning value for the pan effect in a percentage, ranging from -1 (hard left) to 1 (hard right).
func (pan *Pan) Pan() float64 {
	return pan.strength
}

// SetPan sets the panning percentage for the pan effect.
// The possible values range from -1 (hard left) to 1 (hard right).
func (pan *Pan) SetPan(panPercent float64) {
	if panPercent > 1 {
		panPercent = 1
	} else if panPercent < -1 {
		panPercent = -1
	}
	pan.strength = panPercent
}

// Delay is an effect that adds a delay to the sound.
type Delay struct {
	wait         float64
	strength     float64
	feedbackLoop bool
	Source       io.ReadSeeker

	active bool
	buffer [][2]int16
}

// NewDelay creates a new Delay effect. The first and second
// arguments are how many seconds should pass between the initial sound and
// the delay, and how loud (in percentage) the delay should be. The last argument,
// feedbackLoop, is if the delay should feedback into itself.
// If you add this effect to a DSPChannel, there's no need to pass a source, as
// it will take effect for whatever streams are played through the DSPChannel.
func NewDelay(source io.ReadSeeker, delayWait, delayStrength float64, feedbackLoop bool) *Delay {

	return &Delay{
		Source:       source,
		wait:         delayWait,
		strength:     delayStrength,
		feedbackLoop: feedbackLoop,
		buffer:       [][2]int16{},
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

	delay.applyEffect(p)

	return len(p), nil

}

func (delay *Delay) applyEffect(p []byte) {

	if !delay.active {
		return
	}

	sampleRate := audio.CurrentContext().SampleRate()

	for i := 0; i < len(p); i += 4 {
		lc := int16(p[i]) | int16(p[i+1])<<8
		rc := int16(p[i+2]) | int16(p[i+3])<<8

		if delay.feedbackLoop {

			if len(delay.buffer) > 0 {
				lc = addChannelValue(lc, int16(float64(delay.buffer[0][0])*delay.strength))
				rc = addChannelValue(rc, int16(float64(delay.buffer[0][1])*delay.strength))
			}

			delay.buffer = append(delay.buffer, [2]int16{lc, rc})

		} else {

			delay.buffer = append(delay.buffer, [2]int16{lc, rc})
			lc = addChannelValue(lc, int16(float64(delay.buffer[0][0])*delay.strength))
			rc = addChannelValue(rc, int16(float64(delay.buffer[0][1])*delay.strength))

		}

		// 44100 For example
		if len(delay.buffer) > int(float64(sampleRate)*delay.wait) {
			delay.buffer = delay.buffer[1:]
		}

		p[i] = byte(lc)
		p[i+1] = byte(lc >> 8)
		p[i+2] = byte(rc)
		p[i+3] = byte(rc >> 8)

	}

}

func (delay *Delay) Seek(offset int64, whence int) (int64, error) {
	if delay.Source == nil {
		return 0, nil
	}
	return delay.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (delay *Delay) SetActive(active bool) {
	delay.active = active
}

// Active returns if the effect is active.
func (delay *Delay) Active() bool {
	return delay.active
}

// Wait returns the wait time of the Delay effect.
func (delay *Delay) Wait() float64 {
	return delay.wait
}

// SetWait sets the overall wait time of the Delay effect in seconds as it's added on top of the original signal.
// 0 is the minimum value.
func (delay *Delay) SetWait(waitTime float64) {
	if waitTime < 0 {
		waitTime = 0
	}
	delay.wait = waitTime
}

// Strength returns the strength of the Delay effect.
func (delay *Delay) Strength() float64 {
	return delay.strength
}

// SetStrength sets the overall volume of the Delay effect as it's added on top of the original signal.
// 0 is the minimum value.
func (delay *Delay) SetStrength(strength float64) {
	if strength < 0 {
		strength = 0
	}
	delay.strength = strength
}

// FeedbackLoop returns if the delay's results feed back into itself or not.
func (delay *Delay) FeedbackLoop() bool {
	return delay.feedbackLoop
}

// SetFeedbackLoop sets the feedback loop of the delay. If set to on, the delay's results feed back into itself.
func (delay *Delay) SetFeedbackLoop(on bool) {
	delay.feedbackLoop = on
}

type Distort struct {
	Source   io.ReadSeeker
	strength float64
	active   bool
}

// NewDistort creates a new Distort effect. source is the source stream to
// apply the effect to, and the percent is a percentage ranging from 0
// to 1 indicating how strongly the distortion should be.
// If you add this effect to a DSPChannel, you can pass nil as the source, as
// it will take effect for whatever streams are played through the DSPChannel.
func NewDistort(source io.ReadSeeker, strength float64) *Distort {

	return &Distort{
		Source:   source,
		strength: strength,
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

	distort.applyEffect(p)

	return len(p), nil

}

func (distort *Distort) applyEffect(p []byte) {

	if !distort.active {
		return
	}

	clipMax := float64(math.MaxInt16) * distort.strength

	if clipMax < 1 {
		clipMax = 1
	}

	for i := 0; i < len(p); i += 4 {

		lc := int16(p[i]) | int16(p[i+1])<<8
		rc := int16(p[i+2]) | int16(p[i+3])<<8

		lc = int16(math.Floor(float64(lc)/clipMax) * clipMax)
		rc = int16(math.Floor(float64(rc)/clipMax) * clipMax)

		p[i] = byte(lc)
		p[i+1] = byte(lc >> 8)
		p[i+2] = byte(rc)
		p[i+3] = byte(rc >> 8)

	}

}

func (distort *Distort) Seek(offset int64, whence int) (int64, error) {
	if distort.Source == nil {
		return 0, nil
	}
	return distort.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (distort *Distort) SetActive(active bool) {
	distort.active = active
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
func (distort *Distort) SetStrength(strength float64) {
	if strength < 0 {
		strength = 0
	}
	distort.strength = strength
}

type LowpassFilter struct {
	Source    io.ReadSeeker
	active    bool
	prevLeft  float64
	prevRight float64
	strength  float64
}

// NewLowpassFilter creates a new low-pass filter for the given source stream.
// filterPercentage, ranging from 0 (un-filtered) to 1 (fully filtered), indicates
// how strongly the stream should be filtered.
// If you add this effect to a DSPChannel, there's no need to pass a source, as
// it will take effect for whatever streams are played through the DSPChannel.
func NewLowpassFilter(source io.ReadSeeker, filterPercentage float64) *LowpassFilter {

	return &LowpassFilter{
		Source:   source,
		strength: filterPercentage,
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

	lpf.applyEffect(p)

	return len(p), nil

}

func (lpf *LowpassFilter) applyEffect(p []byte) {

	if !lpf.active {
		return
	}

	alpha := math.Sin(lpf.strength * math.Pi / 2)

	for i := 0; i < len(p); i += 4 {

		lc := int16(p[i]) | int16(p[i+1])<<8
		rc := int16(p[i+2]) | int16(p[i+3])<<8

		lcc := float64(lc)
		rcc := float64(rc)

		lcc = (1-alpha)*lcc + (lpf.prevLeft * alpha)
		rcc = (1-alpha)*rcc + (lpf.prevRight * alpha)

		lpf.prevLeft = lcc
		lpf.prevRight = rcc

		if lcc > math.MaxInt16 {
			lcc = math.MaxInt16
		} else if lcc < math.MinInt16 {
			lcc = math.MinInt16
		}

		if rcc > math.MaxInt16 {
			rcc = math.MaxInt16
		} else if rcc < math.MinInt16 {
			rcc = math.MinInt16
		}

		lc = int16(lcc)
		rc = int16(rcc)

		p[i] = byte(lc)
		p[i+1] = byte(lc >> 8)
		p[i+2] = byte(rc)
		p[i+3] = byte(rc >> 8)

	}

}

func (lpf *LowpassFilter) Seek(offset int64, whence int) (int64, error) {
	if lpf.Source == nil {
		return 0, nil
	}
	return lpf.Source.Seek(offset, whence)
}

// SetActive sets the effect to be active.
func (lpf *LowpassFilter) SetActive(active bool) {
	lpf.active = active
}

// Active returns if the effect is active.
func (lpf *LowpassFilter) Active() bool {
	return lpf.active
}

func (lpf *LowpassFilter) Strength() float64 {
	return lpf.strength
}

func (lpf *LowpassFilter) SetStrength(strength float64) {
	if strength < 0 {
		strength = 0
	} else if strength > 1 {
		strength = 1
	}

	lpf.strength = strength
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

// 	reverb.applyEffect(p)

// 	return len(p), nil

// }

// func (reverb *Reverb) applyEffect(p []byte) {

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
