# Resound 🔉

Resound is a library for applying sound effects when playing sounds back with Ebitengine. Resound was made primarily for game development, as you might expect.

## Why did you make this?

C'mon man, you already know what it is

The general advantages of using Resound is two-fold. Firstly, it allows you to easily add non-standard effects (like low-pass filtering, distortion, or panning) to sound or music playback. Secondly, it allows you to easily apply these effects across multiple groups of sounds, like a DSP. The general idea of using buses / channels is, again, taken from how [Godot](https://godotengine.org/) does things, along with other DAWs and music creation tools, like Renoise, Reason, and SunVox.

## How do I use it?

There's a couple of different ways.

1) Create effects and play an audio stream through them. The effects themselves satisfy `io.ReadSeeker`, like an ordinary audio stream from Ebitengine, so you can chain them together.

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

    // But here, we'll create a Delay effect and apply it.
    delay := effects.NewDelay(loop).SetWait(0.1).SetStrength(0.2)

    // Effects in Resound wrap streams (including other effects), so you can just use them
    // like you would an ordinary audio stream in Ebitengine.

    // You can also easily chain effects by using resound.ChainEffects().

    // Now we create a new player of the loop + delay:
	player, err := context.NewPlayer(delay)

	if err != nil {
		panic(err)
	}

    // Play it, and you're good to go.
	player.Play()

}

```

2) Apply effects to a DSP Channel, and then play sounds through there. This allows you to automatically play sounds back using various shared properties (a shared volume, shared panning, shared filter, etc).

```go

// Let's, again, assume our sound is read in or embedded as a series of bytes.
var soundBytes []byte

// Here, though, we'll be creating a DSPChannel.
var dsp *resound.DSPChannel

const sampleRate = 44100

func main() {

    // So first, we'll create an audio context, decode some bytes into a stream,
    // create a loop, etc. 
    audio.NewContext(sampleRate)

    reader := bytes.NewReader(soundBytes)

    stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

    if err != nil {
	    panic(err)
    }

    loop := audio.NewInfiniteLoop(stream, stream.Length())

    // But here, we create a DSPChannel. A DSPChannel represents a group of effects
    // that sound streams play through. When playing a stream through a DSPChannel,
    // the stream takes on the effects applied to the DSPChannel. We don't have to
    // pass a stream to effects when used with a DSPChannel, because every stream
    // played through the channel takes the effect.
    dsp = resound.NewDSPChannel()
    dsp.AddEffect("delay", effects.NewDelay(nil).SetWait(0.1).SetStrength(0.25))
    dsp.AddEffect("distort", effects.NewDistort(nil).SetStrength(0.25))
    dsp.AddEffect("volume", effects.NewVolume(nil).SetStrength(0.25))

    // Now we create a new player from the DSP channel. This will return a
    // *resound.ChannelPlayback object, which works similarly to an audio.Player
    // (in fact, it embeds the *audio.Player).
    player := dsp.CreatePlayer(loop)

    // Play it, and you're good to go, again - this time, it will run its playback
    // through the effect stack in the DSPChannel, in this case Delay > Distort > Volume.
	player.Play()

}

```

## To-do

- [ ] Global Stop - Tracking playing sounds to globally stop all sounds that are playing back
- [ ] DSPChannel Stop - ^, but for a DSP channel
- [x] Volume normalization - done through the AudioProperties struct.
- [ ] Replace all usage of "strength" with "wet/dry".

### Effects

- [X] Volume
- [X] Pan
- [X] Delay
- [X] Distortion
- [X] Low-pass Filter
- [X] Bitcrush (?)
- [ ] High-pass Filter
- [ ] Reverb
- [ ] Mix / Fade (between two streams, or between a stream and silence, and over a customizeable time)
- [ ] Loop (like, looping a signal after so much time has passed or the signal ends)
- [ ] Pitch shifting (playback using effectively a lower / higher sample rate?)
- [ ] 3D Sound (quick and easy panning and volume adjustment based on distance from listener to source)

### Generators

- [ ] Silence
- [ ] Static

  ... And whatever else may be necessary.

# Known Issues

- Currently, effects directly apply on top of streams, which means that any effects that could make streams longer (like reverbs or delays) will get cut off if the source stream ends.