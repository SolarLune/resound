package resound

import (
	"io"
	"math"
)

// AnalysisResult is an object that contains the results of an analysis performed on a stream.
type AnalysisResult struct {
	Normalization float64
}

// AudioProperty is an object that allows associating an AnalysisResult for a specific stream with a name for that stream.
type AudioProperty struct {
	result   AnalysisResult
	analyzed bool
}

func newAudioProperty() *AudioProperty {
	ap := &AudioProperty{}
	ap.ResetAnalyzation()
	return ap
}

// A common byte slice for analysis of various streams
var byteSlice = make([]byte, 512)

// Analyze analyzes the provided audio stream, returning an AnalysisResult object.
// The stream is the audio stream to be used for scanning, and the scanCount is the number of times
// the function should scan various parts of the audio stream. The higher the scan count, the more accurate
// the results should be, but the longer the scan would take.
// A scanCount of 16 means it samples the stream 16 times evenly throughout the file.
// If a scanCount of 0 is provided, it will default to 16.
func (ap *AudioProperty) Analyze(stream io.ReadSeeker, scanCount int64) AnalysisResult {

	if scanCount <= 0 {
		scanCount = 16
	}

	if ap.analyzed {
		return ap.result
	}

	largest := 0.0

	// Get the length of the stream by seeking to the end; we can't seek using io.SeekEnd because it has no end, apparently
	length, _ := stream.Seek(math.MaxInt64, io.SeekStart)

	// Seek back afterwards as necessary
	stream.Seek(0, io.SeekStart)

	seekJump := length / int64(scanCount)

	var err error

	pos := int64(0)

	for err == nil {

		_, err = stream.Read(byteSlice)

		if err != nil {
			break
		}

		audioBuffer := AudioBuffer(byteSlice)

		for i := 0; i < audioBuffer.Len(); i++ {

			l, r := audioBuffer.Get(i)

			la := math.Abs(l)
			ra := math.Abs(r)

			if la > largest {
				largest = la
			}
			if ra > largest {
				largest = ra
			}

		}

		// InfiniteLoops don't return an error if you attempt to seek too far; they just go back to the start when attempting to read
		if pos+seekJump >= length {
			break
		}

		pos += seekJump

		_, err := stream.Seek(seekJump, io.SeekCurrent)

		if err != nil {
			break
		}

	}

	stream.Seek(0, io.SeekStart)

	ap.result = AnalysisResult{
		Normalization: 1.0 / largest,
	}

	ap.analyzed = true

	return ap.result

}

func (ap *AudioProperty) ResetAnalyzation() {
	ap.analyzed = false
	ap.result = AnalysisResult{}
}

type AudioProperties map[string]*AudioProperty

func NewAudioProperties() AudioProperties {
	return AudioProperties{}
}

func (ap AudioProperties) Get(name string) *AudioProperty {

	if _, exists := ap[name]; !exists {
		ap[name] = newAudioProperty()
	}

	return ap[name]

}
