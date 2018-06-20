package main

import "github.com/murlokswarm/app"

func init() {
	app.Import(&MenuBar{})
}

// MenuBar is the component that define the menu bar.
type MenuBar app.ZeroCompo

// Render returns return the HTML describing the menu bar.
func (m *MenuBar) Render() string {
	return `
<menu>
	<menu label="app">
		<menuitem label="Close" selector="performClose:" shortcut="meta+w" />
		<menuitem label="Quit" selector="terminate:" shortcut="meta+q" /> 
	</menu>
	<menu label="Play">
		<menuitem label="Toggle pause/play" onclick="TogglePause" shortcut="meta+p" />
		<menuitem label="Next" onclick="Next" shortcut="meta+n" /> 
		<menuitem label="Back" onclick="Back" shortcut="meta+b" /> 
	</menu>
</menu>
	`
}

// Next advances the playlist to the next song.
func (m *MenuBar) Next() {
	playlist.Next()
}

// Back advances the playlist to the previous song.
func (m *MenuBar) Back() {
	playlist.Back()
}

// TogglePause toggles playback.
func (m *MenuBar) TogglePause() {
	playlist.TogglePause()
}
