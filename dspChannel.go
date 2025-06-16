package resound

// DSPChannel represents an audio channel that can have various effects applied to it.
// Any Players that have a DSPChannel set will take on the effects applied to the channel as well.
type DSPChannel struct {
	Active      bool
	Effects     map[any]IEffect
	EffectOrder []IEffect
	closed      bool

	playingPlayers []*Player
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

func (d *DSPChannel) addPlayerToList(p *Player) {
	p.dspChannel.playingPlayers = append(p.dspChannel.playingPlayers, p)
}

func (d *DSPChannel) clean() {

	for i := len(d.playingPlayers) - 1; i >= 0; i-- {
		if !d.playingPlayers[i].IsPlaying() {
			d.playingPlayers[i] = nil
			d.playingPlayers = append(d.playingPlayers[:i], d.playingPlayers[i+1:]...)
			return
		}
	}

}

// PlayingPlayers returns a copy of the list of all Players currently playing through the DSPChannel.
func (d *DSPChannel) PlayingPlayers() []*Player {
	out := []*Player{}
	copy(out, d.playingPlayers)
	return out
}

// PlayerByID returns a specific Player by its ID.
func (d *DSPChannel) PlayerByID(id any) *Player {
	for _, p := range d.playingPlayers {
		if p.id == id {
			return p
		}
	}
	return nil
}

// IsPlayingPlayer returns if a Player with the specified ID is currently playing back.
func (d *DSPChannel) IsPlayingPlayer(id any) bool {
	d.clean()
	for _, player := range d.playingPlayers {
		if player.IsPlaying() && player.id == id {
			return true
		}
	}
	return false
}
