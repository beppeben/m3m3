package crawler

import (
	. "github.com/beppeben/m3m3/utils"
	"log"
	"net/http"
	"time"
	"strings"
)

var (
	sources   	[]Source
	manager		*IManager
	client    	*http.Client
	tr        	*http.Transport
	wait_time 	time.Duration = 1
	c 			chan int
)

func Start_Crawler() {}

func init() {
	log.Println("[CRAW] Crawler started, updating sources...")
	timeout := time.Duration(5 * time.Second)
	tr = &http.Transport{}
	client = &http.Client{Transport: tr, Timeout: timeout}
	manager = NewManager()
	c = make(chan int)
	go updateSources()
}

func GetItems() string {
	return manager.GetJson()
}

func GetZippedItems() []byte {
	return manager.GetZippedJson()
}

func GetItemByTid (tid string) (*Item, bool) {
	return manager.GetItemByTid(tid)
}

func GetItemById (id int64) (*Item, bool) {
	return manager.GetItemById(id)
}

func NotifyItemId (url string, id int64) {
	manager.NotifyItemId (url, id)
}

func NotifyComment (comment *Comment) {
	manager.NotifyComment (comment)
}

func getSourcesFromFile() {
	lines, err := ReadLines("./config/rss.conf")
	if err != nil {
		log.Printf("[CRAW] Couldn't read rss list: %s", err)
		return
	}
	sources = make([]Source, 0)
	for _, line := range lines {
		parts := strings.Split(line, " --- ")
		if len(parts) != 2 {
			continue
		}
		sources = append(sources, Source{url: parts[0], name: parts[1]})
	}
}

func updateSources() {
	getSourcesFromFile()
	
	//update all sources in parallel
	for i, _ := range sources {
		go sources[i].update(c)
	}
	var total int
	//wait for all the routines to return
	for i := 0; i < len(sources); i++ {
		total += <-c		
	}
	//this is to avoid annoying logs about unsolicited requests to idle conns
	tr.CloseIdleConnections()
	manager.RefreshJson()
	//log.Printf("[CRAW] Sources updated with %d items", total)
	time.Sleep(time.Minute * wait_time)
	updateSources()
}
