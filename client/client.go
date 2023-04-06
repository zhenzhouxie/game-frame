package main

import (
	"fmt"
	"gameframe/protocol/gen/messageId"
	"gameframe/server"
	"os"
	"syscall"
)

type Client struct {
	cli             *server.Client
	inputHandlers   map[string]InputHandler
	messageHandlers map[messageId.MessageId]MessageHandler
	console         *ClientConsole
	chInput         chan *InputParam
}

func NewClient() *Client {
	c := &Client{
		cli:             server.NewClient(":8080"),
		inputHandlers:   map[string]InputHandler{},
		messageHandlers: map[messageId.MessageId]MessageHandler{},
		console:         NewClientConsole(),
	}
	c.cli.OnMessage = c.OnMessage
	c.cli.ChMsg = make(chan *server.Message, 1)
	c.chInput = make(chan *InputParam, 1)
	c.console.chInput = c.chInput
	return c
}

func (c *Client) Run() {
	go func() {
		for input := range c.chInput {
			inputHandler := c.inputHandlers[input.Command]
			if inputHandler != nil {
				inputHandler(input)
			}
		}
	}()
	go c.console.Run()
	go c.cli.Run()
}

func (c *Client) OnMessage(packet *server.ClientPacket) {
	if handler, ok := c.messageHandlers[messageId.MessageId(packet.Msg.Id)]; ok {
		handler(packet)
	}
}

func (c *Client) OnSystemSignal(signal os.Signal) bool {
	fmt.Println("[Client] 收到信号 %v \n", signal)
	tag := true
	switch signal {
	case syscall.SIGHUP:
		//todo
	case syscall.SIGPIPE:
	default:
		fmt.Println("[Client] 收到信号准备退出...")
		tag = false

	}
	return tag
}
