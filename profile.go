// +build profile

package main

import "github.com/pkg/profile"

func init() {
	profile.Start(
		profile.ProfilePath("./profiling"),
		profile.MemProfile,
	)
}
