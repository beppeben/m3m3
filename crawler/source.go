package crawler

import (
	. "github.com/beppeben/m3m3/utils"
	"github.com/beppeben/m3m3/db"
	"io/ioutil"
	//"log"
	"regexp"
)

var (
	//max_item_source 	int   = 5
	max_item_crawl 	int	  = 5
	min_img_size   	int64 = 35 * 1000
	max_img_size   	int64 = 200 * 1000
	max_img_miss 	int   = 10

	img_regx   *regexp.Regexp = regexp.MustCompile("http([^<>\"]+?)\\.(jpg|jpeg)(&quot;|\")")
	items_regx *regexp.Regexp = regexp.MustCompile("<item>([\\S\\s]+?)</item>")
	title_regx *regexp.Regexp = regexp.MustCompile("<title>(<!\\[CDATA\\[)?([\\S\\s]+?)(\\]\\]>)?</title>")
)


type Source struct {
	url      		string
	name				string
	max_updates		int
}


func (source *Source) update(num chan int) {
	var updated, misses int
	defer func (){num <- updated}()
	//log.Printf("getting %s", source.url)
	resp, err := client.Get(source.url)
	if err != nil {
		return
	}
	
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	xml := string(body)

	xml_items := items_regx.FindAllStringSubmatch(xml, -1)

	var title string
	var url_matcher, title_matcher [][]string
	
	outer:
	for count, xml_item := range xml_items {
		if count > max_item_crawl {
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
			if manager.IsManaged(&Item{Tid: ComputeMd5(img_urls[i]), Title: title}) {
				continue outer
			} else if _, err = db.FindItemByUrl(img_urls[i]); err == nil {
				continue outer
			}
	
		}
		
		//retain item, if a sufficently "good" image is found
		for i := 0; i < l; i++ {
			size := GetFileSize(img_urls[i], client)
			if size > min_img_size && size < max_img_size {
				hash, err := SaveTempImage (img_urls[i], client)
				if err != nil {
					continue
				}
				manager.Insert(&Item{Title: title, Tid: hash, 
					Url: img_urls[i], Source: source.name})
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
