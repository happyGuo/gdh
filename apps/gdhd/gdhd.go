package main

import (
	"os"
	"os/signal"
	"syscall"
	"gdhMQ/servers/gdhd"
)

type program struct {
	gdhd *gdhd.GDHD
}

func main()  {


	var s = make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGKILL)
	for {
		cmd := <-s
		if cmd == syscall.SIGKILL {
			break
		}
	}
}
