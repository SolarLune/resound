package resound

import (
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// DSPPlayer embeds audio.Player and so has all of the functions and abilities of the default audio.Player
// while also applying effects placed on the source DSPChannel.
type DSPPlayer struct {
	*audio.Player
	Channel *DSPChannel
	Source  io.ReadSeeker
}

func newChannelPlayback(sourceStream io.ReadSeeker, channel *DSPChannel) *DSPPlayer {

	cp := &DSPPlayer{
		Channel: channel,
		Source:  sourceStream,
	}

	player, err := audio.CurrentContext().NewPlayer(cp)

	if err != nil {
		panic(err)
	}

	cp.Player = player

	return cp

}

// Clone duplicates the DSPPlayer; note that the current playback values will not be cloned.
func (es *DSPPlayer) Clone() *DSPPlayer {
	newES := newChannelPlayback(es.Source, es.Channel)
	return newES
}

func (es *DSPPlayer) Read(p []byte) (n int, err error) {

	n, err = es.Source.Read(p)

	if err != nil || len(es.Channel.EffectOrder) == 0 || !es.Channel.Active {
		return n, err
	}

	for _, effect := range es.Channel.EffectOrder {
		effect.ApplyEffect(p)
	}

	return n, nil

}

func (es *DSPPlayer) Seek(offset int64, whence int) (int64, error) {

	if es.Source == nil {
		return 0, nil
	}

	return es.Source.Seek(offset, whence)

}

// DSPChannel represents a channel that can have various effects applied to it.
type DSPChannel struct {
	Active      bool
	Effects     map[string]IEffect
	EffectOrder []IEffect
}

// NewDSPChannel returns a new DSPChannel.
func NewDSPChannel() *DSPChannel {
	dsp := &DSPChannel{
		Active:      true,
		Effects:     map[string]IEffect{},
		EffectOrder: []IEffect{},
	}
	return dsp
}

// Add adds the specified Effect to the DSPChannel under the given name. Note that effects added to DSPChannels don't need
// to specify source streams, as the DSPChannel automatically handles this.
func (dsp *DSPChannel) Add(name string, effect IEffect) *DSPChannel {
	dsp.Effects[name] = effect
	dsp.EffectOrder = append(dsp.EffectOrder, effect)
	return dsp
}

// CreatePlayer creates a new DSPPlayer to handle playback of a stream through the DSPChannel.
func (dsp *DSPChannel) CreatePlayer(sourceStream io.ReadSeeker) *DSPPlayer {
	playback := newChannelPlayback(sourceStream, dsp)
	return playback
}
