package crawler

import (
	"log"
	"time"
	utils "github.com/beppeben/m3m3/utils"
)

var(		
	sources			[]Source
	Img_urls			[]string
	
	wait_time		time.Duration = 5
)

func init() {
	log.Println("[CRAW] Crawler started, updating sources...")
	go updateSources()	
}


func getSourcesFromFile() {
	rss_urls, err := utils.ReadLines("./config/rss.conf")
	if err != nil {
		log.Printf("[CRAW] Couldn't read rss list: %s", err)
		return
	}
	sources = make([]Source,0)
	for _,s_url := range rss_urls {
		sources = append(sources, Source{url: s_url})
	}
}

func getShuffledUrls() []string {
	for i,_ := range sources {sources[i].resetIterator()}
	result := make([]string, 0)
	exit := false
	for {
		if exit {break}
		exit = true
		for i,_ := range sources {
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

func Start_Crawler(){}

func updateSources() {
	getSourcesFromFile()
	done := make(chan bool)
	for i,_ := range sources {go sources[i].update(done)}	
	for i := 0; i < len(sources); i++ {<-done}	
	Img_urls = getShuffledUrls()
	log.Printf("[CRAW] Sources updated with %d image urls", len(Img_urls))
	time.Sleep(time.Minute * wait_time)
	updateSources()
}

