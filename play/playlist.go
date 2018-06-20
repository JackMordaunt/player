package play

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	id3 "github.com/mikkyang/id3-go"
)

type Playlist struct {
	playing int
	files   []string
	cont    chan struct{}
	tag     Tag
}

type Tag struct {
	Artist string
	Title  string
}

func New(files []string) Playlist {
	return Playlist{
		files: files,
	}
}

func (p *Playlist) Start() {
	for {
		// Set current file.
		file := p.files[p.playing]
		// Read tags
		p.setTags(file)
		play(file)

		p.playing = p.playing + 1

		log.Println(p.playing)

		if p.playing >= len(p.files) {
			Stop()
			mu.Lock()
			isPlaying = false
			p.playing = len(p.files) - 1

			p.cont = make(chan struct{})
			<-p.cont
		}
	}
}

func (p *Playlist) Done() {
	done <- struct{}{}
}

func (p *Playlist) TogglePause() {
	togglePause()
}

func (p *Playlist) Back() {
	p.playing = p.playing - 2
	if p.playing < -1 {
		p.playing = -1
	}
	p.Done()
}

func (p *Playlist) Next() {
	if p.playing >= len(p.files)-2 {
		p.playing = len(p.files) - 2
	}
	p.Done()
}

func (p *Playlist) IsPlaying() bool {
	return IsPlaying()
}

func (p *Playlist) GetSamples() *[][2]float64 {
	return &samples
}

func (p *Playlist) GetTags() Tag {
	return p.tag
}

func play(file string) {
	// Read music file.
	f, err := os.Open(file)

	// Skip if error...
	if err != nil {
		return
	}

	// Decode the data.
	s, format, err := mp3.Decode(f)

	if err != nil {
		return
	}

	// Make a channel to communicate when done.
	done = make(chan struct{})

	// Start playing...
	InitPlayer(
		format.SampleRate,
		format.SampleRate.N(time.Second/10),
	)
	Play(beep.Seq(s, beep.Callback(
		func() {
			close(done)
		})))

	// Wait for done signal, so that the player
	// has finished crunching the file.
	<-done
	isPlaying = false
}

func (p *Playlist) setTags(file string) {
	// Read tags.
	mp3File, _ := id3.Open(file)
	defer mp3File.Close()

	p.tag = Tag{
		Artist: mp3File.Artist(),
		Title:  mp3File.Title(),
	}
}
