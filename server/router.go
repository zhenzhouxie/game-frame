package server

import "fmt"

type Router struct {
	handlers map[MessageId][]func(message *SessionPacket) error //路由注册
}

func NewRouter() *Router {
	return &Router{
		handlers: make(map[MessageId][]func(message *SessionPacket) error),
	}
}

func (r *Router) Use(messageId MessageId, handlers ...func(message *SessionPacket) error) {
	if _, ok := r.handlers[messageId]; !ok {
		r.handlers[messageId] = handlers
	} else {
		fmt.Print("Server messageId already register")
	}
}
