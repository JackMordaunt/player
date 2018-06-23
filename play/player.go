package play

import (
	"log"
	"os"
	"sync"

	"github.com/faiface/beep"
	"github.com/hajimehoshi/oto"
	"github.com/pkg/errors"
)

// PlayerOld can audio streams.
type PlayerOld struct {
	*sync.Mutex
	mixer     beep.Mixer
	samples   [][2]float64
	buf       []byte
	player    *oto.Player
	underrun  func()
	done      chan struct{}
	isPlaying bool
	log       *log.Logger
}

// NewPlayer initialises a Player.
func NewPlayer(sampleRate beep.SampleRate, bufferSize int) (*PlayerOld, error) {
	speaker, err := oto.NewPlayer(int(sampleRate), 2, 2, bufferSize*4)
	if err != nil {
		return nil, errors.Wrap(err, "initialising speaker")
	}
	player := &PlayerOld{
		Mutex:    &sync.Mutex{},
		mixer:    beep.Mixer{},
		samples:  make([][2]float64, bufferSize),
		buf:      make([]byte, bufferSize*4),
		player:   speaker,
		done:     make(chan struct{}),
		underrun: func() {},
		log:      log.New(os.Stdout, "[player]", log.LstdFlags),
	}
	go player.run()
	return player, nil
}

// Play starts playing the provided streamers.
func (p *PlayerOld) Play(s beep.Streamer) {
	p.Lock()
	p.isPlaying = true
	p.mixer.Play(beep.Seq(s, beep.Callback(func() {
		p.Close()
	})))
	p.Unlock()
	p.wait()
}

// Stop playing by emptying the mixer.
func (p *PlayerOld) Stop() {
	p.Lock()
	p.mixer = beep.Mixer{}
	p.isPlaying = false
	p.Unlock()
}

// Close the player.
func (p *PlayerOld) Close() {
	close(p.done)
	p.player.Close()
}

func (p *PlayerOld) wait() {
	<-p.done
}

// IsPlaying reports whether the player is playing or not.
func (p *PlayerOld) IsPlaying() bool {
	return p.isPlaying
}

// TogglePause toggles playback.
func (p *PlayerOld) TogglePause() {
	if p.isPlaying {
		p.Lock()
	} else {
		p.Unlock()
	}
	p.isPlaying = !p.isPlaying
}

// Done signals that we are finished with the currently playing song.
func (p *PlayerOld) Done() {
	p.done <- struct{}{}
}

// run the player.
func (p *PlayerOld) run() {
	for {
		select {
		default:
			if err := p.update(); err != nil {
				p.log.Printf("update: %v", err)
			}
		case <-p.done:
			return
		}
	}
}

func (p *PlayerOld) update() error {
	p.mixer.Stream(p.samples)
	for ii := range p.samples {
		for c := range p.samples[ii] {
			val := p.samples[ii][c]
			if val < -1 {
				val = -1
			}
			if val > +1 {
				val = +1
			}
			valInt16 := int16(val * (1<<15 - 1))
			low := byte(valInt16)
			high := byte(valInt16 >> 8)
			p.buf[ii*4+c*2+0] = low
			p.buf[ii*4+c*2+1] = high
		}
	}
	if _, err := p.player.Write(p.buf); err != nil {
		return errors.Wrap(err, "writing to speaker")
	}
	return nil
}
