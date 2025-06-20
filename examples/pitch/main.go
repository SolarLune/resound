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
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/solarlune/resound"
	"github.com/solarlune/resound/effects"
	"golang.org/x/image/font/basicfont"
)

type Game struct {
	PitchShift *effects.PitchShift
	Time       float64
}

//go:embed song.ogg
var songData []byte

const sampleRate = 44100

func NewGame() *Game {

	// Create the audio context
	audio.NewContext(sampleRate)

	// Load the song data
	reader := bytes.NewReader(songData)

	// Decode
	stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

	if err != nil {
		panic(err)
	}

	// Create a loop
	loop := audio.NewInfiniteLoop(stream, stream.Length())

	// Create a new resound.Player, which is an enhanced audio player similar to
	// Ebitengine's but you can add effects directly on the player or use it with
	// a DSP channel (which can have effects on it).
	player, err := resound.NewPlayer("bgm", loop)

	// Create a pitch shift effect with the given pitch buffer size.
	game := &Game{
		PitchShift: effects.NewPitchShift(1024).SetSource(loop).SetPitch(0.8),
	}
	player.AddEffect("pitch shift", game.PitchShift)

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
		game.PitchShift.SetActive(!game.PitchShift.Active())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		game.PitchShift.SetPitch(game.PitchShift.Pitch() + 0.02)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		game.PitchShift.SetPitch(game.PitchShift.Pitch() - 0.02)
	}

	game.Time += 1.0 / 60.0

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		err = ebiten.Termination
	}

	return err
}

func (game *Game) Draw(screen *ebiten.Image) {

	text.Draw(screen, fmt.Sprintf(`This example shows how the PitchShift
effect works. Press the Space key
to toggle the pitch effect.
The up and down keys pitch
the song up or down.

Pitch On: %t
Pitch Strength:%03f`, game.PitchShift.Active(), game.PitchShift.Pitch()), basicfont.Face7x13, 16, 16, color.White)

}

func (game *Game) Layout(w, h int) (int, int) { return 320, 240 }

func main() {

	ebiten.SetWindowTitle("Resound Demo - Simple")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}

}
