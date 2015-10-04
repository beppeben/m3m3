package crawler

import (
	//"fmt"
	. "github.com/beppeben/m3m3/utils"
	"io/ioutil"
	"log"
	//"net/http"
	"regexp"
	//"github.com/petar/GoLLRB/llrb"
)

var (
	max_item_source 	int   = 5
	max_item_crawl 	int	  = 5
	min_img_size   	int64 = 35 * 1000
	max_img_miss 	int   = 10

	img_regx   *regexp.Regexp = regexp.MustCompile("http([^<>\"]+?)\\.(jpg|jpeg)(&quot;|\")")
	items_regx *regexp.Regexp = regexp.MustCompile("<item>([\\S\\s]+?)</item>")
	title_regx *regexp.Regexp = regexp.MustCompile("<title>(<!\\[CDATA\\[)?([\\S\\s]+?)(\\]\\]>)?</title>")
)



type Source struct {
	url      	string
}


func (source *Source) update(num chan int) {
	var managed, updated, misses int
	defer func (){num <- updated}()
	//log.Printf("getting %s", source.url)
	resp, err := client.Get(source.url)
	if err != nil {
		log.Printf("URL %s is not reachable", source.url)
		return
	}
	
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	xml := string(body)
	//source.items = make([]*Item, 0)

	xml_items := items_regx.FindAllStringSubmatch(xml, -1)

	var title string
	var url_matcher, title_matcher [][]string
	
	outer:
	for count, xml_item := range xml_items {
		if managed >= max_item_source || count > max_item_crawl {
			break
		}
		title_matcher = title_regx.FindAllStringSubmatch(xml_item[1], -1)
		l := len(title_matcher)
		if l == 0 {
			continue
		}
		title = title_matcher[0][2]
		url_matcher = img_regx.FindAllStringSubmatch(xml_item[1], -1)
		l = len(url_matcher)
		if l == 0 {
			continue
		}
		
		//candidate images
		img_urls := make([]string, 0)
		
		
		
		//skip item if already managed
		for i := 0; i < l; i++ {
			img_urls = append(img_urls, "http" + url_matcher[i][1] + "." + url_matcher[i][2])
			if manager.IsManaged(&Item{Img_url: img_urls[i]}) {
				managed++
				continue outer
			} 			
		}
		
		//retain item, if a sufficently large image is found
		for i := 0; i < l; i++ {
			if GetFileSize(img_urls[i], client) > min_img_size {
				manager.Insert(&Item{Title: title, Img_url: img_urls[i]})
				managed++
				updated++
				break
			} else {
				misses++
				//stop crawling source if images are too small
				if misses >= max_img_miss {
					break outer
				}
			}							
		}
	}
	
}
