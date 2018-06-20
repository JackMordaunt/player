package main

import (
	"github.com/ktye/fft"
	"github.com/murlokswarm/app"
	//"log"
	"math"
	"math/cmplx"
	"time"

	"github.com/grocid/mistlur/play"
)

func init() {
	app.Import(&Player{})
}

// Player is the component disPlaying Player.
type Player struct {
	Bar     [10]float64
	PlayBtn string
	Tag     play.Tag
}

const (
	FFTSamples = 1024
	refresh    = 16

	BtnPlay  = "play"
	BtnPause = "pause"
)

var (
	guiIsDone chan struct{}
	fftc      fft.FFT
	csamples  []complex128
)

func Init() {
	fftc, _ = fft.New(FFTSamples)
	csamples = make([]complex128, FFTSamples)
}

// OnMount sets up player state.
func (p *Player) OnMount() {
	p.PlayBtn = BtnPause

	// Make a channel to control UI.
	guiIsDone = make(chan struct{})

	// Rendering loop.
	go func() {
		for range time.Tick(refresh * time.Millisecond) {
			select {
			default:
				// Render pl0x.
				app.Render(p)
			case <-guiIsDone:
				return
			}
		}
	}()

	// FFT loop.
	go func() {
		for range time.Tick(refresh * time.Millisecond) {
			select {
			default:
				if !playlist.IsPlaying() {
					p.ClearBars()
					continue
				}

				s := playlist.GetSamples()
				samples := *s
				// Convert channel slices to complex128 (mono).
				for i := 0; i < FFTSamples; i++ {
					csamples[i] = complex((samples[i][0] + samples[i][1]), 0)
				}
				// An FFT walks into...
				fftc.Transform(csamples)
				// ...a bar...
				for j := 0; j < len(p.Bar); j++ {
					// Consider only half of the frequencies.
					for i := 0; i < FFTSamples/len(p.Bar)/2; i++ {
						p.Bar[j] = 20 * (math.Log(1 + cmplx.Abs(csamples[i+j])))
					}
				}
				// ...and the whole scene unfolds with tedious inevitability.
				// #complexjoke
			case <-guiIsDone:
				return
			}
		}
	}()
	//done <- struct{}{}
}

// Previous plays the previous song.
func (p *Player) Previous() {
	playlist.Back()
}

// Next plays the next song.
func (p *Player) Next() {
	playlist.Next()
}

// TogglePlayback pause/play.
func (p *Player) TogglePlayback() {
	playlist.TogglePause()
}

// ClearBars sets all bars to their initial state.
func (p *Player) ClearBars() {
	for j := 0; j < len(p.Bar); j++ {
		p.Bar[j] = 0
	}
}

// OnDismount stops the playback.
func (p *Player) OnDismount() {
	guiIsDone <- struct{}{}
	playlist.Done()
}

// Render the player.
func (p *Player) Render() string {
	if playlist.IsPlaying() {
		p.PlayBtn = BtnPause
	} else {
		p.PlayBtn = BtnPlay
	}
	p.Tag = playlist.GetTags()
	return `
<div class="center">
	<div>
		<div class="graph">
			<div style="height: 120px; background-color: rgba(0,0,0,0)" class="bar"></div>
				{{ range $key, $data := .Bar }}
				<div style="height: {{$data}}px;" class="bar"></div>
				{{ end }}
			<div style="height: 120px; background-color: rgba(0,0,0,0)" class="bar"></div>
		</div>
	</div>
	<h1>{{ .Tag.Artist }} </h1>
	<h2>{{ .Tag.Title }} </h2>
	<div>
		<button class="button back" onclick="Previous"></button>
		<button class="button {{.PlayBtn}}" onclick="TogglePlayback"></button>
		<button class="button next" onclick="Next"></button>         
	</div>
</div>
`
}
