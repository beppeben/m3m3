package main

import (
	"github.com/beppeben/m3m3/persistence"
	"github.com/beppeben/m3m3/crawler"
	"github.com/beppeben/m3m3/utils"
	"github.com/beppeben/m3m3/core"
	"github.com/beppeben/m3m3/web"
	log "github.com/Sirupsen/logrus"
)

func main() {
	log.SetLevel(log.InfoLevel)
	
	db := persistence.NewMySqlHandler()
	repo := persistence.NewRepo(db)
	
	manager := utils.NewManager()
	
	crawler := crawler.NewCrawler(manager, repo)
	crawler.Start()
	
	interactor := core.NewItemInteractor(repo, manager)
	webhandler := web.NewWebHandler(interactor, repo)
	webhandler.StartServer()
	
	select{}
}
