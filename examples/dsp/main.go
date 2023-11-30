package main

import (
	"bytes"
	"errors"
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

func NewGame() *Game {

	// We create a new audio context using the provided sample rate.
	audio.NewContext(sampleRate)

	// We then create a new DSP Channel. It will use the global audio context we just made.
	game := &Game{
		DSP: resound.NewDSPChannel(),
	}

	// Now we add effects; we don't have to specify a stream because a DSPChannel applies them
	// to all streams played through the channel.

	game.DSP.Add("delay", effects.NewDelay(nil).SetWait(0.1).SetStrength(0.9))
	game.DSP.Add("pan", effects.NewPan(nil))
	game.DSP.Add("volume", effects.NewVolume(nil))

	reader := bytes.NewReader(songData)

	stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

	if err != nil {
		panic(err)
	}

	loop := audio.NewInfiniteLoop(stream, stream.Length())

	// I want to make the music quieter, so I'll actually add a volume
	// effect in the middle of this

	volume := effects.NewVolume(loop).SetStrength(0.6)

	player := game.DSP.CreatePlayer(volume)
	player.SetBufferSize(time.Millisecond * 50)
	player.Play()

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

		game.DSP.CreatePlayer(stream).Play()

	}

	game.Time += 1.0 / 60.0

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		returnCode = errors.New("quit")
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
it shares the properties as the music.

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
