package crawler

import (
	"github.com/beppeben/m3m3/utils"
	"log"
	"net/http"
	"time"
)

var (
	sources   	[]Source
	//Img_urls  []string
	manager		*IManager
	client    	*http.Client
	tr        	*http.Transport
	wait_time 	time.Duration = 1
)

func init() {
	log.Println("[CRAW] Crawler started, updating sources...")
	tr = &http.Transport{}
	client = &http.Client{Transport: tr}
	manager = NewManager()
	go updateSources()
}

func getSourcesFromFile() {
	rss_urls, err := utils.ReadLines("./config/rss.conf")
	if err != nil {
		log.Printf("[CRAW] Couldn't read rss list: %s", err)
		return
	}
	sources = make([]Source, 0)
	for _, s_url := range rss_urls {
		sources = append(sources, Source{url: s_url})
	}
}

/*
func getShuffledUrls() []string {
	for i, _ := range sources {
		sources[i].resetIterator()
	}
	result := make([]string, 0)
	exit := false
	for {
		if exit {
			break
		}
		exit = true
		for i, _ := range sources {
			img, err := sources[i].nextImage()
			if err == nil {
				//Img_urls = append(Img_urls, img)
				result = append(result, img)
				exit = false
			}
		}
	}
	return result
}
*/

func Start_Crawler() {}

func GetItems() string {
	return manager.GetJson()
}

func GetZippedItems() []byte {
	return manager.GetZippedJson()
}

func updateSources() {
	getSourcesFromFile()
	c := make(chan int)
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
