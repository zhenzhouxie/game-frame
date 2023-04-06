package router

import (
	"gameframe/manager"
	"gameframe/server"
)

func Router(gm *manager.GameMgr) *server.Router {
	router := server.NewRouter()
	router.Use(1, gm.PlayerLogin)
	return router
}
