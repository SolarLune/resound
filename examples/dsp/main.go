package main

import (
	"bytes"
	"fmt"
	"image/color"
	"time"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/solarlune/resound"
	"github.com/solarlune/resound/effects"
	"golang.org/x/image/font/basicfont"
)

type Game struct {
	DSP  *resound.DSPChannel
	Time float64
}

//go:embed song.ogg
var songData []byte

//go:embed footstep.wav
var stepData []byte

const sampleRate = 44100

// While setting effects manually works fine, it can be a bit awkward for larger amounts of effects and dynamically altering them.
// It's easier to use DSPChannels and apply Effects directly to DSPPlayers for this.

func NewGame() *Game {

	// We create a new audio context using the provided sample rate.
	audio.NewContext(sampleRate)

	// We then create a new DSP Channel. It will use the global audio context we just made.
	game := &Game{
		DSP: resound.NewDSPChannel(),
	}

	// Now we add effects; we don't have to specify a source because a DSPChannel applies effects
	// to all streams played through the channel.

	game.DSP.AddEffect("delay", effects.NewDelay().SetWait(0.1).SetStrength(0.9))
	game.DSP.AddEffect("pan", effects.NewPan())
	game.DSP.AddEffect("volume", effects.NewVolume())

	reader := bytes.NewReader(songData)

	stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

	if err != nil {
		panic(err)
	}

	loop := audio.NewInfiniteLoop(stream, stream.Length())

	// I want to make the music quieter, so I'll actually add a volume
	// effect here - I'll apply the effect directly to the player, to make it simpler.
	player, err := resound.NewPlayer("bgm", loop)
	if err != nil {
		panic(err)
	}
	player.AddEffect("volume", effects.NewVolume().SetStrength(0.4))

	// We set the DSP channel so the sound player takes on the effects set on the DSP channe.
	player.SetDSPChannel(game.DSP)
	player.Play()
	player.SetBufferSize(time.Millisecond * 50)

	return game
}

func (game *Game) Update() error {

	var returnCode error

	pan := game.DSP.Effects["pan"].(*effects.Pan)
	volume := game.DSP.Effects["volume"].(*effects.Volume)

	panFactor := pan.Pan()

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		panFactor += 0.02
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		panFactor -= 0.02
	}

	pan.SetPan(panFactor)

	volumeStrength := volume.Strength()

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		volumeStrength += 0.02
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		volumeStrength -= 0.02
	}

	volume.SetStrength(volumeStrength)

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {

		reader := bytes.NewReader(stepData)
		stream, err := wav.DecodeWithSampleRate(sampleRate, reader)
		if err != nil {
			panic(err)
		}

		player, err := resound.NewPlayer("footstep", stream)
		if err != nil {
			panic(err)
		}
		player.SetDSPChannel(game.DSP).Play()

	}

	game.Time += 1.0 / 60.0

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		returnCode = ebiten.Termination
	}

	return returnCode
}

func (game *Game) Draw(screen *ebiten.Image) {

	pan := game.DSP.Effects["pan"].(*effects.Pan)
	volume := game.DSP.Effects["volume"].(*effects.Volume)
	text.Draw(screen, fmt.Sprintf(`This is an example showing how
DSPChannels work. You create a
DSPChannel, add effects, and play streams
through it to share the effects.

In this example, left and right arrow keys
alter the pan. Up and down alters the
volume. Press space to play a footstep
sound through the channel. Notice that
it shares the properties as the music
because they're both played on the same
DSP Channel and the effects are applied
to the channel.

Pan level: %.2f
Volume level: %.2f`, pan.Pan(), volume.Strength()), basicfont.Face7x13, 16, 16, color.White)

}

func (game *Game) Layout(w, h int) (int, int) { return 320, 240 }

func main() {

	ebiten.SetWindowTitle("Resound Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}

}
