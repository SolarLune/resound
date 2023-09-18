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
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/solarlune/resound"
	"golang.org/x/image/font/basicfont"
)

type Game struct {
	Delay *resound.Delay
	Time  float64
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

	game.Delay = resound.NewDelay(loop).SetStrength(0.75)

	player, err := context.NewPlayer(game.Delay)

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

	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		game.Delay.SetActive(!game.Delay.Active())
	}

	game.Time += 1.0 / 60.0

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		returnCode = errors.New("quit")
	}

	return returnCode
}

func (game *Game) Draw(screen *ebiten.Image) {

	delayOn := "On"

	if !game.Delay.Active() {
		delayOn = "Off"
	}

	text.Draw(screen, "This is a simple example showing how\nan effect (the Delay effect) is applied\nto a sound stream.\nPress A to toggle the delay effect.\n\nDelay: "+delayOn, basicfont.Face7x13, 16, 16, color.White)

}

func (game *Game) Layout(w, h int) (int, int) { return 320, 240 }

func main() {

	ebiten.SetWindowTitle("Resound Demo - Simple")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}

}
