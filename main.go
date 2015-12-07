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
	log.SetFormatter(&log.TextFormatter{DisableColors: true})
	
	config := utils.NewAppConfig()
	sysutils := utils.NewSysUtils(config)
	emailutils := utils.NewEmailUtils(config)
	
	db := persistence.NewMySqlHandler(config, sysutils)
	repo := persistence.NewRepo(db)
	
	manager := utils.NewManager(sysutils)
	
	crawler := crawler.NewCrawler(manager, repo, sysutils)
	crawler.Start()
	
	interactor := core.NewItemInteractor(repo, manager, sysutils)
	webhandler := web.NewWebHandler(interactor, repo, config, emailutils, sysutils)
	webhandler.StartServer()
	
	select{}
}
