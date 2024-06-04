package main

import (
	"github.com/pkg/profile"
)

func main() {

	// profiling our program:
	defer profile.Start(profile.MemProfileRate(1), profile.MemProfile, profile.ProfilePath(".")).Stop()
	// wait 3 seconds then start sending the file
	go func() {
		// time.Sleep(time.Second * 3)
		SendFile(1024)
	}()
	server := &FilServer{}
	server.Start()
}
