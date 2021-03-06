package crawler

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/beppeben/m3m3/domain"
	"github.com/beppeben/m3m3/utils"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	max_item_crawl    int           = 60
	min_img_size      int64         = 20 * 1000
	max_img_size      int64         = 400 * 1000
	max_img_miss      int           = 60
	frequency_minutes time.Duration = 1
	items_per_period  int           = 1

	img_regx   *regexp.Regexp = regexp.MustCompile("http([^<>\"]+?)\\.(jpg|jpeg)(&quot;|\")")
	items_regx *regexp.Regexp = regexp.MustCompile("<item>([\\S\\s]+?)</item>")
	title_regx *regexp.Regexp = regexp.MustCompile("<title>(<!\\[CDATA\\[)?([\\S\\s]+?)(\\]\\]>)?</title>")
	link_regx  *regexp.Regexp = regexp.MustCompile("<link>(<!\\[CDATA\\[)?([\\S\\s]+?)(\\]\\]>)?</link>")
)

type SysUtils interface {
	SaveTempImage(url string, client *http.Client) (string, error)
	DeleteTempImages() error
	GetRSSDefs() ([]string, error)
}

type Repository interface {
	GetItemByUrl(img_url string) (*Item, error)
}

type Crawler struct {
	feeds     []*Feed
	manager   *utils.IManager
	repo      Repository
	client    *http.Client
	tr        *http.Transport
	wait_time time.Duration
	sutils    SysUtils
}

func NewCrawler(manager *utils.IManager, repo Repository, sutils SysUtils) *Crawler {
	crawler := newCrawlerNoSources(manager, repo, sutils)
	lines, err := crawler.sutils.GetRSSDefs()
	if err != nil {
		panic(err.Error())
	}
	err = crawler.getSourcesFromFile(lines)
	if err != nil {
		panic(err.Error())
	}
	return crawler
}

func newCrawlerNoSources(manager *utils.IManager, repo Repository, sutils SysUtils) *Crawler {
	crawler := newCrawler(manager, repo, sutils)
	err := crawler.sutils.DeleteTempImages()
	if err != nil {
		panic(err.Error())
	}
	return crawler
}

func newCrawler(manager *utils.IManager, repo Repository, sutils SysUtils) *Crawler {
	crawler := &Crawler{}
	timeout := time.Duration(10 * time.Second)
	crawler.tr = &http.Transport{}
	crawler.client = &http.Client{Transport: crawler.tr, Timeout: timeout}
	crawler.manager = manager
	crawler.repo = repo
	crawler.sutils = sutils
	return crawler
}

func (cr *Crawler) getSourcesFromFile(lines []string) error {
	feeds := make([]*Feed, 0)
	for _, line := range lines {
		parts := strings.Split(line, " --- ")
		if len(parts) != 2 {
			continue
		}
		feeds = append(feeds, &Feed{url: parts[0], name: parts[1]})
	}
	cr.feeds = feeds
	return nil
}

func (cr *Crawler) Start() {
	go cr.getFeeds(0)
	log.Infoln("Crawler started")
}

