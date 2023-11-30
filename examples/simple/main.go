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
	"github.com/solarlune/resound/effects"
	"golang.org/x/image/font/basicfont"
)

type Game struct {
	Delay *effects.Delay
	Time  float64
}

//go:embed song.ogg
var songData []byte

const sampleRate = 44100

func NewGame() *Game {

	// Create a new audio context.
	context := audio.NewContext(sampleRate)

	// Create a reader for the audio stream.
	reader := bytes.NewReader(songData)

	// Decode the audio stream.
	stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

	if err != nil {
		panic(err)
	}

	// Create a loop from it.
	loop := audio.NewInfiniteLoop(stream, stream.Length())

	// Create a delay effect for the audio stream.
	game := &Game{
		Delay: effects.NewDelay(loop).SetStrength(0.75),
	}

	// Create a player to play the sound with the effect.
	player, err := context.NewPlayer(game.Delay)

	if err != nil {
		panic(err)
	}

	// Change the buffer size so that we can have some responsiveness
	// when we change effect parameters on the fly; if we leave this
	// default (which is like 200 milliseconds or something like that),
	// then changing effect parameters will seem laggy.
	player.SetBufferSize(time.Millisecond * 50)

	// Finally, play the sound.
	player.Play()

	return game
}

func (game *Game) Update() error {

	var err error

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		game.Delay.SetActive(!game.Delay.Active())
	}

	game.Time += 1.0 / 60.0

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		err = errors.New("quit")
	}

	return err
}

func (game *Game) Draw(screen *ebiten.Image) {

	delayOn := "On"

	if !game.Delay.Active() {
		delayOn = "Off"
	}

	text.Draw(screen, "This is a simple example showing how\nan effect (the Delay effect) is applied\nto a sound stream.\nPress the Space key to toggle\nthe delay effect.\n\nDelay: "+delayOn, basicfont.Face7x13, 16, 16, color.White)

}

func (game *Game) Layout(w, h int) (int, int) { return 320, 240 }

func main() {

	ebiten.SetWindowTitle("Resound Demo - Simple")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}

}
