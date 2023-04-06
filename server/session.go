package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type IPacker interface {
	Pack(message *Message) ([]byte, error)
	Unpack(reader io.Reader) (*Message, error)
}

type Session struct {
	server         *Server //会话所在的服务
	Conn           net.Conn
	IsClose        bool
	packer         IPacker
	WriteCh        chan *Message
	IsPlayerOnline bool
	MessageHandler func(packet *SessionPacket)
	//
}

type SessionPacket struct {
	Msg  *Message
	Sess *Session
}

func NewSession(conn net.Conn, svr *Server) *Session {
	return &Session{
		server: svr,
		Conn:   conn,
		packer: &NormalPacker{
			ByteOrder: binary.BigEndian,
		},
		WriteCh: make(chan *Message, 10),
	}
}

func (s *Session) Run() {
	go s.Read()
	go s.Write()

}

func (s *Session) Read() {
	for {
		err := s.Conn.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			fmt.Println(err)
			continue
		}
		message, err := s.packer.Unpack(s.Conn)
		if _, ok := err.(net.Error); ok {
			continue
		}
		fmt.Println("receive message:", string(message.Data))
		s.MessageHandler(&SessionPacket{
			Msg:  message,
			Sess: s,
		})
	}
}

func (s *Session) Write() {
	for resp := range s.WriteCh {
		s.send(resp)
	}
}

func (s *Session) send(message *Message) {
	err := s.Conn.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		fmt.Println(err)
		return
	}
	bytes, err := s.packer.Pack(message)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.Conn.Write(bytes)

}

func (s *Session) SendMsg(msg *Message) {
	s.WriteCh <- msg
}
