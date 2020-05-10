package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/expectedsh/gomon/tests/x"
)

func main() {
	fmt.Println("STARTING")
	x.T()

	end := make(chan os.Signal, 1)
	signal.Notify(end, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-end

	fmt.Println("EXITING")
	time.Sleep(time.Second * 5)
}
