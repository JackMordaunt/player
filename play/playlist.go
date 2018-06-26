package play

import (
	"log"
	"os"

	id3 "github.com/mikkyang/id3-go"
	"github.com/pkg/errors"
)

// Playlist manages the playback of audio files.
type Playlist struct {
	player *Player

	playing int
	files   []string
	cont    chan struct{}
	tag     Tag
	log     *log.Logger
}

// Tag contains metadata about the song.
type Tag struct {
	Artist string
	Title  string
}

// New creates a new Playlist.
func New(files []string) *Playlist {
	p := &Playlist{
		player: NewPlayer(nil),
		files:  files,
		log:    log.New(os.Stdout, "[playlist]", log.LstdFlags),
	}
	go p.run()
	return p
}

// run is the core play loop that plays each audio file in the list.
func (p *Playlist) run() {
	for {
		file := p.files[p.playing]
		if err := p.setTag(file); err != nil {
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
	p.player.Done()
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

// IsPlaying reports the status of the Player.
func (p *Playlist) IsPlaying() bool {
	return p.player.IsPlaying()
}

// TogglePause toggles the playback state.
func (p *Playlist) TogglePause() {
	if p.IsPlaying() {
		p.player.Pause()
	} else {
		p.player.Resume()
	}
}

// GetSamples returns the samples.
func (p *Playlist) GetSamples() [][2]float64 {
	return p.player.GetSamples()
}

// GetTags gets the tag for the song.
func (p *Playlist) GetTags() Tag {
	return p.tag
}

func (p *Playlist) play(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return errors.Wrap(err, "opening audio file")
	}
	return p.player.Play(Audio{
		ReadCloser: f,
		Format:     MP3,
	})
}

func (p *Playlist) setTag(file string) error {
	mp3File, err := id3.Open(file)
	if err != nil {
		return err
	}
	p.tag = Tag{
		Artist: mp3File.Artist(),
		Title:  mp3File.Title(),
	}
	return mp3File.Close()
}
