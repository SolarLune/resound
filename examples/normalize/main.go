package main

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/solarlune/resound"
	"golang.org/x/image/font/basicfont"
)

// Embed the sound samples

//go:embed square_wave_5.ogg
var square5 []byte

//go:embed square_wave_25.ogg
var square25 []byte

//go:embed square_wave_100.ogg
var square100 []byte

//go:embed encouragement_quiet.ogg
var encouragement []byte

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

// Here we play back audio streams. We pass a name too in order to pull relevant information from audio playback streams.
func (game *Game) Play(name string, sample []byte) {

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

	volume := resound.NewVolume(stream)

	if game.Normalize {
		// Here we analyze the stream, using a chunk size for scanning the audio file.
		// The longer the file and the more variance in the file, the higher the fidelity should be.
		prop := game.AudioProperties.Get(name).Analyze(stream, 16)
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

	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		game.Normalize = !game.Normalize
	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		game.Play("square5", square5)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		game.Play("square25", square25)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		game.Play("square100", square100)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		game.Play("encouragement", encouragement)
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

	text.Draw(screen, "This is a simple example showing how\nvolume normalization works.\nPress 1, 2, 3, and 4 to play a sample\nwith varying volumes, and press A to\ntoggle normalization before playback.\nNormalization is currently: "+normalizationOn, basicfont.Face7x13, 16, 64, color.White)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f, FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()))

}

func (game *Game) Layout(w, h int) (int, int) { return 320, 240 }

func main() {

	ebiten.SetWindowTitle("Resound Demo - Normalization")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}

}
