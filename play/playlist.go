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

// Playlist manages the playback of audio files.
// TODO Hook up Playlist to Player object instead of using global state.
type Playlist struct {
	*Player

	playing int
	files   []string
	cont    chan struct{}
	tag     Tag
}

// Tag contains metadata about the song.
type Tag struct {
	Artist string
	Title  string
}

// New creates a new Playlist.
func New(files []string) *Playlist {
	p := &Playlist{
		files: files,
		log:   log.New(os.Stdout, "[playlist]", log.LstdFlags),
	}
	go p.run()
	return p
}

// run is the core play loop that plays each audio file in the list.
func (p *Playlist) run() {
	for {
		file := p.files[p.playing]
		if err := p.setTags(file); err != nil {
			p.log.Printf("setting tag: %v", err)
		}
		if err := p.play(file); err != nil {
			p.log.Printf("playing file %s: %v", file, err)
		}
		p.playing = p.playing + 1
		log.Println(p.playing)
		if p.playing >= len(p.files) {
			p.playing = 0
		}
	}
}

// Done signals that we are finished with the currently playing song.
func (p *Playlist) Done() {
	done <- struct{}{}
}

// Back plays the previous song.
func (p *Playlist) Back() {
	p.playing = p.playing - 2
	if p.playing < -1 {
		p.playing = -1
	}
	p.Done()
}

// Next plays the next song.
func (p *Playlist) Next() {
	if p.playing >= len(p.files)-2 {
		p.playing = len(p.files) - 2
	}
	p.Done()
}

// IsPlaying reports the status of the Playlist.
func (p *Playlist) IsPlaying() bool {
	return IsPlaying()
}

// GetSamples returns the samples.
func (p *Playlist) GetSamples() [][2]float64 {
	return samples
}

// GetTags gets the tag for the song.
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
			// This causes <-done to return when the song finishes.
			// Other <-done is coming from some external source to
			// end early.
			close(done)
		})))
	// waits on done signal which can be used to return early(er) than the
	// the song. If the song completes it signals done.
	<-done
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
