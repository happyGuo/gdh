package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"gdhMQ/servers/gdhd"
)

func main()  {

	g := gdhd.New()
	g.Entry()
	var s = make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGKILL, syscall.SIGTERM)
	fmt.Println(os.Getpid())
	<-s
}
