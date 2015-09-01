package crawler

import (
	"net/http"
	"io/ioutil"
	"log"
	"regexp"
	"fmt"
	utils "github.com/beppeben/m3m3/utils"
)

var(
	max_img_source	int	= 10
	min_img_size		int64 = 35*1000
	
	img_regx		*regexp.Regexp = regexp.MustCompile("http([^<>\"]+?)\\.(jpg|png|jpeg)(&quot;|\")")
	items_regx 	*regexp.Regexp = regexp.MustCompile("<item>([\\S\\s]+?)</item>")	
)

type Source struct{
	url 			string
	img_urls		[]string
	index		int
}

func (source *Source) resetIterator() {source.index = 0}

func (source *Source) nextImage() (string, error) {	
	if source.index >= len(source.img_urls){
		return "", fmt.Errorf("No next image in source %s", source.url)
	}
	img := source.img_urls[source.index]
	source.index++
	return img, nil
}


func (source *Source) update(done chan bool) {
	resp, err := http.Get(source.url)
	if err != nil {
		log.Printf("URL %s is not reachable", source.url)
		return
	}
	//log.Printf("[CRAW] Updating source %s", source.url)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	xml := string(body)
	source.img_urls = make([]string, 0)
		
	items := items_regx.FindAllStringSubmatch(xml,-1)
		
	var imgurl string
	var url_matcher [][]string

	for _, item := range items {
		if len(source.img_urls) >= max_img_source {break}
		url_matcher = img_regx.FindAllStringSubmatch(item[1],-1)
		l := len(url_matcher)
		if l == 0 {continue}
		//one image per item, if a sufficently large one is found
		for i := 0; i < l; i++{
			imgurl = "http" + url_matcher[i][1] + "." + url_matcher[i][2]
			if (utils.GetFileSize(imgurl) > min_img_size){
				source.img_urls = append(source.img_urls, imgurl)
				break
			}	
		}			
	}	
	//log.Printf("[CRAW] Got %d images", len(source.img_urls))	
	done <- true
}