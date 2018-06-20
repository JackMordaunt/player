package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jackmordaunt/player/play"

	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/mac"
)

var (
	playlist play.Playlist
)

func main() {
	log.Println(os.Args)
	runtime.GOMAXPROCS(1)
	if len(os.Args) < 2 {
		return
	}

	var files []string
	for _, path := range os.Args[1:] {
		fi, err := os.Stat(path)
		if err != nil {
			panic(err)
		}
		if fi.IsDir() {
			entries, err := ioutil.ReadDir(path)
			if err != nil {
				panic(err)
			}
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				files = append(files, filepath.Join(path, e.Name()))
			}
		} else {
			files = append(files, path)
		}
	}

	playlist = play.New()
	playlist.Init(files)

	go playlist.Start()

	app.Run(&mac.Driver{
		OnRun: func() {
			newMainWindow()
		},
		OnReopen: func(visible bool) {
			if !visible {
				newMainWindow()
			}
		},
	})
}

func newMainWindow() {
	app.NewWindow(app.WindowConfig{
		Title:  "player",
		Width:  400,
		Height: 400,
		Mac: app.MacWindowConfig{
			BackgroundVibrancy: app.VibeDark,
		},
		DefaultURL: "/Player",
		OnClose: func() bool {
			playlist.Done()
			os.Exit(0)
			return true
		},
	})
}
