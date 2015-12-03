package crawler

import (
	. "github.com/beppeben/m3m3/domain"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/beppeben/m3m3/utils"
	"testing"
)

type FakeRepo int

func (r FakeRepo) GetItemByUrl(img_url string) (*Item, error) {
	return nil, errors.New("no such item")
}

func TestUpdateFeed(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{DisableColors: true})
	var repo FakeRepo
	manager := utils.NewManager()
	cr := NewCrawler(manager, repo)
	
	u := "http://www.lifehack.org/feed"
	feed := &Feed{url: u, name: "test feed"}
	c := make(chan int)
	cr.update(feed, 10, c)
}