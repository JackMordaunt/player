package main

import (
	"math"
	"math/cmplx"
	"time"

	"github.com/ktye/fft"
	"github.com/murlokswarm/app"

	"github.com/grocid/mistlur/play"
)

func init() {
	app.Import(&Player{})
}

// Player is the component displaying Player.
type Player struct {
	Bar [10]float64
	Tag play.Tag

	done     chan struct{}
	fftc     fft.FFT
	csamples []complex128
}

const (
	fftSamples = 1024
	refresh    = 16
)

// OnMount sets up player state.
func (p *Player) OnMount() {
	p.fftc, _ = fft.New(fftSamples)
	p.csamples = make([]complex128, fftSamples)
	p.done = make(chan struct{})
	// Trigger rendering, since modifying the bar values doesn't automatically
	// trigger a re-render.
	go func() {
		for range time.Tick(refresh * time.Millisecond) {
			select {
			default:
				app.Render(p)
			case <-p.done:
				return
			}
		}
	}()
	go func() {
		for range time.Tick(refresh * time.Millisecond) {
			select {
			default:
				p.computeBars()
			case <-p.done:
				return
			}
		}
	}()
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
	for ii := 0; ii < len(p.Bar); ii++ {
		p.Bar[ii] = 1
	}
}

// OnDismount stops the playback.
func (p *Player) OnDismount() {
	p.done <- struct{}{}
	p.done <- struct{}{}
	playlist.Done()
}

// IsPlaying reports whether the player is actively playing a song.
func (p *Player) IsPlaying() bool {
	return playlist.IsPlaying()
}

func (p *Player) computeBars() {
	if !playlist.IsPlaying() {
		p.ClearBars()
		return
	}
	s := playlist.GetSamples()
	samples := *s
	// Convert channel slices to complex128 (mono).
	for i := 0; i < fftSamples; i++ {
		p.csamples[i] = complex((samples[i][0] + samples[i][1]), 0)
	}
	p.fftc.Transform(p.csamples)
	for j := 0; j < len(p.Bar); j++ {
		// Consider only half of the frequencies.
		for i := 0; i < fftSamples/len(p.Bar)/2; i++ {
			p.Bar[j] = 20 * (math.Log(1 + cmplx.Abs(p.csamples[i+j])))
		}
	}
}

// Render the player.
func (p *Player) Render() string {
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
		<button class="button {{if .IsPlaying}}pause{{else}}play{{end}}" onclick="TogglePlayback"></button>
		<button class="button next" onclick="Next"></button>         
	</div>
</div>
`
}
