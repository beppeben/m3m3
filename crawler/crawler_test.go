package crawler

import (
	. "github.com/beppeben/m3m3/domain"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/beppeben/m3m3/utils"
	"testing"
	"github.com/spf13/viper"
)

type FakeRepo int

func (r FakeRepo) GetItemByUrl(img_url string) (*Item, error) {
	return nil, errors.New("no such item")
}



func TestUpdateFeed(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{DisableColors: true})
	var repo FakeRepo
	v := viper.New()
	v.Set("HTTP_DIR", "/var/www/m3m3/")
	config := utils.NewCustomAppConfig(v)
	sysutils := utils.NewSysUtils(config)
	manager := utils.NewManager(sysutils)
	cr := newCrawlerNoSources(manager, repo, sysutils)
	
	u := "http://feeds.feedburner.com/DamnLOL"
	feed := &Feed{url: u, name: "test feed"}
	c := make(chan int)
	go cr.update(feed, 10, c)
	updated := <-c
	if updated < 4 {
		t.Error("Too few images for this feed")
	}
}

