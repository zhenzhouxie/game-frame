package main

import (
	"gameframe/manager"
	"gameframe/router"
	"gameframe/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	service := server.NewServer(":8080")
	game := manager.NewGameMgr()
	service.Router(router.Router(game))
	service.Run()
	WaitSignal(service.OnSystemSignal)
}

func WaitSignal(fn func(sig os.Signal) bool) {
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGPIPE)

	for sig := range ch {
		if !fn(sig) {
			close(ch)
			break
		}
	}
}
