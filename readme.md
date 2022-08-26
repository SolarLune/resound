# Resound 🔉

Resound is a library for applying sound effects when playing sounds back with Ebitengine. Resound was made primarily for game development, as you might expect.

## Why did you make this?

C'mon man, you already know what it is

## How do I use it?

There's a couple of different ways.

1) Create effects and play the stream through them. The effects themselves satisfy `io.ReadSeeker`, like an ordinary audio stream from Ebitengine, so you can chain them together.

```go

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
    delay := resound.NewDelay(loop, sampleRate, 0.1, 0.2)

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

var soundBytes []byte

var dsp *resound.DSPChannel

const sampleRate = 44100

func main() {

    // So first, we'll create an audio context, decode some bytes into a stream,
    // create a loop, etc. 
    context := audio.NewContext(sampleRate)

    dsp = resound.NewDSPChannel(context)
    dsp.AddEffect("delay", NewDelay(nil, 0.1, 0.25))
    dsp.AddEffect("distort", NewDistort(nil, 0.25))
    dsp.AddEffect("volume", NewVolume(0.25))

    reader := bytes.NewReader(soundBytes)

    stream, err := vorbis.DecodeWithSampleRate(sampleRate, reader)

    if err != nil {
	panic(err)
    }

    loop := audio.NewInfiniteLoop(stream, stream.Length())

    // Now we create a new player from the DSP channel. This will return a
    // *resound.ChannelPlayback object, which works similarly to an audio.Player.
    player := dsp.CreatePlayer(loop)

    // Play it, and you're good to go, again - this time, it will run its playback
    // through the effect stack in the DSPChannel, in this case Delay > Distort > Volume.
	player.Play()

}

```

## What effects are implemented?

[x] Volume
[x] Pan
[x] Delay
[x] Distortion
[x] Low-pass Filter
[ ] High-pass Filter
[ ] Reverb
[ ] Mix
[ ] Fade
[ ] Loop
[ ] Auto-stream lengthening for delays / reverbs / effects that make the stream longer (fundamentally, the way this works is by altering the input stream; if the stream ends, then the effect also ends, which may or may not be what you want.)
... And whatever else may be necessary.
