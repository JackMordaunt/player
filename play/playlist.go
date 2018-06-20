package play

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	id3 "github.com/mikkyang/id3-go"
	"github.com/pkg/errors"
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
		file := p.files[p.playing]
		if err := p.setTags(file); err != nil {
			// TODO log err.
		}
		if err := play(file); err != nil {
			// TODO log err.
		}
		p.playing = p.playing + 1
		log.Println(p.playing)
		if p.playing >= len(p.files) {
			p.playing = 0
		}
	}
}

func (p *Playlist) Done() {
	done <- struct{}{}
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

func play(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	s, format, err := mp3.Decode(f)
	if err != nil {
		return errors.Wrap(err, "decoding mp3")
	}
	done = make(chan struct{})
	InitPlayer(
		format.SampleRate,
		format.SampleRate.N(time.Second/10),
	)
	Play(beep.Seq(s, beep.Callback(
		func() {
			close(done)
		})))
	<-done
	isPlaying = false
	return nil
}

func (p *Playlist) setTags(file string) error {
	// Read tags.
	mp3File, err := id3.Open(file)
	if err != nil {
		return err
	}
	p.tag = Tag{
		Artist: mp3File.Artist(),
		Title:  mp3File.Title(),
	}
	mp3File.Close()
	return nil
}
