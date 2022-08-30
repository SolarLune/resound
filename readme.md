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

    // But here, we'll create a Delay effect and apply it. The first and second
    // arguments are how many seconds should pass between the initial sound and
    // the delay, and how loud (in percentage) the delay should be. The last argument
    // is if the delay should feedback into itself.
    delay := resound.NewDelay(loop, sampleRate, 0.1, 0.2, false)

    // Effects in Resound wrap streams (including other effects), so you can just use them
    // like you would an ordinary stream.

    // Now we create a new player of the loop + delay:
	player, err := audio.NewPlayer(context, delay)

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
    context := audio.NewContext(sampleRate)

    reader := bytes.NewReader(soundBytes)

    stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

    if err != nil {
	panic(err)
    }

    loop := audio.NewInfiniteLoop(stream, stream.Length())

    // But here, we create a DSPChannel. A DSPChannel represents a group of effects
    // that sound streams play through. When playing a stream through a DSPChannel,
    // the stream takes on the effects applied to the DSPChannel.
    dsp = resound.NewDSPChannel(context)
    dsp.AddEffect("delay", NewDelay(nil, 0.1, 0.25))
    dsp.AddEffect("distort", NewDistort(nil, 0.25))
    dsp.AddEffect("volume", NewVolume(0.25))

    // Now we create a new player from the DSP channel. This will return a
    // *resound.ChannelPlayback object, which works similarly to an audio.Player
    // (in fact, it embeds the *audio.Player).
    player := dsp.CreatePlayer(loop)

    // Play it, and you're good to go, again - this time, it will run its playback
    // through the effect stack in the DSPChannel, in this case Delay > Distort > Volume.
	player.Play()

}

```

## What effects are implemented?

- [X] Volume
- [X] Pan
- [X] Delay
- [X] Distortion
- [X] Low-pass Filter
- [ ] High-pass Filter
- [ ] Reverb
- [ ] Mix
- [ ] Fade
- [ ] Loop

  ... And whatever else may be necessary.

# Known Issues

- Currently, effects directly apply on top of streams, which means that any effects that could make streams longer (like reverbs or delays) will get cut off if the stream ends.