package main

import (
	"bytes"
	"errors"
	"image/color"

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

// Embed the sound samples

//go:embed square_wave_5.ogg
var square5Data []byte

//go:embed square_wave_25.ogg
var square25Data []byte

//go:embed square_wave_100.ogg
var square100Data []byte

//go:embed song_quiet.ogg
var songData []byte

// Set the sample rate as a globally-available constant

const sampleRate = 44100

type Game struct {
	Normalize       bool
	AudioProperties resound.AudioProperties
	PlayingSound    *audio.Player
}

func NewGame() *Game {

	game := &Game{
		AudioProperties: resound.NewAudioProperties(),
	}

	audio.NewContext(sampleRate)

	return game
}

// Here we play back audio streams. We pass a name too in order to store audio playback stream analysis information.
func (game *Game) Play(name string, sample []byte) {

	// Stop an existing sound if it's playing back.
	if game.PlayingSound != nil && game.PlayingSound.IsPlaying() {
		game.PlayingSound.Pause()
		game.PlayingSound.Close()
	}

	audioContext := audio.CurrentContext()

	reader := bytes.NewReader(sample)

	stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

	if err != nil {
		panic(err)
	}

	volume := effects.NewVolume(stream)

	if game.Normalize {
		// Here we analyze the stream, using a chunk size for scanning the audio file.
		// The longer the file and the more variance in the file, the higher the fidelity should be.
		// If the stream has been analyzed already, then this will simply return the results.
		prop, err := game.AudioProperties.Get(name).Analyze(stream, 16)
		if err != nil {
			panic(err)
		}
		volume.SetNormalizationFactor(prop.Normalization)
	}

	player, err := audioContext.NewPlayer(volume)
	if err != nil {
		panic(err)
	}
	player.Play()
	game.PlayingSound = player

}

func (game *Game) Update() error {

	var returnCode error

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		game.Normalize = !game.Normalize
	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		game.Play("square5", square5Data)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		game.Play("square25", square25Data)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		game.Play("square100", square100Data)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		game.Play("encouragement", songData)
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		returnCode = errors.New("quit")
	}

	return returnCode
}

func (game *Game) Draw(screen *ebiten.Image) {

	normalizationOn := "off"
	if game.Normalize {
		normalizationOn = "on"
	}

	text.Draw(screen, "This is a simple example showing how\nvolume normalization works.\nPress 1, 2, 3, and 4 to play a sample\nwith varying volumes, and press the Space\nkey to toggle normalization before\nplayback.\n\nNormalization is currently: "+normalizationOn, basicfont.Face7x13, 16, 16, color.White)

}

func (game *Game) Layout(w, h int) (int, int) { return 320, 240 }

func main() {

	ebiten.SetWindowTitle("Resound Demo - Normalization")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}

}
