package main

import (
	"bytes"
	"errors"
	"image/color"
	"time"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/solarlune/resound"
	"golang.org/x/image/font/basicfont"
)

type Game struct {
	Time float64
}

//go:embed encouragement.ogg
var songData []byte

const sampleRate = 44100

func NewGame() *Game {

	context := audio.NewContext(sampleRate)

	game := &Game{}

	reader := bytes.NewReader(songData)

	stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

	if err != nil {
		panic(err)
	}

	loop := audio.NewInfiniteLoop(stream, stream.Length())

	// Here, we want to chain effects together, so this utility function can be used
	// to make it simpler, as otherwise, it's more difficult to reorder effects.

	sfx := resound.ChainEffects(
		resound.NewDelay(loop).SetWait(0.15).SetStrength(0.75).SetFeedbackLoop(true),
		resound.NewPan(nil).SetPan(0.75),
		resound.NewLowpassFilter(nil).SetStrength(0.9),
	)

	player, err := context.NewPlayer(sfx)

	// Change the buffer size so that we can have some responsiveness
	// when we change effect parameters on the fly
	player.SetBufferSize(time.Millisecond * 50)

	if err != nil {
		panic(err)
	}

	player.Play()

	return game
}

func (game *Game) Update() error {

	var returnCode error

	game.Time += 1.0 / 60.0

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		returnCode = errors.New("quit")
	}

	return returnCode
}

func (game *Game) Draw(screen *ebiten.Image) {

	text.Draw(screen, "This is another simple example showing how\neffects can be chained. In this example,\na delay effect is chained into\na pan filter, which is chained into\na lowpass filter.", basicfont.Face7x13, 16, 16, color.White)

}

func (game *Game) Layout(w, h int) (int, int) { return 320, 240 }

func main() {

	ebiten.SetWindowTitle("Resound Demo - Simple")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}

}