func (cr *Crawler) getFeeds(arrears int) {
	threshold := int(float32(cr.manager.MaxShowItems())/float32(len(cr.feeds)) + 1)
	to_update := utils.PositivePart(cr.manager.MaxShowItems()-cr.manager.Count()) +
		items_per_period + arrears
	goal := to_update
	sort.Sort(ByNbManaged(cr.feeds))
	flag := to_update <= len(cr.feeds)

	for _, f := range cr.feeds {
		log.Debugf("feed %s: %d items", f.name, f.nb_managed)
	}

	log.WithFields(log.Fields{
		"threshold":     threshold,
		"to_update":     to_update,
		"max_manager":   cr.manager.MaxShowItems(),
		"count_manager": cr.manager.Count(),
		"n_feeds":       len(cr.feeds),
	}).Debugln("Updating feeds")

	//update all sources in parallel
	c := make(chan int)
	var i, total int
	for i, _ = range cr.feeds {
		cand := 1
		if !flag {
			cand = utils.PositivePart(threshold - cr.feeds[i].nb_managed)
			if cand == 0 {
				cand = 1
			}
			if cand > to_update {
				cand = to_update
			}
		}
		to_update -= cand

		go cr.update(cr.feeds[i], cand, c)
		if to_update <= 0 {
			break
		}
	}
	//wait for all the routines to return
	for k := 0; k <= i; k++ {
		total += <-c
	}
	//this is to avoid annoying logs about unsolicited requests to idle conns
	cr.tr.CloseIdleConnections()
	cr.manager.RefreshJson()
	log.Debugf("Crawler got %d items", total)
	time.Sleep(time.Minute * frequency_minutes)
	//if you crawl less than desired, you'll crawl more in the next step
	arrears = goal - total
	if arrears > 7 {
		arrears = 7
	}
	//if cr.manager.Count() < cr.manager.MaxShowItems() {
	//	arrears = 0
	//}
	cr.getFeeds(arrears)
}

func (cr *Crawler) update(f *Feed, to_update int, num chan int) {
	var updated, up_misses, lo_misses int
	defer func() {
		log.WithFields(log.Fields{
			"feed":      f.name,
			"updated":   updated,
			"to_update": to_update,
			"up_misses": up_misses,
			"lo_misses": lo_misses,
		}).Debug()
		num <- updated
	}()
	resp, err := cr.client.Get(f.url)
	if err != nil {
		log.Debugf("%v", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	xml := string(body)
	log.Debugf("%s", xml)
	xml_items := items_regx.FindAllStringSubmatch(xml, -1)

	var title, link string
	var url_matcher, title_matcher, link_matcher [][]string

outer:
	for count, xml_item := range xml_items {
		if count > max_item_crawl || updated >= to_update {
			break
		}
		unesc := html.UnescapeString(xml_item[1])
		title_matcher = title_regx.FindAllStringSubmatch(unesc, -1)
		l := len(title_matcher)
		if l == 0 {
			log.Debug("Unmatched title")
			continue
		}
		title = title_matcher[0][2]

		link_matcher = link_regx.FindAllStringSubmatch(unesc, -1)
		l = len(link_matcher)
		if l == 0 {
			log.Debug("Unmatched link")
			continue
		}
		link = link_matcher[0][2]

		url_matcher = img_regx.FindAllStringSubmatch(unesc, -1)
		l = len(url_matcher)
		if l == 0 {
			log.Debugf("Unmatched image url in %s", unesc)
			continue
		}

		//candidate images
		img_urls := make([]string, 0)

		//skip item if already managed
		for i := 0; i < l; i++ {
			img_urls = append(img_urls, "http"+url_matcher[i][1]+"."+url_matcher[i][2])
			if cr.manager.IsManaged(&Item{Tid: utils.Hash(img_urls[i]), Title: title}) {
				continue outer
			} else if _, err = cr.repo.GetItemByUrl(img_urls[i]); err == nil {
				continue outer
			}

		}

		//retain item, if a sufficently "good" image is found
		for i := 0; i < l; i++ {
			size := utils.GetFileSize(img_urls[i], cr.client)
			if size > min_img_size && size < max_img_size {
				hash, err := cr.sutils.SaveTempImage(img_urls[i], cr.client)
				if err != nil {
					log.Debugf("Did not save temp image %s: %v", img_urls[i], err)
					continue
				}
				if cr.manager.Insert(&Item{Title: title, Tid: hash,
					Url: img_urls[i], Source: f.name, Src: f, Link: link}) {
					log.Debugf("Inserted image %s", img_urls[i])
					updated++
				}
				break
			} else {
				if size < min_img_size {
					log.Debugf("Image too small: %s", img_urls[i])
					lo_misses++
				} else {
					log.Debugf("Image too big: %s", img_urls[i])
					up_misses++
				}
				//stop crawling source if images are too small/big
				if up_misses+lo_misses >= max_img_miss {
					break outer
				}
			}
		}
	}
}
