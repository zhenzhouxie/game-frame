package manager

import (
	"gameframe/player"
	"gameframe/server"
)

type GameMgr struct {
	Players map[player.Uin]*player.Player //在线玩家
}

func NewGameMgr() *GameMgr {
	return &GameMgr{
		Players: make(map[player.Uin]*player.Player),
	}
}

// 玩家登录
func (pm *GameMgr) PlayerLogin(mession *server.SessionPacket) error {
	return nil
}
