package play

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/hajimehoshi/oto"
	"github.com/pkg/errors"
)

// Player plays audio streams.
type Player struct {
	isPlaying bool
	done      chan struct{}
	play      chan struct{}
	pause     chan struct{}

	getsamples chan struct{}
	samples    chan [][2]float64

	log *log.Logger
}

// NewPlayer returns a Player ready to play Audio streams.
func NewPlayer(l *log.Logger) *Player {
	if l == nil {
		l = log.New(os.Stdout, "[player] ", log.LstdFlags)
	}
	return &Player{
		done:       make(chan struct{}),
		play:       make(chan struct{}),
		pause:      make(chan struct{}),
		getsamples: make(chan struct{}),
		samples:    make(chan [][2]float64),
		log:        l,
	}
}

// Play the audio until the audio is finished or Done is called.
// Play is synchronous, returning when the audio is finished.
func (p *Player) Play(a Audio) error {
	streamer, format, err := p.decode(a)
	if err != nil {
		return errors.Wrap(err, "decoding")
	}
	defer streamer.Close()
	mixer := beep.Mixer{}
	mixer.Play(beep.Seq(streamer, beep.Callback(func() {
		p.Done()
	})))
	bufferSize := format.SampleRate.N(time.Millisecond*64) * 4
	speaker, err := oto.NewPlayer(int(format.SampleRate), 2, 2, bufferSize)
	if err != nil {
		return errors.Wrap(err, "initialising speaker")
	}
	p.loop(
		mixer,
		speaker,
		make([][2]float64, bufferSize),
		make([]byte, bufferSize*4),
	)
	return nil
}

// Done stops the current playback without waiting for the audio to finish.
func (p *Player) Done() {
	p.done <- struct{}{}
}

// Resume the playback from a paused state.
func (p *Player) Resume() {
	p.play <- struct{}{}
}

// Pause the playback from a playing state.
func (p *Player) Pause() {
	p.pause <- struct{}{}
}

// IsPlaying returns the state of the Player.
// Note that IsPlaying == false could mean either the audio has finished or is
// simply paused.
func (p *Player) IsPlaying() bool {
	return p.isPlaying
}

// GetSamples returns a snapshot of the samples.
func (p *Player) GetSamples() [][2]float64 {
	p.getsamples <- struct{}{}
	return <-p.samples
}

func (p *Player) decode(a Audio) (beep.StreamSeekCloser, beep.Format, error) {
	if a.Format != MP3 {
		return nil, beep.Format{}, fmt.Errorf("format not supported: %s", a.Format)
	}
	return mp3.Decode(a)
}

// loop writes audio data to the speaker.
func (p *Player) loop(
	mixer beep.Mixer,
	speaker io.Writer,
	samples [][2]float64,
	buffer []byte,
) {
	defer func() {
		p.isPlaying = false
	}()
	p.isPlaying = true
	for {
		select {
		case <-p.pause:
			p.isPlaying = false
			select {
			case <-p.play:
				p.isPlaying = true
			case <-p.done:
				return
			}
		case <-p.done:
			return
		case <-p.getsamples:
			buffer := make([][2]float64, len(samples))
			copy(buffer, samples)
			p.samples <- buffer
		default:
			mixer.Stream(samples)
			for ii := range samples {
				for c := range samples[ii] {
					val := samples[ii][c]
					if val < -1 {
						val = -1
					}
					if val > +1 {
						val = +1
					}
					valInt16 := int16(val * (1<<15 - 1))
					low := byte(valInt16)
					high := byte(valInt16 >> 8)
					buffer[ii*4+c*2+0] = low
					buffer[ii*4+c*2+1] = high
				}
			}
			if _, err := speaker.Write(buffer); err != nil {
				p.log.Printf("writing to speaker: %v", err)
			}
		}
	}
}

// Audio is a stream of bytes containing encoded audio.
type Audio struct {
	io.ReadCloser
	Format Format
}

// Format is a string that signals the audio encoding (mp3, etc).
type Format int

const (
	MP3 Format = iota
	FLAC
	WAV
)

func (f Format) String() string {
	switch f {
	case MP3:
		return "mp3"
	case FLAC:
		return "flac"
	case WAV:
		return "wav"
	default:
		return "unknown"
	}
}
