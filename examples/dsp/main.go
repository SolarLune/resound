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
	"golang.org/x/image/font/basicfont"
)

type Game struct {
	DSP  *resound.DSPChannel
	Time float64
}

//go:embed encouragement.ogg
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

	game.DSP.Add("delay", resound.NewDelay(nil, 0.1, 0.9, false))
	game.DSP.Add("pan", resound.NewPan(nil, 0))
	game.DSP.Add("volume", resound.NewVolume(nil, 1))

	reader := bytes.NewReader(songData)

	stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

	if err != nil {
		panic(err)
	}

	loop := audio.NewInfiniteLoop(stream, stream.Length())

	// I want to make the music quieter, so I'll actually add a volume
	// effect in the middle of this

	volume := resound.NewVolume(loop, 0.25)

	player := game.DSP.CreatePlayer(volume)
	player.SetBufferSize(time.Millisecond * 50)
	player.Play()

	return game
}

func (game *Game) Update() error {

	var returnCode error

	pan := game.DSP.Effects["pan"].(*resound.Pan)
	volume := game.DSP.Effects["volume"].(*resound.Volume)

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		pan.Percent += 0.01
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		pan.Percent -= 0.01
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		volume.Percent += 0.01
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		volume.Percent -= 0.01
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {

		reader := bytes.NewReader(stepData)
		stream, err := wav.DecodeWithSampleRate(sampleRate, reader)
		if err != nil {
			panic(err)
		}

		game.DSP.CreatePlayer(stream).Play()

	}

	if pan.Percent > 1 {
		pan.Percent = 1
	}

	if pan.Percent < -1 {
		pan.Percent = -1
	}

	// game.Pan.Percent += 0.01
	// if game.Pan.Percent > 1 {
	// 	game.Pan.Percent = -1
	// }

	// fmt.Println(game.Pan.Percent)

	game.Time += 1.0 / 60.0

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		returnCode = errors.New("quit")
	}

	return returnCode
}

func (game *Game) Draw(screen *ebiten.Image) {

	pan := game.DSP.Effects["pan"].(*resound.Pan)
	volume := game.DSP.Effects["volume"].(*resound.Volume)
	text.Draw(screen, fmt.Sprintf("This is an example showing how\nDSPChannels work. You create a\nDSPChannel, add effects, and play streams\nthrough it to share the effects.\n\nIn this example, left and right arrow keys\nalter the pan. Up and down alters the\nvolume. Press space to play a footstep\nsound through the channel.\n\nPan level: %.2f\nVolume level: %.2f", pan.Percent, volume.Percent), basicfont.Face7x13, 16, 16, color.White)

}

func (game *Game) Layout(w, h int) (int, int) { return 320, 240 }

func main() {

	ebiten.SetWindowTitle("Resound Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}

}
