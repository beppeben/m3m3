package crawler

import (
	"testing"
	"fmt"
	//"strconv"
	"net/url"
)

func TestUtils(t *testing.T) {
	/*
	manager := NewManager()
	it1 := &Item{Img_url: "www.test.com", Ncomments: 10, timestamp: 12}
	it2 := &Item{Img_url: "www.test3.com", Title: "lala", Ncomments: 4, timestamp: 120}
	it3 := &Item{Img_url: "www.tezfst.com", Title: "frlala", Ncomments: 1, timestamp: 2}
	manager.Insert(it1)
	manager.Insert(it2)
	manager.Insert(it3)
	fmt.Println("item managed: " + strconv.FormatBool(manager.IsManaged(it1)))
	manager.RefreshJson()
	fmt.Println(manager.GetJson())
	*/
	img_url := "http://bilder.bild.de/fotos/miley-cyrus-saengerin-plant-nackt-konzert-48047778-43015058/Bild/2.bild.jpg"
	escaped := url.QueryEscape(img_url)
	fmt.Println(escaped)
}