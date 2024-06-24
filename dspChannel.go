package resound

// DSPChannel represents an audio channel that can have various effects applied to it.
// Any Players that have a DSPChannel set will take on the effects applied to the channel as well.
type DSPChannel struct {
	Active      bool
	Effects     map[any]IEffect
	EffectOrder []IEffect
	closed      bool
}

// NewDSPChannel returns a new DSPChannel.
func NewDSPChannel() *DSPChannel {
	dsp := &DSPChannel{
		Active:      true,
		Effects:     map[any]IEffect{},
		EffectOrder: []IEffect{},
	}
	return dsp
}

// Close closes the DSP channel. When closed, any players that play on the channel do not play and automatically close their sources.
// Closing the channel can be used to stop any sounds that might be playing back on the DSPChannel.
func (d *DSPChannel) Close() {
	d.closed = true
}

// AddEffect adds the specified Effect to the DSPChannel under the given identification. Note that effects added to DSPChannels don't need
// to specify source streams, as the DSPChannel automatically handles this.
func (d *DSPChannel) AddEffect(id any, effect IEffect) *DSPChannel {
	d.Effects[id] = effect
	d.EffectOrder = append(d.EffectOrder, effect)
	return d
}
