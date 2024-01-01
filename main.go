package main

import "time"

func main() {
	go func() {
		time.Sleep(time.Second * 4)
		SendFile(1000000000)
	}()
	server := &FilServer{}
	server.Start()
}
