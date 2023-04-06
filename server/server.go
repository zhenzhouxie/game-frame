package server

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

type MessageId uint64

type Message struct {
	Id   MessageId
	Data []byte
}

type Server struct {
	tcpListener net.Listener                                       //tcp监听
	address     string                                             //服务配置
	handlers    map[MessageId][]func(message *SessionPacket) error //路由注册
	sessionMgr  []*Session                                         //会话管理
}

func NewServer(address string) *Server {
	sev := &Server{
		address:  address,
		handlers: make(map[MessageId][]func(message *SessionPacket) error),
	}
	return sev
}

func (s *Server) Use(messageId MessageId, handlers ...func(message *SessionPacket) error) {
	if _, ok := s.handlers[messageId]; !ok {
		s.handlers[messageId] = handlers
	} else {
		fmt.Print("Server messageId already register")
	}
}

func (s *Server) Router(router *Router) {
	for id, handlers := range router.handlers {
		if _, ok := s.handlers[id]; !ok {
			s.handlers[id] = handlers
		}
	}
}

func (s *Server) Run() {
	resolveTCPAddr, err := net.ResolveTCPAddr("tcp6", s.address)
	if err != nil {
		panic(err)
	}
	tcpListener, err := net.ListenTCP("tcp6", resolveTCPAddr)
	if err != nil {
		panic(err)
	}
	s.tcpListener = tcpListener
	go s.tcpListening()
}

func (s *Server) tcpListening() {
	for {
		conn, err := s.tcpListener.Accept()
		if _, ok := err.(net.Error); ok {
			fmt.Println(err)
			continue
		}

		newSession := NewSession(conn, s)
		newSession.MessageHandler = s.OnSessionPacket
		newSession.Run()
		s.sessionMgr = append(s.sessionMgr, newSession)
	}
}

func (s *Server) OnSessionPacket(packet *SessionPacket) {
	if handler, ok := s.handlers[packet.Msg.Id]; ok {
		for _, f := range handler {
			f(packet)
		}
		return
	} else {
		fmt.Println("messageId", packet.Msg.Id, "unregistered")
	}
}

func (s *Server) OnSystemSignal(signal os.Signal) bool {
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
