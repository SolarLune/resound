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
	Audio *resound.Player
	Time  float64
}

//go:embed song.ogg
var songData []byte

const sampleRate = 44100

func NewGame() *Game {

	// Create a new audio context.
	audio.NewContext(sampleRate)

	// Create a reader for the audio stream.
	reader := bytes.NewReader(songData)

	// Decode the audio stream.
	stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

	if err != nil {
		panic(err)
	}

	// Create a loop from it.
	loop := audio.NewInfiniteLoop(stream, stream.Length())

	game := &Game{}

	// Here, we'll create a new resound.Player to play our audio.
	player, err := resound.NewPlayer(loop)
	if err != nil {
		panic(err)
	}

	game.Audio = player

	// Change the buffer size of audio so that we can have some responsiveness
	// when we change effect parameters on the fly; if we leave this
	// set to the default (which is like 200 milliseconds or something like that),
	// then changing effect parameters will feel laggy.
	game.Audio.SetBufferSize(time.Millisecond * 50)

	// We will also create a new effect for it - the delay.
	game.Audio.AddEffect("delay", effects.NewDelay().SetStrength(0.75).SetWait(0.1).SetFeedback(0.5))
	// The Volume effect will be used for fading.
	game.Audio.AddEffect("volume", effects.NewVolume())

	// Effects fulfill io.Reader, so you can use them to create a Player with, and set their source using IEffect.SetSource(),
	// but it's easier to apply the effect directly to a resound.Player using Player.AddEffect().

	// Finally, play the sound.
	player.Play()

	return game
}

func (game *Game) Update() error {

	var err error

	delay := game.Audio.Effect("delay").(*effects.Delay)
	volume := game.Audio.Effect("volume").(*effects.Volume)

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		delay.SetActive(!delay.Active())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		volume.StartFade(1, 0, 2)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		volume.StartFade(0, 1, 2)
	}

	game.Time += 1.0 / 60.0

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		err = ebiten.Termination
	}

	return err
}

func (game *Game) Draw(screen *ebiten.Image) {

	delay := game.Audio.Effects["delay"].(*effects.Delay)

	delayOn := "On"

	if !delay.Active() {
		delayOn = "Off"
	}

	text.Draw(screen,

		fmt.Sprintf(`This is a simple example showing how
an effect (the Delay effect) is applied
to a sound stream.
Press the Space key to toggle
the delay effect.
Press the A key to fade the song out,
and the S key to fade the song in.

Delay: %s`, delayOn),

		basicfont.Face7x13, 16, 16, color.White)

}

func (game *Game) Layout(w, h int) (int, int) { return 320, 240 }

func main() {

	ebiten.SetWindowTitle("Resound Demo - Simple")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}

}
