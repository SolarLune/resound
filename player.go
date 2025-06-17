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
	dspChannel *DSPChannel
	Source     io.ReadSeeker
	id         any

	effectOrder []IEffect
	effects     map[any]IEffect
}

// NewPlayer creates a new Player with a customizeable ID to playback an io.ReadSeeker-fulfilling audio stream.
func NewPlayer(id any, sourceStream io.ReadSeeker) (*Player, error) {

	cp := &Player{
		id:      id,
		Source:  sourceStream,
		effects: map[any]IEffect{},
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
		effects: map[any]IEffect{},
	}

	return cp

}

// ID returns the ID associated with the given Player.
func (p *Player) ID() any {
	return p.id
}

// AddEffect adds the specified Effect to the Player, with the given ID.
func (p *Player) AddEffect(id any, effect IEffect) *Player {
	p.effects[id] = effect
	p.effectOrder = append(p.effectOrder, effect)
	return p
}

// Effect returns the effect associated with the given id.
// If an effect with the provided ID doesn't exist, this function will return nil.
func (p *Player) Effect(id any) IEffect {
	return p.effects[id]
}

// SetDSPChannel sets the DSPChannel to be used for playing audio back through the Player.
func (p *Player) SetDSPChannel(c *DSPChannel) *Player {
	p.dspChannel = c
	return p
}

// DSPChannel returns the current DSP channel associated with this Player.
func (p *Player) DSPChannel() *DSPChannel {
	return p.dspChannel
}

// CopyProperties copies the properties (effects, current DSP Channel, etc) from one resound.Player to the other.
// Note that this won't duplicate the current state of playback of the internal audio stream.
func (p *Player) CopyProperties(other *Player) *Player {

	for k, v := range p.effects {
		other.effects[k] = v
	}
	other.effectOrder = append(other.effectOrder, p.effectOrder...)

	other.dspChannel = p.dspChannel

	return p

}

func (p *Player) Read(bytes []byte) (n int, err error) {

	if p.dspChannel != nil {

		if !p.dspChannel.Active {
			return
		} else if p.dspChannel.closed {
			p.Close() // Close player if the DSPChannel it's playing on is also closed
			p.Source = nil
			return 0, io.EOF
		}

	}

	if n, err = p.Source.Read(bytes); err != nil && err != io.EOF {
		return
	}

	for _, effect := range p.effectOrder {
		effect.ApplyEffect(bytes, n)
	}

	if p.dspChannel != nil {
		for _, effect := range p.dspChannel.effectOrder {
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

func (p *Player) Play() {
	if p.dspChannel != nil {
		p.dspChannel.clean()
		p.dspChannel.addPlayerToList(p)
	}
	p.Player.Play()
}
