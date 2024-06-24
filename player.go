package resound

import (
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// Player handles playback of audio and effects.
// Player embeds audio.Player and so has all of the functions and abilities of the default audio.Player
// while also applying effects either played from its source, through the Player's Effects, or through the
// Player's DSPChannel.
type Player struct {
	*audio.Player
	DSPChannel *DSPChannel
	Source     io.ReadSeeker

	EffectOrder []IEffect
	Effects     map[any]IEffect
}

// NewPlayer creates a new Player to playback an io.ReadSeeker-fulfilling audio stream.
func NewPlayer(sourceStream io.ReadSeeker) (*Player, error) {

	cp := &Player{
		Source:  sourceStream,
		Effects: map[any]IEffect{},
	}

	player, err := audio.CurrentContext().NewPlayer(cp)

	if err != nil {
		return nil, err
	}

	cp.Player = player

	return cp, nil

}

// NewPlayerFromPlayer creates a new resound.Player from an existing *audio.Player.
func NewPlayerFromPlayer(player *audio.Player) *Player {

	cp := &Player{
		Player:  player,
		Effects: map[any]IEffect{},
	}

	return cp

}

// AddEffect adds the specified Effect to the Player, with the given ID.
func (p *Player) AddEffect(id any, effect IEffect) *Player {
	p.Effects[id] = effect
	p.EffectOrder = append(p.EffectOrder, effect)
	return p
}

// Effect returns the effect associated with the given id.
// If an effect with the provided ID doesn't exist, this function will return nil.
func (p *Player) Effect(id any) IEffect {
	return p.Effects[id]
}

// SetDSPChannel sets the DSPChannel to be used for playing audio back through the Player.
func (p *Player) SetDSPChannel(c *DSPChannel) *Player {
	p.DSPChannel = c
	return p
}

// CopyProperties copies the properties (effects, current DSP Channel, etc) from one resound.Player to the other.
// Note that this won't duplicate the current state of playback of the internal audio stream.
func (p *Player) CopyProperties(other *Player) *Player {

	for k, v := range p.Effects {
		other.Effects[k] = v
	}
	other.EffectOrder = append(other.EffectOrder, p.EffectOrder...)

	other.DSPChannel = p.DSPChannel

	return p

}

func (p *Player) Read(bytes []byte) (n int, err error) {

	if p.DSPChannel != nil {

		if !p.DSPChannel.Active {
			return
		} else if p.DSPChannel.closed {
			p.Close() // Close player if the DSPChannel it's playing on is also closed
			p.Source = nil
			return 0, io.EOF
		}

	}

	if n, err = p.Source.Read(bytes); err != nil {
		return
	}

	for _, effect := range p.EffectOrder {
		effect.ApplyEffect(bytes, n)
	}

	if p.DSPChannel != nil {
		for _, effect := range p.DSPChannel.EffectOrder {
			effect.ApplyEffect(bytes, n)
		}
	}

	return

}

func (p *Player) Seek(offset int64, whence int) (int64, error) {

	if p.Source == nil {
		return 0, nil
	}

	return p.Source.Seek(offset, whence)

}
