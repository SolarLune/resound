# Resound 🔉

Resound is a library for applying sound effects when playing sounds back with Ebitengine. Resound was made primarily for game development, as you might expect.

## Why did you make this?

C'mon man, you already know what it is

The general advantages of using Resound is two-fold. Firstly, it allows you to easily add non-standard effects (like low-pass filtering, distortion, or panning) to sound or music playback. Secondly, it allows you to easily apply these effects across multiple groups of sounds, like a DSP. The general idea of using buses / channels is, again, taken from how [Godot](https://godotengine.org/) does things, along with other DAWs and music creation tools, like Renoise, Reason, and SunVox.

## How do I use it?

There's a couple of different ways to use resound.

1. For single sounds, it's easiest to create a `resound.Player` and apply Effects directly onto the player. This is generally the advised route, as it's simpler to grasp and easier to refactor or rework as necessary.

```go

var soundBytes []byte

const sampleRate = 44100

func main() {

    // Create a new audio context - we don't need to pass it anywhere
    // because there can only be one and it's accessible globally.
    audio.NewContext(sampleRate)

    reader := bytes.NewReader(soundBytes)

    // Decode the audio stream.
    stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

    if err != nil {
        panic(err)
    }

    // Create a loop from it.
    loop := audio.NewInfiniteLoop(stream, stream.Length())

    // Now we can create a new resound.Player, which is an enhanced
    // audio player, and add a delay effect to it.
    player, err := resound.NewPlayer(loop).AddEffect(
        "delay", effects.NewDelay().SetWait(0.1).SetStrength(0.2),
    )

    player.Play()


}

```

2. For applying multiple pre-set effects to a group of sounds, you have DSP Channels. This allows you to automatically play sounds back using various shared properties (a shared volume, shared panning, shared filter, etc).

```go

// Let's, again, assume our sound is read in or embedded as a series of bytes.
var soundBytes []byte

// Here, though, we'll be creating a DSPChannel.
var dsp *resound.DSPChannel

const sampleRate = 44100

func main() {

    audio.NewContext(sampleRate)

    reader := bytes.NewReader(soundBytes)

    stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

    if err != nil {
        panic(err)
    }

    loop := audio.NewInfiniteLoop(stream, stream.Length())

    // So here, we create a DSPChannel. A DSPChannel represents a group of effects
    // that sound streams play through. When playing a stream through a DSPChannel,
    // the stream takes on the effects applied to the DSPChannel.
    dsp = resound.NewDSPChannel()
    dsp.AddEffect("delay", effects.NewDelay().SetWait(0.1).SetStrength(0.25))
    dsp.AddEffect("distort", effects.NewDistort().SetStrength(0.25))
    dsp.AddEffect("volume", effects.NewVolume().SetStrength(0.25))

    // Now we create a new player and specify the DSP channel.
    player := resound.NewPlayer("bgm", loop).SetDSPChannel(dsp)

    // Play it, and you're good to go, again - this time, it will run its playback
    // through the effect stack in the DSPChannel, in this case Delay > Distort > Volume.
    player.Play()

}

```

3. Effects themselves satisfy `io.ReadSeeker`, like an ordinary audio stream from Ebitengine, so you can finally also just supply an audio stream to them and then create a player from them using Ebiten's built-in `context.NewPlayer()` functionality:

```go

// Let's assume our sound is read in or embedded as a series of bytes.
var soundBytes []byte

const sampleRate = 44100

func main() {

    // So first, we'll create an audio context, decode some bytes into a stream,
    // create a loop, etc.
    context := audio.NewContext(sampleRate)

    reader := bytes.NewReader(soundBytes)

    stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

    if err != nil {
        panic(err)
    }

    loop := audio.NewInfiniteLoop(stream, stream.Length())

    delay := effects.NewDelay().SetWait(0.1).SetStrength(0.2)

    // Effects in Resound wrap streams (including other effects), so you could just use them
    // like you would an ordinary audio stream in Ebitengine. Just set their source and
    // play through them:

    delay.SetSource(loop)
    player, err := context.NewPlayer(delay)

    // (Note that if you're going to change effect parameters in real time, you may want to
    // lower the internal buffer size for Players using (*audio.Player).SetBufferSize())

    if err != nil {
        panic(err)
    }

    // Play it, and you're good to go.
    player.Play()


}

```

## To-do

-   [ ] Global Stop - Tracking playing sounds to globally stop all sounds that are playing back
-   [ ] DSPChannel Stop - ^, but for a DSP channel
-   [x] Volume normalization - done through the AudioProperties struct.
-   [ ] Beat / rhythm analysis?
-   [ ] Replace all usage of "strength" with "wet/dry".

### Effects

-   [x] Volume
-   [x] Pan
-   [x] Delay
-   [x] Distortion
-   [x] Low-pass Filter
-   [x] Bitcrush (?)
-   [ ] High-pass Filter
-   [ ] Reverb
-   [x] Mix / Fade (between two streams, or between a stream and silence, and over a customizeable time) - Fading is now partially implemented, but not mixing
-   [ ] Loop (like, looping a signal after so much time has passed or the signal ends)
-   [x] Pitch shifting
-   [ ] Playback speed adjustment
-   [ ] 3D Sound (quick and easy panning and volume adjustment based on distance from listener to source)

### Generators

-   [ ] Silence
-   [ ] Static

    ... And whatever else may be necessary.

# Known Issues

-   Currently, effects directly apply on top of streams, which means that any effects that could make streams longer (like reverbs or delays) will get cut off if the source stream ends.
-   All effect parameters are ordinary values (floats, bools, etc), but since they're accessible through user-facing functions as well as used within the effect's Read() function, this creates a race condition, particularly for fading a Volume effect. This should probably be properly synchronized in some way.
