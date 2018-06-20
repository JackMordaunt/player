package main

import (
	"log"
	"os"
	"runtime"

	"github.com/grocid/mistlur/play"

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

	playlist = play.New()
	playlist.Init(os.Args[1:])

	go func() {
		playlist.Start()
	}()

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
