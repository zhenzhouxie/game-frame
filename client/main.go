package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	c := NewClient()
	c.InputHandlerRegister()
	c.MessageHandlerRegister()
	c.Run()
	WaitSignal(OnSystemSignal)
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

func OnSystemSignal(signal os.Signal) bool {
	fmt.Println("[MgrMgr] 收到信号 %v \n", signal)
	tag := true
	switch signal {
	case syscall.SIGHUP:
		//todo
	case syscall.SIGPIPE:
	default:
		fmt.Println("[MgrMgr] 收到信号准备退出...")
		tag = false

	}
	return tag
}
